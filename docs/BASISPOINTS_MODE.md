# Basispoints 出站模式

jzg-lab 定制版新增的**账号级开关**：
让 OpenAI OAuth 账号的 `/v1/responses` 请求改走 OpenAI 内部的 basispoints 端点，
而不是默认的 `chatgpt.com/backend-api/codex`。

> ⚠️ **风险提示**
> - basispoints 是**未公开的内部端点**，并复用 ChatGPT 登录令牌调用，**很可能违反 OpenAI 服务条款，存在封号风险**。
> - 端点随时可能变更、加校验或下线。
> - 推理强度上限仍为 **`xhigh`，没有 `max`**。本开关只切换路由端点，不解锁更高推理档位。
> - 请只在自己的账号上、自担风险地使用。

## 行为说明

开关默认关闭；关闭时行为与改动前完全一致。

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

### 与账号保护（一键防降智 / 模式一）的关系

两者独立，可以同时开启：

- 模式一的请求完整性校验（`validateMode1StagedRequest`）仍在构造出站请求前执行，校验的是请求体，basispoints 不改请求体，因此不受影响。
- 账号保护策略配置的 Codex 身份收敛（`identity_mode`）**在 basispoints 模式下不会发送**：basispoints 端点不认 codex 身份头，所有 codex 专用头都被跳过。
- 策略里的 TLS 指纹模板和并发上限不受影响，照常生效。
- basispoints 模式强制 HTTP SSE，即使账号开启了 WebSocket v2 也不会走 WebSocket。

### 与 OpenAI OAuth 插件的关系

如果账号命中了 OpenAI OAuth 插件（插件管理里的灰度绑定），出站请求由插件进程代发。
插件收到的是已经改写为 basispoints URL / Host / 请求头的请求，但插件是否原样转发由插件自己决定。
测试 basispoints 时建议使用**未绑定插件**的账号，避免结果混淆。

## 开启方法

后台编辑账号时 `extra` 会被整体替换，推荐直接在数据库中合并写入（PostgreSQL）：

```sql
-- 开启
UPDATE accounts
SET extra = COALESCE(extra, '{}'::jsonb) || '{"openai_basispoints_mode": true}'::jsonb
WHERE id = <账号ID>;

-- 关闭
UPDATE accounts
SET extra = COALESCE(extra, '{}'::jsonb) || '{"openai_basispoints_mode": false}'::jsonb
WHERE id = <账号ID>;
```

使用 `deploy/compose.relay.yml` 部署时（数据库用户和库名默认均为 `relay`）：

```bash
cd deploy
docker compose --env-file .env -f compose.relay.yml exec postgres psql -U relay -d relay -c \
  "UPDATE accounts SET extra = COALESCE(extra,'{}'::jsonb) || '{\"openai_basispoints_mode\": true}'::jsonb WHERE id = 1;"
docker compose --env-file .env -f compose.relay.yml restart app
```

修改后需要重启应用，避免账号缓存导致不生效。建议先只对一个账号开启，确认返回正常后再扩大范围。

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

运行测试：

```bash
cd backend
go test ./internal/service/ -run Basispoints -v
```

## 与原版 Wei-Shaw/sub2api 的关系

本仓库（jzg-lab 定制版）与原版没有共同的 Git 历史，**不能直接 `git merge` 原版**。
本功能最初基于原版 0.2.8 实现，移植到本仓库时解决了 `buildUpstreamRequest` 中的两处差异
（本仓库特有的 `validateMode1StagedRequest` 调用和 4 参数的 `applyOpenCodeSessionHeader`）。
若以后要从原版移植其他改动，需逐项 cherry-pick 并单独验证。
