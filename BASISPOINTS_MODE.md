# Basispoints 出站模式（本 fork 新增）

本 fork 基于 [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api)，新增一个**账号级开关**：
让 OpenAI OAuth 账号的 `/v1/responses` 请求改走 OpenAI 内部的 basispoints 端点，
而不是默认的 `chatgpt.com/backend-api/codex`。

> ⚠️ **风险提示**
> - basispoints 是**未公开的内部端点**，并复用 ChatGPT 登录令牌调用，**很可能违反 OpenAI 服务条款，存在封号风险**。
> - 端点随时可能变更、加校验或下线。
> - 推理强度上限仍为 **`xhigh`，没有 `max`**。本开关只切换路由端点，不解锁更高推理档位。
> - 请只在自己的账号上、自担风险地使用。

## 行为说明

开关默认关闭；关闭时行为与上游 sub2api 完全一致。

对开启开关的 OpenAI OAuth / Setup Token 账号：

| 项目 | 默认（codex） | basispoints 模式 |
|---|---|---|
| 上游 URL | `https://chatgpt.com/backend-api/codex/responses` | `https://bps.openai.com/basispoints/api/responses` |
| Host | `chatgpt.com` | `bps.openai.com` |
| 鉴权 | `Authorization: Bearer <access_token>` | 相同 |
| 账号头 | `chatgpt-account-id` | `chatgpt-account-id` + `x-openai-account-id` + `x-basispoints-auth-mode: chatgpt` |
| Codex 身份头（`originator` / `version` / `OpenAI-Beta` / `x-codex-*` / 指纹 / 会话头） | 发送 | **不发送**（客户端透传的也会被剥离） |
| 上游传输 | 可能走 WebSocket v2 | **强制 HTTP SSE** |
| 请求体 | Codex OAuth 转换 | 相同（未改动） |

API Key 账号不受影响。

## 开启方法

在管理后台操作即可，不需要命令行：

1. 进入 **账号管理**，找到要开启的 OpenAI 账号（类型为 OAuth 或 Setup Token），点 **编辑**。
2. 在编辑窗口中找到 **Basispoints 出站模式** 开关（位于"摊平 Codex namespace 工具"开关下方），打开它。
3. 点 **保存**。保存后立即生效，不需要重启服务。

关闭时同样在编辑窗口里把开关关掉并保存。API Key 账号不显示这个开关。

建议先只对一个账号开启，确认返回正常后再扩大范围。

<details>
<summary>也可以直接改数据库（PostgreSQL）</summary>

```sql
-- 开启
UPDATE accounts
SET extra = COALESCE(extra, '{}'::jsonb) || '{"openai_basispoints_mode": true}'::jsonb
WHERE id = <账号ID>;

-- 关闭
UPDATE accounts SET extra = extra - 'openai_basispoints_mode' WHERE id = <账号ID>;
```

直接改数据库不会通知调度缓存，改完需要重启 sub2api 服务才会生效。

</details>

## 验证是否生效

以 debug 模式运行时，上游请求日志中的 URL 应为 `bps.openai.com/basispoints/api/responses`，
WebSocket 协议决策原因为 `account_basispoints_mode`。

## 代码改动

| 文件 | 改动 |
|---|---|
| `backend/internal/service/openai_gateway_service.go` | 新增常量 `openaiBasispointsURL` |
| `backend/internal/service/account.go` | 新增 `IsOpenAIBasispointsModeEnabled()`，读取 `extra.openai_basispoints_mode` |
| `backend/internal/service/openai_chatgpt_headers.go` | 新增 `setOpenAIBasispointsHeaders` / `applyOpenAIBasispointsHeaders` |
| `backend/internal/service/openai_gateway_forward.go` | `buildUpstreamRequest` 中按开关切换 URL / Host / 头，门控所有 codex 专用头 |
| `backend/internal/service/openai_ws_protocol_resolver.go` | 开关开启时强制 HTTP |
| `backend/internal/service/openai_basispoints_mode_test.go` | 新增单元测试 |
| `frontend/src/components/account/EditAccountModal.vue` | 账号编辑窗口新增开关 |
| `frontend/src/i18n/locales/{zh,en}/admin/accounts.ts` | 开关的中英文文案 |
| `frontend/src/components/account/__tests__/EditAccountModal.spec.ts` | 开关的前端测试 |

运行测试：

```bash
cd backend
go test ./internal/service/ -run Basispoints -v

cd ../frontend
pnpm exec vitest run src/components/account/__tests__/EditAccountModal.spec.ts
```

## 同步上游

```bash
git remote add upstream https://github.com/Wei-Shaw/sub2api.git
git fetch upstream
git merge upstream/main
```

改动集中在上述文件中，合并冲突时以保留 `basispoints` / `codexProtocol` 门控逻辑为准。
