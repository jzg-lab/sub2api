//go:build unit

package antigravity

import (
	"context"
	"net/url"
	"strings"
	"testing"
)

func TestPrivateOAuthConfigMissing(t *testing.T) {
	for _, tc := range []struct {
		name, id, secret, missing string
	}{
		{"both absent", "", "", AntigravityOAuthClientIDEnv},
		{"id absent", "  ", "test-secret", AntigravityOAuthClientIDEnv},
		{"secret absent", "test-client", "  ", AntigravityOAuthClientSecretEnv},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(AntigravityOAuthClientIDEnv, tc.id)
			t.Setenv(AntigravityOAuthClientSecretEnv, tc.secret)
			authURL, err := BuildAuthorizationURL("state", "challenge")
			if err == nil || !strings.Contains(err.Error(), tc.missing) || authURL != "" {
				t.Fatalf("expected %s error and no URL, got %q, %v", tc.missing, authURL, err)
			}
			// A nil HTTP client proves configuration fails before any network access.
			client := &Client{}
			if _, err := client.ExchangeCode(context.Background(), "code", "verifier"); err == nil || !strings.Contains(err.Error(), tc.missing) {
				t.Fatalf("exchange should fail before HTTP: %v", err)
			}
			if _, err := client.RefreshToken(context.Background(), "refresh"); err == nil || !strings.Contains(err.Error(), tc.missing) {
				t.Fatalf("refresh should fail before HTTP: %v", err)
			}
		})
	}
}

func TestPrivateOAuthConfigTrimsAndDoesNotExposeSecret(t *testing.T) {
	t.Setenv(AntigravityOAuthClientIDEnv, "  private-test-client  ")
	t.Setenv(AntigravityOAuthClientSecretEnv, "  private-test-secret  ")
	authURL, err := BuildAuthorizationURL("state", "challenge")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("client_id") != "private-test-client" || strings.Contains(authURL, "private-test-secret") {
		t.Fatal("authorization URL must contain only the trimmed client ID, not the secret")
	}
	secret, err := getClientSecret()
	if err != nil || secret != "private-test-secret" {
		t.Fatal("client secret must be read and trimmed from the environment")
	}
}
