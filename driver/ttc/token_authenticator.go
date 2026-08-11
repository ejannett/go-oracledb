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
	tokenAuthenticationOCI = "OCI_TOKEN"
	authToken              = "AUTH_TOKEN"
	authHeader             = "AUTH_HEADER"
	authSignature          = "AUTH_SIGNATURE"
	ociTokenFileName       = "token"
	ociPrivateKeyFileName  = "oci_db_key.pem"
)

type TokenAuthenticator struct {
	username            common.B1Array
	tokenAuthentication string
	tokenLocation       string
	connectString       string
	logonMode           int64
	shelf               ttiShelf[common.MessageType]
	sessionContext      *common.SessionContext
}

func NewTokenAuthenticator(tokenAuthentication, tokenLocation, connectString string) *TokenAuthenticator {
	return &TokenAuthenticator{
		tokenAuthentication: strings.ToUpper(strings.TrimSpace(tokenAuthentication)),
		tokenLocation:       strings.TrimSpace(tokenLocation),
		connectString:       connectString,
		logonMode:           0x20000001,
	}
}

func (ta *TokenAuthenticator) SetShelf(shelf *ttiShelf[common.MessageType]) {
	ta.shelf = *shelf
}

func (ta *TokenAuthenticator) SetSessionContext(sessCtx *common.SessionContext) {
	ta.sessionContext = sessCtx
}

func (ta *TokenAuthenticator) Authenticate(ctx context.Context) error {
	common.Odl.Debug("Start TOKEN authentication")
	if ta.tokenAuthentication != tokenAuthenticationOCI {
		return common.NewOracleError(common.NoAuthenticatorError, nil, ta.tokenAuthentication)
	}

	tokenPath, err := ta.resolveTokenDirectory()
	if err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}

	tokenBytes, err := os.ReadFile(filepath.Join(tokenPath, ociTokenFileName))
	if err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return common.NewOracleError(common.AuthenticatorError, nil, "empty token file")
	}
	if err := validateJWTExpiration(token); err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}

	signer, err := readOCIPrivateKey(filepath.Join(tokenPath, ociPrivateKeyFileName))
	if err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}

	return ta.doOAuth(ctx, token, signer)
}

func (ta *TokenAuthenticator) doOAuth(ctx context.Context, token string, signer crypto.Signer) error {
	shelf := ta.shelf
	streamer := shelf.GetMessageStreamer().(MessageStreamerInterface)

	oauthMsg, err := shelf.GetMessageFactory().(Factory).GetMessageForFunction(TTIFUN, oauth)
	if err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}

	oauthPacket := oauthMsg.(*oAuth)
	oauthPacket.setConnectString(ta.connectString)
	oauthPacket.setLogonMode(ta.logonMode)
	if err := oauthPacket.prepareForTokenOAUTH(ta.username); err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}
	header, err := ta.generateTokenHeader()
	if err != nil {
		return common.NewOracleError(common.AuthenticatorError, err, nil)
	}
	if err := oauthPacket.setTokenKeyValsForOAUTH(token, header, signer); err != nil {
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

func (ta *TokenAuthenticator) resolveTokenDirectory() (string, error) {
	if ta.tokenLocation != "" {
		info, err := os.Stat(ta.tokenLocation)
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			return ta.tokenLocation, nil
		}
		return filepath.Dir(ta.tokenLocation), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".oci", "db-token"), nil
}

func (ta *TokenAuthenticator) generateTokenHeader() (string, error) {
	serviceName, err := extractServiceName(ta.connectString)
	if err != nil {
		return "", err
	}
	remoteAddr := ta.sessionContext.GetSessionProperties().GetProperty("REMOTE_ADDRESS")

	return fmt.Sprintf(
		"date: %s\n(request-target): %s\nhost: %s",
		time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"),
		serviceName,
		remoteAddr,
	), nil
}

func (ta *TokenAuthenticator) String() string {
	return fmt.Sprintf("TokenAuthenticator{tokenAuthentication=%s, tokenLocation=%s}", ta.tokenAuthentication, ta.tokenLocation)
}

func extractServiceName(connectString string) (string, error) {
	common.Odl.Debug("ConnectionString", "connectString", connectString)
	return extractAddressValue(connectString, "SERVICE_NAME")
}

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
		return errors.New("configured OCI token has expired")
	}
	return nil
}

func signTokenHeader(header string, signer crypto.Signer) (string, error) {
	sum := sha256.Sum256([]byte(header))
	signature, err := signer.Sign(rand.Reader, sum[:], crypto.SHA256)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}
