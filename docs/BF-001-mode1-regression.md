# BF-001：模式一把合法协议转换误判为请求损失

日期：2026-09-12。状态：静态差异与完整 Forward 的本地记录器 A/B 复现已通过；兼容保护 v3 已修复、回归并部署。真实上游账号验收待用户提供账号。本文不把此前交接日志当作本次测试结果。

用户报告初代可用，升级版开启防降智后账号调用全部失败。已发现一条足以让常见请求在发送上游之前确定性失败的代码路径：升级版把原始请求与 OAuth 协议转换后的请求严格逐字段比较，但初代继承的转换本来就会改变这些字段。[V，E-01～E-05] 用户实际所有失败是否均来自该路径，仍需原失败请求及服务端日志核对。[T]

## 证据边界与来源

| 代号 | 来源 | 完整位置 |
| --- | --- | --- |
| B | 用户提供的初代源码 | `C:\Users\30214\Desktop\sub2本地\sub2初代\extracted\server-deployed-20260911-173007\source` |
| U | 用户提供的升级版冻结源码 | `C:\Users\30214\Desktop\sub2本地\sub2最新\sub2最新源码及项目交接\mode1-portable-20260911-190815\source` |
| N | 本次独立开发副本 | `C:\Users\30214\Desktop\sub2本地\sub2新站` |

初代 ZIP 已检查全部条目目标路径并解压到独立目录，原包保留；4208 条目，共 66,039,609 字节。ZIP SHA-256 为 `DB0A1CC5DD9419419B74476C2FF72050A577A38D6339EC95AA8112B7C1ED2AA2`。[V]

B、U 的 `openai_codex_transform.go` 哈希相同，均为 `39DA22A01C2A35C3EB626BCDDEDC1336A42308C3FF2C1D3D8CAE1EB35A6FEB5E`。U 新增的 `openai_mode1_integrity.go` 哈希为 `F3BB5E2302EF60F7E1994FE1F551AAAF2891BF8F391E63ED549FF231E9DF442C`。下面行号固定指向 U，避免 N 的后续编辑使证据漂移。[V]

| 证据 | 代码位置 | 可直接确认的事实 |
| --- | --- | --- |
| E-01 | [U / openai_gateway_forward.go:22](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/openai_gateway_forward.go:22) | `Forward` 最早保存原始请求；第 28～88 行之后才运行过滤和兼容归一化。 |
| E-02 | [U / openai_mode1_integrity.go:48](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/openai_mode1_integrity.go:48) | 既存 `model/input/instructions/reasoning/tools/tool_choice/parallel_tool_calls/text/previous_response_id/max_output_tokens/session` 任何值变化都会报 `MODE1_LOSSY_TRANSFORM`。 |
| E-03 | [U / openai_codex_transform.go:154](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/openai_codex_transform.go:154) | OAuth 不支持字段表包含 `max_output_tokens`；第 215 行循环删除。该转换在初代完全相同。 |
| E-04 | [U / openai_codex_transform.go:311](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/openai_codex_transform.go:311) | input 数组会做内容、工具及引用规范化；第 326 行起把字符串 input 转换为单条用户消息数组。 |
| E-05 | [U / openai_gateway_forward.go:1346](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/openai_gateway_forward.go:1346)；[U / openai_gateway_passthrough.go:585](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/openai_gateway_passthrough.go:585) | 两个出站请求构造入口在建立请求前执行严格检查，错误可发生于网络拨号之前。 |
| E-06 | [U / openai_ws_forwarder_ingress.go:475](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/openai_ws_forwarder_ingress.go:475) | WS 也比较原始 `trimmed` 和归一化/策略应用后的 payload，失败用策略关闭错误返回。第 683、969 行另有出站检查。 |
| E-07 | [U / openai_plugin_transport.go:20](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/openai_plugin_transport.go:20) | 模式一在全局 TLS 关闭时直接 `MODE1_TLS_DISABLED`；第 23 行插件路由命中时直接 `MODE1_PLUGIN_TRANSPORT_UNVERIFIED`。账号测试路径第 50、53 行相同。 |
| E-08 | [U / account_mode1_protection.go:118](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/account_mode1_protection.go:118) | 模式一预览只检查账号状态，没有访问全局 TLS 或当前插件路由；可以先成功应用、随后每个请求才因 E-07 失败。 |
| E-09 | [U / account_mode1_protection_test.go:183](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/backend/internal/service/account_mode1_protection_test.go:183) | 原守卫正例是原始 JSON 与自身比较；第 210 行 Forward 正例仅使用已经规范的消息数组，无输出预算、工具调用或续链字段。 |
| E-10 | [U / MODE1_LOCAL_HANDOFF.md](/C:/Users/30214/Desktop/sub2本地/sub2最新/sub2最新源码及项目交接/mode1-portable-20260911-190815/source/MODE1_LOCAL_HANDOFF.md) | 历史交接明确仅本地回环测试，未发送真实 OpenAI 请求、未验收全新服务器；模型质量改善未证明。 |
| E-11 | [本次升级版复现日志](/C:/Users/30214/Desktop/sub2本地/sub2新站/verification/rebuild/upgrade-regression-reproduction.log)；[本次复现测试](/C:/Users/30214/Desktop/sub2本地/.tools/upgraded-baseline/internal/service/relay_regression_reproduction_test.go:21) | 完整升级版 Forward 本地记录器测试四个子用例通过：两种输入在关闭保护时到达上游记录器；开启后均在调用上游前返回各自字段的 `MODE1_LOSSY_TRANSFORM`。使用 dummy 凭据，没有外网上游调用。 |
| E-12 | [修复后定向回归](../verification/rebuild/backend-focused-final.jsonl)；[后续 Mode1 回归](../verification/rebuild/mode1-agent-final.log)；[WSS 回归](../verification/rebuild/ws-v3-repeated.log)；[会话池回归](../verification/rebuild/ws-pool-repeated.log) | 定向联合回归包含 4 个包、781 项测试/子测试通过，0 失败、0 跳过；日志包含修复后完整 Forward 正例及真实语义损失负例。后续 Mode1、WSS 与会话池复测通过。这些是本地测试，不是实际供应商账号调用。 |
| E-13 | [服务端验收](../verification/rebuild/server-smoke-final.json)；[部署镜像状态](../verification/rebuild/final-state.json)；[源码比对](../verification/rebuild/source-comparison.json) | 15 项服务器 HTTP、鉴权与数据库支撑的 API 检查通过；记录的部署镜像健康；2254 个本地/服务端源码文件匹配，0 差异。验收脚本明确不包含上游账号调用。 |

