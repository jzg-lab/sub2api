//go:build unit

package service

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
)

func TestAntigravityOAuthService_PrivateCredentials(t *testing.T) {
	svc := NewAntigravityOAuthService(nil)
	defer svc.sessionStore.Stop()
	t.Setenv(antigravity.AntigravityOAuthClientIDEnv, "")
	t.Setenv(antigravity.AntigravityOAuthClientSecretEnv, "")
	result, err := svc.GenerateAuthURL(context.Background(), nil)
	if err == nil || result != nil || !strings.Contains(err.Error(), antigravity.AntigravityOAuthClientIDEnv) {
		t.Fatalf("expected configuration error and no session result: %v", err)
	}
	t.Setenv(antigravity.AntigravityOAuthClientIDEnv, "test-client")
	t.Setenv(antigravity.AntigravityOAuthClientSecretEnv, "test-secret")
	result, err = svc.GenerateAuthURL(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(result.AuthURL)
	if err != nil || parsed.Query().Get("client_id") != "test-client" {
		t.Fatal("authorization URL did not use the private client")
	}
	if _, ok := svc.sessionStore.Get(result.SessionID); !ok {
		t.Fatal("successful authorization must save its session")
	}
}

func TestGeminiOAuthService_PrivateCredentials(t *testing.T) {
	svc := NewGeminiOAuthService(nil, nil, nil, nil, &config.Config{})
	defer svc.Stop()
	t.Setenv(geminicli.GeminiCLIOAuthClientIDEnv, "")
	t.Setenv(geminicli.GeminiCLIOAuthClientSecretEnv, "")
	for _, kind := range []string{"code_assist", "google_one"} {
		result, err := svc.GenerateAuthURL(context.Background(), nil, "", "", kind, "")
		if err == nil || result != nil || !strings.Contains(err.Error(), geminicli.GeminiCLIOAuthClientIDEnv) {
			t.Fatalf("%s should report private credential configuration error: %v", kind, err)
		}
	}
}
