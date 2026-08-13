/*
** Copyright (c) 2026 Oracle and/or its affiliates.
**
** The Universal Permissive License (UPL), Version 1.0
 */

package ttc

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oracle/go-driver/driver/common"
)

const (
	// authToken is the TTC key used to send the bearer token itself.
	authToken = "AUTH_TOKEN"
	// authHeader is the TTC key used for the OCI signed header payload.
	authHeader = "AUTH_HEADER"
	// authSignature is the TTC key used for the OCI signature of authHeader.
	authSignature = "AUTH_SIGNATURE"
	// tokenFileName is the default token file name used in token directories.
	tokenFileName = "token"
	// ociPrivateKeyFileName is the OCI database private key file name used for
	// OCI IAM token authentication.
	ociPrivateKeyFileName = "oci_db_key.pem"
)

// tokenProviderContext carries provider-specific inputs derived from the
// connection configuration and live session state.
type tokenProviderContext struct {
	connectString  string
	sessionContext *common.SessionContext
	tokenLocation  string
}

// tokenProvider encapsulates the behavior that differs between token
// authentication providers such as OCI IAM and generic OAuth.
type tokenProvider interface {
	logonMode() int64
	resolveTokenPath(tokenLocation string) (string, error)
	applyAuthData(oauthPacket *oAuth, token string, ctx tokenProviderContext) error
}

// TokenAuthenticator performs the common TTC OAUTH login roundtrip for
// token-based authentication and delegates provider-specific behavior to a
// tokenProvider implementation.
type TokenAuthenticator struct {
	tokenAuthentication common.TokenAuthenticationType
	accessToken         string
	tokenLocation       string
	connectString       string
	provider            tokenProvider
	shelf               ttiShelf[common.MessageType]
	sessionContext      *common.SessionContext
}

type ociTokenProvider struct{}

type oauthTokenProvider struct{}

// NewTokenAuthenticator creates a token authenticator for the configured token
// authentication mode. If accessToken is non-empty it is used directly;
// otherwise the authenticator reads the token from tokenLocation.
func NewTokenAuthenticator(tokenAuthentication common.TokenAuthenticationType, accessToken, tokenLocation, connectString string) *TokenAuthenticator {
	normalizedAuth := common.TokenAuthenticationType(strings.ToUpper(strings.TrimSpace(tokenAuthentication.String())))
	return &TokenAuthenticator{
		tokenAuthentication: normalizedAuth,
		accessToken:         strings.TrimSpace(accessToken),
		tokenLocation:       strings.TrimSpace(tokenLocation),
		connectString:       connectString,
		provider:            newTokenProvider(normalizedAuth),
	}
}

// newTokenProvider returns the provider implementation for the requested token
// authentication type, or nil when the type is not supported.
func newTokenProvider(tokenAuthentication common.TokenAuthenticationType) tokenProvider {
	switch common.TokenAuthenticationType(strings.ToUpper(strings.TrimSpace(tokenAuthentication.String()))) {
	case common.TokenAuthenticationOCI:
		return ociTokenProvider{}
	case common.TokenAuthenticationOAuth:
		return oauthTokenProvider{}
	default:
		return nil
	}
}

// SetShelf assigns the TTC shelf used to create and stream authentication
// messages.
func (ta *TokenAuthenticator) SetShelf(shelf *ttiShelf[common.MessageType]) {
	ta.shelf = *shelf
}

// SetSessionContext assigns the session context used to persist properties
// returned by the server and to expose live network metadata to providers.
func (ta *TokenAuthenticator) SetSessionContext(sessCtx *common.SessionContext) {
	ta.sessionContext = sessCtx
}

// Authenticate validates the configured token and performs the TTC OAUTH
// authentication exchange with the server.
func (ta *TokenAuthenticator) Authenticate(ctx context.Context) error {
	common.Odl.Debug("Start TOKEN authentication")
	if ta.provider == nil {
		return common.NewOracleError(common.NoAuthenticatorError, nil, ta.tokenAuthentication)
	}

	token, err := ta.resolveAccessToken()
	if err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}
	if err := validateJWTExpiration(token); err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}

	return ta.doOAuth(ctx, token)
}

