package ttc

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oracle/go-driver/driver/common"
)

type mockTokenConn struct {
	remote string
}

func (m *mockTokenConn) Read([]byte) (int, error)         { return 0, os.ErrClosed }
func (m *mockTokenConn) Write(b []byte) (int, error)      { return len(b), nil }
func (m *mockTokenConn) Close() error                     { return nil }
func (m *mockTokenConn) LocalAddr() net.Addr              { return &net.TCPAddr{} }
func (m *mockTokenConn) SetDeadline(time.Time) error      { return nil }
func (m *mockTokenConn) SetReadDeadline(time.Time) error  { return nil }
func (m *mockTokenConn) SetWriteDeadline(time.Time) error { return nil }

func (m *mockTokenConn) RemoteAddr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", m.remote)
	if err != nil {
		return &net.TCPAddr{}
	}
	return addr
}

func TestGetAuthenticator_UsesTokenAuthenticatorForOCIToken(t *testing.T) {
	t.Parallel()

	cfg := common.NewOracleDriverConfig()
	cfg.Credentials.TokenAuthentication = common.TokenAuthenticationOCI
	cfg.Credentials.TokenLocation = t.TempDir()
	cfg.ConnectDescriptor = "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=127.0.0.1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=freepdb1)))"

	authenticator, err := GetAuthenticator(cfg)
	if err != nil {
		t.Fatalf("GetAuthenticator returned error: %v", err)
	}
	if _, ok := authenticator.(*TokenAuthenticator); !ok {
		t.Fatalf("expected TokenAuthenticator, got %T", authenticator)
	}
}

func TestGetAuthenticator_UsesTokenAuthenticatorForOAuth(t *testing.T) {
	t.Parallel()

	cfg := common.NewOracleDriverConfig()
	cfg.Credentials.TokenAuthentication = common.TokenAuthenticationOAuth
	cfg.Credentials.TokenLocation = filepath.Join(t.TempDir(), "token")
	cfg.ConnectDescriptor = "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=127.0.0.1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=freepdb1)))"

	authenticator, err := GetAuthenticator(cfg)
	if err != nil {
		t.Fatalf("GetAuthenticator returned error: %v", err)
	}
	if _, ok := authenticator.(*TokenAuthenticator); !ok {
		t.Fatalf("expected TokenAuthenticator, got %T", authenticator)
	}
}

func TestOCITokenProviderResolveTokenPath_Default(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("HOME", homeDir)

	got, err := (ociTokenProvider{}).resolveTokenPath("")
	if err != nil {
		t.Fatalf("resolveTokenPath returned error: %v", err)
	}

	want := filepath.Join(homeDir, ".oci", "db-token", tokenFileName)
	if got != want {
		t.Fatalf("resolveTokenPath = %q, want %q", got, want)
	}
}

func TestOAuthSetTokenKeyValsForOAUTH_AddsTokenHeaderAndSignature(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	oauth := NewOAuth().(*oAuth)
	oauth.keyValList = NewKeyValueList()
	header := "date: Mon, 10 Aug 2026 10:00:00 GMT\n(request-target): freepdb1\nhost: 127.0.0.1:1521"
	if err := oauth.setTokenKeyValsForOAUTH("token-value", header, privateKey); err != nil {
		t.Fatalf("setTokenKeyValsForOAUTH returned error: %v", err)
	}

	if oauth.keyValList.Len() != 3 {
		t.Fatalf("expected 3 key/value pairs, got %d", oauth.keyValList.Len())
	}

	got := map[string]string{}
	for e := oauth.keyValList.Front(); e != nil; e = e.Next() {
		kv := e.Value.(*common.KeyValue)
		got[common.B1ArrayToString(kv.Key)] = common.B1ArrayToString(kv.Value)
	}

	if got[authTokenKey] != "token-value" {
		t.Fatalf("AUTH_TOKEN = %q, want token-value", got[authTokenKey])
	}
	if got[authHeaderKey] != header {
		t.Fatalf("AUTH_HEADER = %q, want %q", got[authHeaderKey], header)
	}
	if got[authSignatureKey] == "" {
		t.Fatal("AUTH_SIGNATURE should not be empty")
	}
	if _, err := base64.StdEncoding.DecodeString(got[authSignatureKey]); err != nil {
		t.Fatalf("AUTH_SIGNATURE is not valid base64: %v", err)
	}
}