## 发现与失败路径

| 发现 | Evidence → Finding → Path | 结论强度 |
| --- | --- | --- |
| F-01：合法 input 规范化被拒绝 | E-01、E-02、E-04、E-05、E-11 → 字符串与等价消息数组无法 DeepEqual → OAuth Responses 普通转发：stage → transform → build → 本地错误 | [V] 代码路径及本次完整 Forward 记录器 A/B 复现 |
| F-02：输出预算字段必被删除后又要求存在 | E-01～E-03、E-05、E-11 → OAuth 转换删除 `max_output_tokens`，守卫要求它保持原值 → 已规范 input 也会因预算字段而拒绝 | [V] 代码路径及本次完整 Forward 记录器 A/B 复现 |
| F-03：模型、工具、system 等兼容变换同样冲突 | E-02、E-04；Forward 第 379、389 行；transform 第 261、293 行 → 语法差异与真正内容损失未区分 → 需要映射/转换的请求失败 | [V] 比较冲突；[T] 各入口真实样本覆盖 |
| F-04：WS 复制了相同误判 | E-06 → 原始 payload 与策略/兼容处理结果直接比较 → 变更受保护字段即策略关闭；不能以 HTTP 修复推断 WS 已修复 | [V] 升级版代码路径；N 的本地 WSS 回归见 E-12；[T] 真实上游 WS 验收 |
| F-05：保护开启后运行配置冲突才暴露 | E-07、E-08 → 应用成功不代表传输条件成立 → 全局 TLS 关闭或命中插件路由的账号每次发请求才报错 | [V] 代码路径；[T] 是否符合用户原现场配置 |

初代没有 `openai_mode1_integrity.go`，全树检索无 `validateMode1` / `MODE1_LOSSY_TRANSFORM`；其 `openai_plugin_transport.go` 先尝试插件，未接管则使用原 HTTP 上游。升级在未改变现有 OAuth 转换语义的情况下插入了上述守卫及硬传输门槛。此差异可以解释“关闭保护可调用、开启保护失败”，不能单凭代码证明用户现场只有这一种原因。[V/H]

最小回归输入可以不带真实凭据，使用本地上游记录器执行完整 Forward：

```json
{"model":"gpt-5.2","input":"Return OK","stream":true}
```

```json
{"model":"gpt-5.2","input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"Return OK"}]}],"max_output_tokens":128,"stream":true}
```

前者触发 F-01，后者独立触发 F-02。应断言冻结 U 返回对应字段错误且上游调用次数为 0，N 返回成功且记录器实际收到请求；仅直接调用校验函数不能覆盖 Forward 接线。