// doOAuth builds and sends the TTC OAUTH message, then processes the server's
// authentication reply and updates the session context with returned
// connection properties.
func (ta *TokenAuthenticator) doOAuth(ctx context.Context, token string) error {
	shelf := ta.shelf
	streamer := shelf.GetMessageStreamer().(MessageStreamerInterface)

	oauthMsg, err := shelf.GetMessageFactory().(Factory).GetMessageForFunction(TTIFUN, oauth)
	if err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}

	oauthPacket := oauthMsg.(*oAuth)
	oauthPacket.setConnectString(ta.connectString)
	oauthPacket.setLogonMode(ta.provider.logonMode())
	if err := oauthPacket.prepareForTokenOAUTH(common.B1Array{}); err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}
	if err := ta.provider.applyAuthData(oauthPacket, token, tokenProviderContext{
		connectString:  ta.connectString,
		sessionContext: ta.sessionContext,
		tokenLocation:  ta.tokenLocation,
	}); err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}

	if err := streamer.Push(ctx, oauthMsg); err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}
	if err := streamer.Flush(ctx); err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}

	oauthRPACallBack := func(t *messageHeader) (common.Message[common.MessageType], error) {
		return shelf.GetMessageFactory().(Factory).GetMessageForFunction(TTIRPA, oauth)
	}
	streamer.RegisterPreUnmarshallCallback(TTIRPA, oauthRPACallBack)
	defer streamer.UnRegisterPreUnmarshallCallback(TTIRPA)

	oerCallback := func(t *messageHeader) (common.Message[common.MessageType], error) {
		return shelf.GetMessageFactory().(Factory).GetMessage(TTIOER)
	}
	streamer.RegisterPreUnmarshallCallback(TTIOER, oerCallback)
	defer streamer.UnRegisterPreUnmarshallCallback(TTIOER)

	var oauthrpa *OAuthRPA
	for {
		msg, err := streamer.Pull(ctx, TTIRPA, TTIOER, TTIWRN)
		if err != nil {
			return common.NewOracleError(common.AuthenticatorError, err, nil)
		}
		switch msg.GetMsgCode() {
		case TTIRPA:
			oauthrpa = msg.(*OAuthRPA)
			common.Odl.Debug("Authenticator:", "oAuth-RPA", oauthrpa)
		case TTIOER:
			ttioer := msg.(tTIOerIface)
			if err := ttioer.getError(); err != nil {
				return err
			}
			if oauthrpa == nil {
				return common.NewOracleError(common.InternalError, nil, nil)
			}
			ta.sessionContext.UpdateSessionProperties(oauthrpa.connectionValues)
			return nil
		case TTIWRN:
			logAuthenticationWarning(msg.(*tTIwrn))
		default:
			return common.NewOracleError(common.InternalError, nil, nil)
		}
	}
}

// resolveAccessToken returns the configured access token directly when present,
// otherwise it reads the token from the provider-resolved token location.
func (ta *TokenAuthenticator) resolveAccessToken() (string, error) {
	if ta.accessToken != "" {
		return ta.accessToken, nil
	}

	tokenPath, err := ta.provider.resolveTokenPath(ta.tokenLocation)
	if err != nil {
		return "", err
	}

	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", errors.New("empty token file")
	}
	return token, nil
}

// logonMode returns the OCI IAM token-authentication logon mode, combining the
// standard logon bit with the token-authentication bit.
func (provider ociTokenProvider) logonMode() int64 {
	return common.KpzLogon.Value() | common.KpzLogonToken.Value()
}

// logonMode returns the generic OAuth token-authentication logon mode, which
// uses the standard logon bit without the OCI token bit.
func (oauthTokenProvider) logonMode() int64 {
	return common.KpzLogon.Value()
}

// resolveTokenPath returns the OCI IAM token file path, defaulting to the
// standard OCI db-token directory when tokenLocation is empty.
func (provider ociTokenProvider) resolveTokenPath(tokenLocation string) (string, error) {
	tokenDir, err := resolveTokenDirectory(tokenLocation, filepath.Join(".oci", "db-token"))
	if err != nil {
		return "", err
	}
	return filepath.Join(tokenDir, tokenFileName), nil
}

// resolveTokenPath returns the generic OAuth token file path from either a file
// path or a token directory.
func (oauthTokenProvider) resolveTokenPath(tokenLocation string) (string, error) {
	return resolveTokenLocation(tokenLocation)
}

// applyAuthData adds OCI-specific token authentication fields, including the
// signed header and signature derived from the OCI private key.
func (provider ociTokenProvider) applyAuthData(oauthPacket *oAuth, token string, ctx tokenProviderContext) error {
	header, err := provider.generateTokenHeader(ctx)
	if err != nil {
		return err
	}
	keyPath, err := resolveOCIPrivateKeyPath(ctx.tokenLocation)
	if err != nil {
		return err
	}
	signer, err := readOCIPrivateKey(keyPath)
	if err != nil {
		return err
	}
	return oauthPacket.setTokenKeyValsForOAUTH(token, header, signer)
}

// applyAuthData adds generic OAuth authentication data using only the bearer
// token.
func (oauthTokenProvider) applyAuthData(oauthPacket *oAuth, token string, _ tokenProviderContext) error {
	return oauthPacket.setTokenKeyValsForOAUTH(token, "", nil)
}

