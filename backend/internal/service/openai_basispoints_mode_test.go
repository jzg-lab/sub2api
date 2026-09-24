package service

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func buildBasispointsTestRequest(t *testing.T, account *Account) *http.Request {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := &OpenAIGatewayService{}
	body := []byte(`{"model":"gpt-5.6","stream":true,"prompt_cache_key":"client-session"}`)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Set("api_key", &APIKey{ID: 77})
	c.Request.Header.Set("User-Agent", "codex_cli_rs/0.144.0")
	c.Request.Header.Set("originator", "codex_cli_rs")
	c.Request.Header.Set("x-codex-installation-id", "client-installation")
	c.Request.Header.Set("x-codex-window-id", "client-window")
	c.Request.Header.Set("x-codex-turn-metadata", `{"turn_id":"client-turn"}`)

	req, err := svc.buildUpstreamRequest(context.Background(), c, account, body, "oauth-token", true, "client-session", true)
	require.NoError(t, err)
	return req
}

func TestBuildUpstreamRequestBasispointsModeRoutesToBasispoints(t *testing.T) {
	account := &Account{
		ID:          11,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-account-11"},
		Extra:       map[string]any{"openai_basispoints_mode": true},
	}

	req := buildBasispointsTestRequest(t, account)

	require.Equal(t, openaiBasispointsURL, req.URL.String())
	require.Equal(t, "bps.openai.com", req.Host)
	require.Equal(t, "Bearer oauth-token", req.Header.Get("Authorization"))
	require.Equal(t, "chatgpt-account-11", req.Header.Get("chatgpt-account-id"))
	require.Equal(t, "chatgpt-account-11", req.Header.Get("x-openai-account-id"))
	require.Equal(t, "chatgpt", req.Header.Get("x-basispoints-auth-mode"))
	require.Equal(t, "text/event-stream", req.Header.Get("accept"))

	for _, header := range []string{
		"originator", "OpenAI-Beta", "version", "session_id", "conversation_id",
		"x-codex-installation-id", "x-codex-window-id", "x-codex-turn-metadata",
		"x-codex-turn-state", "x-codex-beta-features",
	} {
		require.Empty(t, req.Header.Get(header), "codex header must not reach basispoints: %s", header)
	}
}

func TestBuildUpstreamRequestBasispointsModeOffKeepsCodexRoute(t *testing.T) {
	account := &Account{
		ID:          11,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"chatgpt_account_id": "chatgpt-account-11"},
	}

	req := buildBasispointsTestRequest(t, account)

	require.Equal(t, chatgptCodexURL, req.URL.String())
	require.Equal(t, "chatgpt.com", req.Host)
	require.Equal(t, "chatgpt-account-11", req.Header.Get("chatgpt-account-id"))
	require.NotEmpty(t, req.Header.Get("originator"))
	require.Empty(t, req.Header.Get("x-basispoints-auth-mode"))
	require.Empty(t, req.Header.Get("x-openai-account-id"))
}

func TestIsOpenAIBasispointsModeEnabledIgnoresNonOpenAIAndNonBool(t *testing.T) {
	require.False(t, (*Account)(nil).IsOpenAIBasispointsModeEnabled())
	require.False(t, (&Account{Platform: PlatformAnthropic, Extra: map[string]any{"openai_basispoints_mode": true}}).IsOpenAIBasispointsModeEnabled())
	require.False(t, (&Account{Platform: PlatformOpenAI, Extra: map[string]any{"openai_basispoints_mode": "true"}}).IsOpenAIBasispointsModeEnabled())
	require.True(t, (&Account{Platform: PlatformOpenAI, Extra: map[string]any{"openai_basispoints_mode": true}}).IsOpenAIBasispointsModeEnabled())
}

func TestOpenAIWSProtocolResolverForcesHTTPForBasispointsMode(t *testing.T) {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.Enabled = true
	cfg.Gateway.OpenAIWS.OAuthEnabled = true
	cfg.Gateway.OpenAIWS.ResponsesWebsocketsV2 = true

	account := &Account{
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Extra: map[string]any{
			"openai_basispoints_mode":                      true,
			"openai_oauth_responses_websockets_v2_enabled": true,
		},
	}

	resolver := NewOpenAIWSProtocolResolver(cfg)
	decision := resolver.Resolve(account)
	require.Equal(t, OpenAIUpstreamTransportHTTPSSE, decision.Transport)
	require.Equal(t, "account_basispoints_mode", decision.Reason)

	account.Extra["openai_basispoints_mode"] = false
	require.Equal(t, OpenAIUpstreamTransportResponsesWebsocketV2, resolver.Resolve(account).Transport)
}