2026-09-12 03:30:36 UTC 的本次复现已运行前半项（E-11）：`string_input/protection_off` 与 `token_budget/protection_off` 均 `upstream_attempted=true`；对应开启保护的子用例均 `upstream_attempted=false`，分别报 `input` 与 `max_output_tokens`。此处 PASS 表示“故障按预期被重现”，并非升级版可用或真实 OpenAI 账号已测试通过。

修复后 N 的 `TestMode1ForwardRegressionV2AndV3` 对 v2/v3 × 普通/透传 × 5类输入共20项通过，其中包括上述字符串 input 和 token预算请求；测试明确断言到达标准上游传输。`TestMode1SemanticGuardStillRejectsActualLoss` 覆盖真正的文本、工具、推理和关联丢失负例。证据为 `verification/rebuild/backend-focused-final.jsonl` 和后续 `mode1-agent-final.log`。[V]

WSS 双轮（含工具结果）和生命周期相关测试连续20次通过；v2/v3会话池隔离连续20次通过。见 `verification/rebuild/ws-v3-repeated.log`、`ws-pool-repeated.log`。服务器应用已从修复源码构建并健康运行，真实数据库支撑的15项HTTP/API验收通过，见 `server-smoke-final.json`。这些证据没有包含真实上游账号鉴权或模型质量评测。[V/T]

复现用测试源码另保存在 `verification/rebuild/upgrade-regression-reproduction_test.go.txt`，它应复制进升级冻结源码的独立副本 `backend/internal/service/` 并恢复 `.go` 后缀，再执行 `go test ./internal/service -run '^TestReproduceUpgradeMode1CompatibilityRegression$' -count=1 -v`；不要将这个“期望旧版失败”的测试直接放进修复源码编译。

## 修复审查约束

修复需要保留初代的协议兼容能力，同时识别真正的内容丢失。不能把删除整个守卫、移动快照到所有危险过滤之后、或在前后两边无条件重跑完整有损转换，当作已经证明“请求保真”。

| 允许变换的候选 | 必须同时验证的不变量 |
| --- | --- |
| 字符串 input ↔ 单条 user message；字符串 content ↔ input_text 数组 | 文本字节、角色及消息顺序保持；不得顺带删除其他消息。 |
| system 文本提升为 instructions / developer | 原 system 每段内容仍可找回，顺序明确；原 instructions 不能被覆盖。 |
| 工具结构扁平化、合法工具名映射、call_id 规范化 | 工具参数 schema、描述、调用参数、输出内容及调用关联一致；工具名和 ID 的映射成对且可逆。 |
| `max_output_tokens` 因 Codex OAuth 端点不支持而省略 | 限于明确端点能力例外；应在 UI/文档声明上游不支持该预算，不声称请求上限被严格执行；API Key 路径不可照搬删除。 |
| 已知模型别名与 reasoning 语法规范化 | 只能用已验证映射；原生模型标识变化、高推理变低推理、任意 reasoning 删除必须拒绝或明确暴露。 |
| 加入服务端必需的 store/stream、设备元数据 | 不影响既有用户输入、工具定义、上下文和推理要求；会话不能跨账号或用户混用。 |

必要负例：改用户文本、删除中间消息、删除 encrypted reasoning、丢 tool output、错配 call_id、删 tool schema、`high → low`、未授权模型替换。每个负例应断言未发上游，并与一个仅有合法表示差异的正例配对。若某模型不支持图片/工具，应明确能力错误或选择支持的路由，不能靠“归一化”偷偷删掉它们。

配置策略需在“应用预览/应用”和运行路径保持一致。允许保留初代普通传输/已配置插件时，状态必须说明实际采用的传输；选择严格 TLS 时，应提前显示冲突、提供可还原配置，而不是显示成功后把账号的全部调用变成失败。配置是否符合保护目标与账号凭据是否失效应分开报告。

## 本次验收矩阵

每个适用组合使用同一账号、模型、请求体，对保护关闭和开启执行交替 A/B 请求。记录请求摘要哈希、模型、入口、流模式、开始时间、总耗时、首 token 耗时、状态、错误层级、上游是否被调用、usage 与工具结果；不记录凭据、Cookie 或用户敏感正文。

