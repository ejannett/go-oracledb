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
	cfg.Credentials.TokenAuthentication = tokenAuthenticationOCI
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

func TestTokenAuthenticatorResolveTokenDirectory_Default(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("HOME", homeDir)

	authenticator := NewTokenAuthenticator(tokenAuthenticationOCI, "", "")
	got, err := authenticator.resolveTokenDirectory()
	if err != nil {
		t.Fatalf("resolveTokenDirectory returned error: %v", err)
	}

	want := filepath.Join(homeDir, ".oci", "db-token")
	if got != want {
		t.Fatalf("resolveTokenDirectory = %q, want %q", got, want)
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

func TestTokenAuthenticatorGenerateTokenHeader(t *testing.T) {
	t.Parallel()

	sessContext := common.NewSessionContext()
	sessionProperties := common.NewProperties[string]()
	sessionProperties.SetProperty("REMOTE_ADDRESS", "192.0.2.10:1522")
	sessContext.UpdateSessionProperties(sessionProperties)

	authenticator := NewTokenAuthenticator(tokenAuthenticationOCI, "", "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCPS)(HOST=adb.us-phoenix-1.oraclecloud.com)(PORT=1522))(CONNECT_DATA=(SERVICE_NAME=freepdb1)))")
	authenticator.SetSessionContext(sessContext)

	header, err := authenticator.generateTokenHeader()
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

func TestValidateJWTExpiration_Expired(t *testing.T) {
	t.Parallel()

	token := "eyJhbGciOiJub25lIn0.eyJleHAiOjF9."
	err := validateJWTExpiration(token)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired token error, got %v", err)
	}
}