func TestOCITokenProviderGenerateTokenHeader(t *testing.T) {
	t.Parallel()

	sessContext := common.NewSessionContext()
	sessionProperties := common.NewProperties[string]()
	sessionProperties.SetProperty("REMOTE_ADDRESS", "192.0.2.10:1522")
	sessContext.UpdateSessionProperties(sessionProperties)

	provider := ociTokenProvider{}

	header, err := provider.generateTokenHeader(tokenProviderContext{
		connectString:  "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=adb.us-phoenix-1.oraclecloud.com)(PORT=1522))(CONNECT_DATA=(SERVICE_NAME=freepdb1)))",
		sessionContext: sessContext,
	})
	if err != nil {
		t.Fatalf("generateTokenHeader returned error: %v", err)
	}
	if !strings.Contains(header, "(request-target): freepdb1") {
		t.Fatalf("header missing service name: %q", header)
	}
	if !strings.Contains(header, "host: 192.0.2.10:1522") {
		t.Fatalf("header missing remote ip:port: %q", header)
	}
	if strings.Contains(header, "adb.us-phoenix-1.oraclecloud.com") {
		t.Fatalf("header should not use descriptor hostname for host: %q", header)
	}
	if !strings.Contains(header, " GMT\n(request-target): ") {
		t.Fatalf("header date should use GMT format: %q", header)
	}
	if strings.Contains(header, " UTC\n(request-target): ") {
		t.Fatalf("header date should not use UTC label: %q", header)
	}
}

func TestOAuthTokenProviderResolveTokenPath_File(t *testing.T) {
	t.Parallel()

	tokenFile := filepath.Join(t.TempDir(), "jwtbearertoken")
	if err := os.WriteFile(tokenFile, []byte("token-value"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	got, err := (oauthTokenProvider{}).resolveTokenPath(tokenFile)
	if err != nil {
		t.Fatalf("resolveTokenPath returned error: %v", err)
	}
	if got != tokenFile {
		t.Fatalf("resolveTokenPath = %q, want %q", got, tokenFile)
	}
}

func TestOAuthTokenProviderApplyAuthData_AddsTokenOnly(t *testing.T) {
	t.Parallel()

	oauth := NewOAuth().(*oAuth)
	oauth.keyValList = NewKeyValueList()

	if err := (oauthTokenProvider{}).applyAuthData(oauth, "token-value", tokenProviderContext{}); err != nil {
		t.Fatalf("applyAuthData returned error: %v", err)
	}

	if oauth.keyValList.Len() != 1 {
		t.Fatalf("expected 1 key/value pair, got %d", oauth.keyValList.Len())
	}

	kv := oauth.keyValList.Front().Value.(*common.KeyValue)
	if got := common.B1ArrayToString(kv.Key); got != authTokenKey {
		t.Fatalf("unexpected key %q, want %q", got, authTokenKey)
	}
	if got := common.B1ArrayToString(kv.Value); got != "token-value" {
		t.Fatalf("unexpected value %q, want %q", got, "token-value")
	}
}

func TestTokenAuthenticatorResolveAccessToken_UsesConfiguredAccessToken(t *testing.T) {
	t.Parallel()

	authenticator := NewTokenAuthenticator(common.TokenAuthenticationOAuth, " direct-token ", "", "")

	got, err := authenticator.resolveAccessToken()
	if err != nil {
		t.Fatalf("resolveAccessToken returned error: %v", err)
	}
	if got != "direct-token" {
		t.Fatalf("resolveAccessToken = %q, want %q", got, "direct-token")
	}
}

func TestTokenAuthenticatorResolveAccessToken_PrefersAccessTokenOverLocation(t *testing.T) {
	t.Parallel()

	tokenFile := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenFile, []byte("file-token"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	authenticator := NewTokenAuthenticator(common.TokenAuthenticationOAuth, "inline-token", tokenFile, "")

	got, err := authenticator.resolveAccessToken()
	if err != nil {
		t.Fatalf("resolveAccessToken returned error: %v", err)
	}
	if got != "inline-token" {
		t.Fatalf("resolveAccessToken = %q, want %q", got, "inline-token")
	}
}

func TestValidateJWTExpiration_Expired(t *testing.T) {
	t.Parallel()

	token := "eyJhbGciOiJub25lIn0.eyJleHAiOjF9."
	err := validateJWTExpiration(token)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired token error, got %v", err)
	}
}
