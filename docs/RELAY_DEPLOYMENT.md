# 澄川测试部署与恢复

> 2026-09-15 已更新到 `optional-controls-20260915`。当前 Compose 目录为 `/opt/chengchuan-relay/releases/optional-controls-20260915/source/deploy`。本次备份、验证与回滚见 [最新部署记录](../SERVER_DEPLOYMENT_2026-09-15.md)。下文是早期部署记录，不应据其旧镜像标签直接更新当前站点。

## 当前部署

服务器为用户提供的新装 Ubuntu 24.04，4 核、约 8 GiB 内存；安装了发行版 Docker、Compose 和 Buildx。实际从源码构建应用，使用 PostgreSQL 18 和 Redis 7。应用进程 UID/GID=1000，NoNewPrivs=1；仅应用8080端口对外发布，数据库/Redis没有主机端口映射。

目录为 `/opt/chengchuan-relay/`：`source/` 源码、`source/deploy/.env` 私有设置、`evidence/` 构建和验收日志、`PRIVATE_ACCESS.txt` 管理员访问资料。秘密文件权限600，不包含在本机源码包内。当前固定镜像标签 `chengchuan-relay:verified-20260912`，镜像ID见 `verification/rebuild/final-state.json`；普通界面测试用户已经停用、软删除，登录返回401。

## 新服务器安装

需要 Linux、Docker/Compose、Python3 和外网依赖下载能力。在源码目录执行：

```bash
cd deploy
python3 init-relay-env.py --bind 127.0.0.1
sh build-relay.sh
```

初次构建会从锁文件安装前端依赖并下载 Go 模块。测试需要公网预览时，可以显式将私有 `.env` 的 `BIND_HOST` 设为 `0.0.0.0`。正式部署应由反向代理提供 HTTPS，不能将测试 HTTP 入口视作正式安全入口。

## 检查运行状态

```bash
cd /opt/chengchuan-relay/source/deploy
docker compose --env-file .env -f compose.relay.yml ps
curl --fail http://127.0.0.1:8080/health
python3 smoke-relay.py --env .env
docker compose --env-file .env -f compose.relay.yml logs --tail 100 app
```

`smoke-relay.py` 验证健康、前端壳、鉴权拒绝、管理员登录和真实数据库支撑的用户统计/密钥/分组接口。它不会输出密码、JWT 或账户资料，也不声称替代真实上游请求。

## 数据备份

```bash
cd /opt/chengchuan-relay/source/deploy
mkdir -p /opt/chengchuan-relay/backups
umask 077
docker compose --env-file .env -f compose.relay.yml exec -T postgres pg_dump -U relay -d relay -Fc > /opt/chengchuan-relay/backups/relay.dump
```

备份含业务数据，应作为私有文件保管；将私有 `.env` 单独备份，保留JWT/TOTP等密钥。不可用 `docker compose down -v` 做普通更新，因为它会删除数据库卷。

## 应用回滚

更新前记录 `docker inspect chengchuan-relay-app-1 --format '{{.Image}}'` 并为该镜像添加独立标签。将 `.env` 中 `RELAY_VERSION` 改为保留镜像的标签，再执行：

```bash
docker compose --env-file .env -f compose.relay.yml up -d --no-build --wait
```

本次未新增数据库迁移。后续若有迁移，需要先验证旧二进制与新schema兼容；数据库恢复不是常规镜像回滚步骤。

账号保护的业务回滚应使用账号编辑页“还原”按钮，保留原始快照；不应在数据库中盲目删除全部 `extra`。原始用户提供的初代归档与升级版目录仍在本地保留。

## 未完成的外部验收

当前没有真实上游账号。需用户通过SSH转发或HTTPS登录后台，处理首次管理员确认，导入可用测试账号、分组并告知可以开始对照。对照至少包含保护关/v3开、SSE与非流式、工具结果、多轮上下文、WSS；记录上游状态、返回模型和错误归属。

暂未执行第三方支付、真实充值、发信、外部OAuth登录、公开收费或模型质量对照。首页不宣传这些环节已经验收。