| 编号 | 账号/入口 | 场景 | 成功证据 |
| --- | --- | --- | --- |
| A-01 | OAuth / Responses HTTP SSE | 规范消息、字符串 input、显式空 instructions | 真实 SSE 完成事件、内容与 usage；本地不得误拦截。 |
| A-02 | OAuth / Responses HTTP 非流 | 与 A-01 相同输入，`stream:false` | 对外完整 JSON；内部上游 SSE 可正确聚合。 |
| A-03 | OAuth / Responses HTTP | 输出预算、reasoning、受支持工具 schema | 请求到达上游；可解释能力转换；无无声推理降低。 |
| A-04 | OAuth / Responses WS | 首轮与多轮、连接复用、重连 | 101 后真实 response 完成；每轮上下文、账户和会话关联正确。 |
| A-05 | OAuth / HTTP 与 WS | function_call → function_call_output → 最终回答 | 工具名称与 call_id 成对；第二轮引用前轮结果；无孤立输出或工具丢失。 |
| A-06 | OAuth / Chat Completions、Messages 兼容入口 | system + user + tool 组合 | 兼容转换后真实模型完成；所有角色内容保留。 |
| A-07 | API Key / Responses SSE 与非流 | 规范/字符串 input、输出预算、工具 | API Key 行为不受 OAuth 特例污染；输出预算未被错误删除。 |
| A-08 | API Key / Chat Completions | 流式、非流式、多轮和工具 | 可用响应与 usage；现有客户端可直接接入。 |
| A-09 | API Key / WS（仅账户/供应商支持时） | 首轮、工具续链 | 支持时完成真实请求；不支持时明确 capability 状态，不伪称通过。 |
| A-10 | 两类账号 / 真实负例 | 无效模型、失效凭据、429、网络失败 | 错误准确归因；可重试与不可重试分开；不把本地兼容错误累计为账号失效。 |
| A-11 | OAuth / 传输配置 | 直连、现有固定代理、全局 TLS 开/关、已配置插件 | 实际路由与配置预览一致；证书失败不绕过校验；普通可用路径不被无关开关封死。 |
| A-12 | OAuth / 并发与会话 | 同账号两会话、两账号同客户端 ID、重启后复用 | 稳定设备种子持久；对话隔离；并发上限执行；未跨账号复用上下文。 |

上游可达、HTTP 200、回环握手与固定设备身份均不能独立证明模型质量提升。质量验证另需固定题集、模型/推理参数及账号控制变量，比较准确率、工具完成率、输出完整性和失败率；不以单次“回答正常”作为质量结论。

## 验证状态与回滚

| 项目 | 本文写入时状态 |
| --- | --- |
| 初代安全解压、源码哈希与差异 | `tool-verified`，本次执行 |
| 升级版完整 Forward 复现 | `tool-verified`，E-11；四个子用例通过，证明两条本地误拦截；不是外网账号测试 |
| N 修复后回归测试 | `tool-verified`，E-12；4 包、781 项测试/子测试通过，后续 Mode1、WSS 与会话池复测通过；不是全仓库测试或真实上游验收 |
| N 服务端部署与接口验收 | `tool-verified`，E-13；修复源码已构建部署，15 项检查通过；镜像标签为 `chengchuan-relay:verified-20260912` |
| 真实上游账号 A/B | 未执行；测试服务器是新装环境，仍需可用测试账号。网络可达不等于账号调用成功 |
| 独立审查 | 本报告提供代码证据与实际测试日志；未附独立审查报告或签署，不以测试通过代替独立审查 |
| 历史 `verification/` 内容 | 从 U 复制的旧日志仅作历史记录，不计入本次通过数；本次证据单列在 `verification/rebuild/` |
| Git 提交/标签 | 未创建，不能宣称有可用 commit 回退点 |
| 服务端回滚 | 当前部署镜像标签及 ID 已记录于 E-13；操作步骤见[部署与恢复说明](RELAY_DEPLOYMENT.md)。本次未提供回滚演练或数据库备份/恢复成功证据，不能将现有镜像记录当作数据恢复验证 |

本次 N 已完成 service 与 UI 修复并部署，原始 B、U 和初代归档保持保留。核心补丁为：[账号保护 v3](../backend/internal/service/account_mode1_protection.go) 使用初代标准传输，显式迁移 v2 并保留身份种子和还原快照；[语义校验](../backend/internal/service/openai_mode1_integrity.go)及[Codex 规范化](../backend/internal/service/openai_mode1_semantics.go) 区分合法表示变化与实际内容损失；[出站传输](../backend/internal/service/openai_plugin_transport.go) 修复全局 TLS 开关和插件路径的硬阻断，并同步 HTTP、账号测试及 WS 调用路径。站点与保护界面变更、浏览器验收范围见[重建记录](REBUILD_PROGRESS.md)与[页面验收](BROWSER_VERIFICATION.md)。

原始 B、U 可用于逐文件恢复新副本或重新构建基线；账号保护配置通过“还原”恢复其保存的原始快照。服务器业务数据恢复必须依赖实际生成并验证的数据库与私有配置备份，不能用源码包或镜像标签替代。仍需可用测试账号完成上表适用的真实上游 A/B，才能评估实际调用成功率与模型输出。