// generateTokenHeader builds the OCI signed header using the service name and
// the connected remote network endpoint.
func (provider ociTokenProvider) generateTokenHeader(ctx tokenProviderContext) (string, error) {
	serviceName, err := extractServiceName(ctx.connectString)
	if err != nil {
		return "", err
	}
	if ctx.sessionContext == nil {
		return "", errors.New("missing session context")
	}
	remoteAddrValue := ctx.sessionContext.GetSessionProperties().GetProperty("REMOTE_ADDRESS")
	remoteAddr, ok := remoteAddrValue.(string)
	if !ok {
		return "", errors.New("missing remote address")
	}
	remoteAddr = strings.TrimSpace(remoteAddr)
	if remoteAddr == "" {
		return "", errors.New("missing remote address")
	}
	return fmt.Sprintf(
		"date: %s\n(request-target): %s\nhost: %s",
		time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"),
		serviceName,
		remoteAddr,
	), nil
}

// resolveTokenDirectory resolves a token directory from either an explicit
// location or a provider-specific default relative to the user's home
// directory.
func resolveTokenDirectory(tokenLocation, defaultRelativePath string) (string, error) {
	if tokenLocation != "" {
		info, err := os.Stat(tokenLocation)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return tokenLocation, nil
		}
		return filepath.Dir(tokenLocation), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, defaultRelativePath), nil
}

// resolveTokenLocation resolves a generic OAuth token location, accepting
// either a direct token file path or a directory containing a file named
// token.
func resolveTokenLocation(tokenLocation string) (string, error) {
	if strings.TrimSpace(tokenLocation) == "" {
		return "", errors.New("missing token location")
	}
	info, err := os.Stat(tokenLocation)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return filepath.Join(tokenLocation, tokenFileName), nil
	}
	return tokenLocation, nil
}

// resolveOCIPrivateKeyPath returns the OCI private key path associated with an
// OCI token location.
func resolveOCIPrivateKeyPath(tokenLocation string) (string, error) {
	tokenDir, err := resolveTokenDirectory(tokenLocation, filepath.Join(".oci", "db-token"))
	if err != nil {
		return "", err
	}
	return filepath.Join(tokenDir, ociPrivateKeyFileName), nil
}

// String returns a short diagnostic description of the token authenticator.
func (ta *TokenAuthenticator) String() string {
	return fmt.Sprintf("TokenAuthenticator{tokenAuthentication=%s, tokenLocation=%s}", ta.tokenAuthentication, ta.tokenLocation)
}

// extractServiceName returns SERVICE_NAME from an Oracle connect descriptor.
func extractServiceName(connectString string) (string, error) {
	common.Odl.Debug("ConnectionString", "connectString", connectString)
	return extractAddressValue(connectString, "SERVICE_NAME")
}

// extractAddressValue extracts a single "(KEY=value)" component from an Oracle
// connect descriptor.
func extractAddressValue(connectString, key string) (string, error) {
	upper := strings.ToUpper(connectString)
	idx := strings.Index(upper, key+"=")
	if idx == -1 {
		return "", fmt.Errorf("missing %s in connect descriptor", key)
	}
	start := idx + len(key) + 1
	end := strings.IndexAny(connectString[start:], ")")
	if end == -1 {
		return "", fmt.Errorf("unterminated %s in connect descriptor", key)
	}
	return strings.TrimSpace(connectString[start : start+end]), nil
}

// readOCIPrivateKey reads and validates the OCI database private key used to
// sign OCI IAM authentication headers.
func readOCIPrivateKey(path string) (crypto.Signer, error) {
	keyPEM, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, errors.New("invalid PEM private key")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	signer, ok := key.(crypto.Signer)
	if !ok {
		return nil, errors.New("private key does not implement crypto.Signer")
	}
	if _, ok := signer.(*rsa.PrivateKey); !ok {
		return nil, errors.New("OCI token authentication requires an RSA private key")
	}
	return signer, nil
}

// validateJWTExpiration checks the JWT exp claim when present and returns an
// error if the token has expired.
func validateJWTExpiration(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}

	common.Odl.Debug("JWTToken", "token", token, "payload", payload)
	var claims struct {
		Exp *int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == nil {
		return nil
	}
	if time.Unix(*claims.Exp, 0).Before(time.Now()) {
		return errors.New("configured access token has expired")
	}
	return nil
}

// signTokenHeader signs the OCI authentication header using SHA-256 and
// returns the Base64-encoded signature.
func signTokenHeader(header string, signer crypto.Signer) (string, error) {
	sum := sha256.Sum256([]byte(header))
	signature, err := signer.Sign(rand.Reader, sum[:], crypto.SHA256)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}
