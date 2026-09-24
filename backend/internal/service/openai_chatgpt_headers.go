package service

import (
	"context"
	"net/http"
)

func setOpenAIChatGPTAccountHeaders(headers http.Header, account *Account) {
	if headers == nil || account == nil || !account.IsOpenAIOAuthLike() {
		return
	}
	if chatgptAccountID := account.GetChatGPTAccountID(); chatgptAccountID != "" {
		headers.Set("chatgpt-account-id", chatgptAccountID)
	}
	if account.IsChatGPTAccountFedRAMP() {
		headers.Set("x-openai-fedramp", "true")
	} else {
		headers.Del("x-openai-fedramp")
	}
}

// resolveAndSetOpenAIChatGPTAccountHeaders 解析 spark 影子账号至其母账号（凭据透传），
// 再调用 setOpenAIChatGPTAccountHeaders 写入 chatgpt-account-id / x-openai-fedramp 头。
// 普通账号（非影子）为直通，行为与直接调用 setOpenAIChatGPTAccountHeaders 一致。
func resolveAndSetOpenAIChatGPTAccountHeaders(ctx context.Context, repo AccountRepository, headers http.Header, account *Account) error {
	credAccount, err := resolveCredentialAccount(ctx, repo, account)
	if err != nil {
		return err
	}
	setOpenAIChatGPTAccountHeaders(headers, credAccount)
	return nil
}

// setOpenAIBasispointsHeaders 写入 basispoints 端点所需的身份头。
// 端点以 chatgpt auth_mode 鉴权（Authorization Bearer 已在上游设置），另需：
//   - chatgpt-account-id / x-openai-account-id：账号标识（同源自 chatgpt_account_id）
//   - x-basispoints-auth-mode: chatgpt：声明以 ChatGPT 身份调用
//
// 与 codex 端点不同，此处不发送 originator/version/OpenAI-Beta 等 codex 私有头。
func setOpenAIBasispointsHeaders(headers http.Header, account *Account) {
	if headers == nil || account == nil || !account.IsOpenAIOAuthLike() {
		return
	}
	if accountID := account.GetChatGPTAccountID(); accountID != "" {
		headers.Set("chatgpt-account-id", accountID)
		headers.Set("x-openai-account-id", accountID)
	}
	headers.Set("x-basispoints-auth-mode", "chatgpt")
	if account.IsChatGPTAccountFedRAMP() {
		headers.Set("x-openai-fedramp", "true")
	} else {
		headers.Del("x-openai-fedramp")
	}
}

// applyOpenAIBasispointsHeaders 解析 spark 影子账号至母账号后写入 basispoints 身份头。
// 与 resolveAndSetOpenAIChatGPTAccountHeaders 对称，供 basispoints 出站路径使用。
func applyOpenAIBasispointsHeaders(ctx context.Context, repo AccountRepository, headers http.Header, account *Account) error {
	credAccount, err := resolveCredentialAccount(ctx, repo, account)
	if err != nil {
		return err
	}
	setOpenAIBasispointsHeaders(headers, credAccount)
	return nil
}
