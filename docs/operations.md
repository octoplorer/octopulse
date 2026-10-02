# 部署与运维

本文中的源码构建和 Docker Compose 命令均在仓库根目录执行。独立二进制发行包按包目录执行，并使用包内实际的二进制路径。源码构建前，先按 [本地开发](development.md#本地开发) 准备开发工具和依赖。

## 构建和二进制运行

```sh
mise run build
./bin/octopulse serve
```

打开 `http://127.0.0.1:8080/app`。二进制和 `web/dist` 需要一起交付；以仓库根目录或包含 `web/dist` 的发行包目录作为工作目录，也可指定 `OCTOPULSE_STATIC_DIR` 的绝对路径。数据库迁移嵌入二进制，启动时取得实例锁后自动升级；数据库版本超过该二进制支持的版本时拒绝启动。

进程收到 SIGINT/SIGTERM 后停止接受新请求，等待已有 HTTP 请求和后台任务结束，再释放数据库；HTTP 排空上限为 10 秒，超时强制关闭连接。`GET /healthz` 检查数据库连接与实例锁，无法就绪时返回 503。手动检查通过 API 接受任务后由调度器执行，可在客户端等待结束后继续完成；轮次结果通过监控详情和历史查询。

## 配置

程序读取进程环境变量，不会自行加载 `.env`。二进制运行可复制 [.env.example](../.env.example) 后在 shell 或服务管理器中加载：

mise 只固定开发工具和任务，不覆盖已导出的数据库运行环境变量；未设置时使用下表中的 Go 默认配置。

```sh
cp .env.example .env
# 按实际部署编辑 .env，随后加载：
set -a
. ./.env
set +a
./bin/octopulse serve
```

Docker Compose 自动读取仓库根目录的 `.env`，用于 Compose 文件声明的变量。当前文件转发管理主机、Secure Cookie、加密密钥、数据库连接数、统计周期和 PostgreSQL 密码；其他应用环境变量应在服务的 `environment` 中配置。

| 变量                                    | 默认值 / 含义                                                                |
| --------------------------------------- | ---------------------------------------------------------------------------- |
| `OCTOPULSE_ADDR`                        | `127.0.0.1:8080`；容器内为 `0.0.0.0:8080`                                    |
| `OCTOPULSE_DB_DRIVER`                   | `sqlite`；另一选项是 `postgres`                                              |
| `OCTOPULSE_DB_DSN`                      | `file:.data/octopulse.db`；PostgreSQL 使用连接 URI                           |
| `OCTOPULSE_DATA_DIR`                    | `.data`；保存加密密钥及 `uploads/`                                           |
| `OCTOPULSE_STATIC_DIR`                  | `web/dist`；前端构建目录                                                     |
| `OCTOPULSE_ENCRYPTION_KEY`              | 32 字节随机值的 Base64 编码；留空时生成 `DATA_DIR/encryption.key`，权限 0600 |
| `OCTOPULSE_ADMIN_HOSTS`                 | `localhost,127.0.0.1`；允许后台/API 的精确主机名列表，以逗号分隔             |
| `OCTOPULSE_COOKIE_SECURE`               | `false`；HTTPS 部署设为 `true`                                               |
| `OCTOPULSE_DB_MAX_CONNECTIONS`          | `10`，范围 2–100；PostgreSQL 池含实例锁专用连接                              |
| `OCTOPULSE_STATISTICS_INTERVAL_SECONDS` | `60`，范围 5–3600；聚合及留存清理周期                                        |

组织名称、语言、时区、历史保留和可绑定的公开域名在后台设置中管理。个人语言和显示时区可在个人设置中覆盖。

## SQLite 与 PostgreSQL

SQLite 默认使用本地持久文件。程序统一配置 WAL、FULL 同步、外键和 busy timeout；一个写连接及只读池处理写入和查询。不要把数据库文件放在网络文件系统，也不要让多个服务进程共同打开同一数据库。运行时文件锁阻止重复实例。

PostgreSQL 启动示例：

```sh
OCTOPULSE_DB_DRIVER=postgres \
OCTOPULSE_DB_DSN='postgres://octopulse:YOUR_PASSWORD@127.0.0.1:5432/octopulse?sslmode=disable' \
./bin/octopulse serve
```

上例适用于本机连接；远程连接按数据库部署配置 TLS。连接用户需要应用表的读写和迁移权限。PostgreSQL advisory lock 保证同一数据库只有一个采集实例；丢失锁连接会停止服务。两种数据库使用分别维护的 goose 迁移和 sqlc 查询，共用业务、事务与统计语义，详见 [双数据库设计](database-design.md)。改变驱动不会自动迁移已有数据，首版不提供 SQLite/PostgreSQL 数据互转命令。

## Docker Compose

SQLite：

```sh
docker compose up --build -d
docker compose logs -f octopulse
```

访问 `http://127.0.0.1:8080/app`。默认仅向宿主机回环地址发布 8080 端口；`octopulse-data` 卷保存数据库、加密密钥和上传图片。镜像以 UID/GID 10001 运行。

PostgreSQL：在 `.env` 中设置 `POSTGRES_PASSWORD`，然后执行：

```sh
docker compose -f compose.yaml -f compose.postgres.yaml up --build -d
docker compose -f compose.yaml -f compose.postgres.yaml logs -f octopulse
```

覆盖文件使用 PostgreSQL 18.6，数据库数据保存在 `postgres-data` 卷，应用密钥和图片仍保存在 `octopulse-data`。此 Compose 的密码直接放入连接 URI，建议使用足够长的随机十六进制值，避免 URI 保留字符需要转义。容器仅暴露应用端口，数据库在 Compose 网络内访问。

## 状态页域名和反向代理

例如管理入口 `example.com/app`，状态页 slug `status1`：

1. 设置 `OCTOPULSE_ADMIN_HOSTS=example.com`、`OCTOPULSE_COOKIE_SECURE=true`。
2. 在后台系统设置允许 `status1.example.com`，在状态页草稿绑定该域名并发布。
3. 将 `status1.example.com` 的 A/AAAA 或 CNAME 指向部署入口。
4. 反向代理为 `example.com` 和 `status1.example.com` 提供 HTTPS，将请求转发到同一 Go 地址，保留原始 HTTP `Host`。

发布后 `https://example.com/status1` 和 `https://status1.example.com/` 共用页面配置与公共数据。事件子路由、资源和公共 API 均由同一服务处理，不需要另行部署前端。

服务用实际 `Host` 做精确匹配，不使用 `X-Forwarded-Host` 或 `Forwarded` 选择管理/页面入口。公开域名不能与管理主机重合；本地可用 `status.localhost` 测试域名入口。平台不修改 DNS，也不自动申请证书。草稿修改 slug、域名或自定义配置不改变已发布页面；显式发布时才原子切换入口与公共配置。

Logo 可上传 PNG/JPEG/GIF，或使用外部 HTTPS 图片；上传按实际图片内容校验。品牌、深浅色、服务公开别名及布局只作用于对应状态页。

## 备份与恢复

一次可恢复的备份包含数据库、原加密密钥和 `DATA_DIR/uploads`。数据库中的秘密使用 AES-256-GCM 加密；只恢复数据库或换用新密钥不能解密原凭据。若通过 `OCTOPULSE_ENCRYPTION_KEY` 提供密钥，应由部署的秘密管理方式单独备份该值，不会生成密钥文件。备份文件应保存到限制访问的目录。

SQLite 二进制部署先停止服务，再使用同一数据库与数据目录配置运行 CLI：

```sh
umask 077
backup_dir="backups/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$backup_dir"
./bin/octopulse backup "$backup_dir/octopulse.db"
if [ -f .data/encryption.key ]; then
  cp .data/encryption.key "$backup_dir/encryption.key"
fi
if [ -d .data/uploads ]; then
  tar -C .data -czf "$backup_dir/uploads.tar.gz" uploads
fi
```

`backup` 使用 SQLite `VACUUM INTO` 输出一致数据库，并设置权限 0600。CLI 也会取得实例锁，因此不能在已运行的服务旁另起进程执行。恢复时保持服务停止，把数据库恢复到配置的 DSN 文件，把原密钥和图片恢复到数据目录；新目录恢复最简单。保留原目录副本后再切换，并确认数据库及密钥由服务用户所有。

SQLite Compose 部署可在停止服务后备份完整数据卷：

```sh
umask 077
mkdir -p backups
docker compose stop octopulse
docker compose run --rm --no-deps -T --entrypoint tar octopulse \
  -C /data -czf - . > backups/octopulse-data.tar.gz
docker compose up -d octopulse
```

恢复到新的空 `octopulse-data` 卷，保持服务停止，然后执行：

```sh
docker compose run --rm --no-deps -T --entrypoint tar octopulse \
  -C /data -xzf - < backups/octopulse-data.tar.gz
docker compose up -d octopulse
```

PostgreSQL 使用同 major 或更新的 PostgreSQL 客户端执行 `pg_dump --format=custom --no-owner`，恢复到新建空库时使用 `pg_restore --no-owner --exit-on-error --dbname=目标库`。连接信息可通过 `PGHOST`、`PGPORT`、`PGUSER`、`PGDATABASE` 和受保护的密码文件提供，避免将密码写入命令参数。恢复前停止 Octopulse，恢复后使用原密钥和图片目录启动。数据库角色及数据库本身需单独创建；应用不会替代 PostgreSQL 的角色管理。

Compose PostgreSQL 备份示例：

```sh
umask 077
mkdir -p backups
docker compose -f compose.yaml -f compose.postgres.yaml stop octopulse
docker compose -f compose.yaml -f compose.postgres.yaml exec -T postgres \
  pg_dump -U octopulse -d octopulse --format=custom --no-owner > backups/octopulse.dump
docker compose -f compose.yaml -f compose.postgres.yaml run --rm --no-deps -T \
  --entrypoint tar octopulse -C /data -czf - . > backups/octopulse-data.tar.gz
docker compose -f compose.yaml -f compose.postgres.yaml up -d octopulse
```

在已创建的空 PostgreSQL 数据库恢复：

```sh
docker compose -f compose.yaml -f compose.postgres.yaml exec -T postgres \
  pg_restore -U octopulse --no-owner --exit-on-error --dbname=octopulse < backups/octopulse.dump
docker compose -f compose.yaml -f compose.postgres.yaml run --rm --no-deps -T \
  --entrypoint tar octopulse -C /data -xzf - < backups/octopulse-data.tar.gz
docker compose -f compose.yaml -f compose.postgres.yaml up -d octopulse
```

数据库备份恢复后先验证登录、秘密引用、监控检查和已发布页面，再保留或删除旧部署。默认历史留存为轮次 14 天、尝试 3 天、5 分钟聚合 90 天、小时聚合及状态区间 13 个月，可在系统设置修改。

## Beszel 接入

管理员在秘密凭据页保存专用 Hub 账号密码，然后在服务器页配置 Hub URL、普通账号和密码秘密引用。专用账号应只获得需要展示的系统读取权限；正常 `readonly` 账号已验证。凭据由后端解密使用，不传给 SPA。

当前兼容范围为 Beszel `0.20.x`，实测基线 `v0.20.0`。其他版本显示不兼容状态；MFA 或禁用密码认证的账号不在当前认证路径内。摘要、历史和容器标记来源、同步时间、失联与过期；只持久化最新摘要，历史按 Hub 数据按需查询并有缓存上限。具体协议、单位与真实 Hub 测试说明见 [Beszel 适配器说明](../internal/beszel/README.md)。
