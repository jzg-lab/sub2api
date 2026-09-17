//go:build unit

package geminicli

import (
	"strings"
	"testing"
)

func TestPrivateOAuthConfigMissing(t *testing.T) {
	for _, tc := range []struct {
		name, id, secret, missing string
	}{
		{"both absent", "", "", GeminiCLIOAuthClientIDEnv},
		{"id absent", " ", "test-secret", GeminiCLIOAuthClientIDEnv},
		{"secret absent", "test-client", " ", GeminiCLIOAuthClientSecretEnv},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(GeminiCLIOAuthClientIDEnv, tc.id)
			t.Setenv(GeminiCLIOAuthClientSecretEnv, tc.secret)
			_, err := EffectiveOAuthConfig(OAuthConfig{}, "code_assist")
			if err == nil || !strings.Contains(err.Error(), tc.missing) {
				t.Fatalf("expected %s error, got %v", tc.missing, err)
			}
			authURL, err := BuildAuthorizationURL(OAuthConfig{}, "state", "challenge", GeminiCLIRedirectURI, "", "code_assist")
			if err == nil || authURL != "" {
				t.Fatalf("missing configuration must not produce an authorization URL: %q, %v", authURL, err)
			}
		})
	}
}

func TestPrivateOAuthConfigTrimsAndPreservesCustomPrecedence(t *testing.T) {
	t.Setenv(GeminiCLIOAuthClientIDEnv, "  private-test-client  ")
	t.Setenv(GeminiCLIOAuthClientSecretEnv, "  private-test-secret  ")
	cfg, err := EffectiveOAuthConfig(OAuthConfig{}, "code_assist")
	if err != nil || cfg.ClientID != "private-test-client" || cfg.ClientSecret != "private-test-secret" {
		t.Fatal("expected trimmed environment credentials")
	}
	t.Setenv(GeminiCLIOAuthClientIDEnv, "")
	t.Setenv(GeminiCLIOAuthClientSecretEnv, "")
	cfg, err = EffectiveOAuthConfig(OAuthConfig{ClientID: "custom-id", ClientSecret: "custom-secret"}, "ai_studio")
	if err != nil || cfg.ClientID != "custom-id" || cfg.ClientSecret != "custom-secret" || cfg.Scopes != DefaultAIStudioScopes {
		t.Fatal("explicit custom configuration must work without CLI environment credentials")
	}
}
