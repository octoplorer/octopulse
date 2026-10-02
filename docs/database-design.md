# SQLite 与 PostgreSQL 持久层方案

状态：2026-10-02 文档核对时，双数据库持久层已纳入 `main`。首版实施阶段记录了完整 SQLite/真实 PostgreSQL18.6 竞态套件、两库容量及 arm64 容器启动、持久化与完整备份恢复通过，见 [验收映射](acceptance.md)。本文按当前代码描述版本、目录与运行细节；历史验收不代表当前提交已重跑这些检查，也不将本地负载结果推导为任意生产负载的性能保证。

## 共享业务模型与方言边界

一个部署只连接一个数据库，由启动配置选择 `sqlite` 或 `postgres`。同一 Go 二进制包含两种驱动，默认 SQLite；切换连接配置只会打开另一个数据库，不会自动迁移已有数据。

首版只运行一个 Go 服务实例。SQLite 在规范化后的数据库路径旁取得非阻塞进程文件锁；PostgreSQL 使用独立会话持有应用 advisory lock。运行锁在迁移前取得，覆盖整个服务生命周期；PostgreSQL 每 5 秒检查锁连接，丢失时停止服务。不同 schema 也受同一数据库运行锁约束，不将 PostgreSQL 支持解释为多副本调度支持。

监控状态机、阈值、维护规则、时间统计、通知策略和权限只实现一次。持久层提供统一的业务数据类型和事务入口，两种适配器处理 SQL、列类型及错误转换，不把数据库类型判断散落在业务服务中。

| 层次            | 实际选择                         | 作用                               |
| --------------- | -------------------------------- | ---------------------------------- |
| 连接与事务      | Go `database/sql`                | 共用连接池、context 和事务编程接口 |
| SQLite 驱动     | `modernc.org/sqlite`             | CGO-free，便于交付 Go 二进制       |
| PostgreSQL 驱动 | `github.com/jackc/pgx/v5/stdlib` | 将 pgx 接入 `database/sql`         |
| 查询代码        | sqlc                             | 分别校验和生成两种方言的带类型查询 |
| Schema 迁移     | goose                            | 按数据库版本执行嵌入二进制的迁移   |

sqlc 支持在同一配置中设置多个 engine/schema/query/output，Go 可使用 `database/sql`。它不承担把任意 PostgreSQL SQL 翻译成 SQLite SQL 的职责；项目维护两套方言查询，使用同一组业务验收证明等价。[sqlc 配置](https://docs.sqlc.dev/en/latest/reference/config.html)、[查询示例](https://docs.sqlc.dev/en/latest/howto/select.html)

pgx 的 stdlib 是 `database/sql` 兼容层，modernc 的驱动不依赖 CGO。当前 `go.mod/go.sum` 固定 pgx v5.11.0、modernc SQLite v1.60.1、goose v3.28.0；mise 固定 Go 1.27.1 和 sqlc 1.30.0。[pgx stdlib](https://pkg.go.dev/github.com/jackc/pgx/v5/stdlib)、[modernc SQLite](https://pkg.go.dev/modernc.org/sqlite)

实际结构：

```text
internal/store/              # 统一业务类型、事务入口及方言映射
  sqlitequery/              # sqlc SQLite 生成包
  postgresquery/            # sqlc PostgreSQL 生成包
db/
  sqlite/migrations/        # goose 迁移，也是 sqlc schema 输入
  sqlite/queries/
  postgres/migrations/
  postgres/queries/
sqlc.yaml                   # 两组 engine，分别生成
```

业务服务接收 Store；sqlc 生成的行类型不直接暴露到 Huma API。API 的 Go DTO 仍是前后端契约来源。

## 配置与运行方式

运行时环境变量：

```dotenv
# 默认自托管部署
OCTOPULSE_DB_DRIVER=sqlite
OCTOPULSE_DB_DSN=file:/data/octopulse.db
```

```dotenv
# 使用已有 PostgreSQL
OCTOPULSE_DB_DRIVER=postgres
OCTOPULSE_DB_DSN=postgres://octopulse:example-password@db.example.com:5432/octopulse?sslmode=verify-full
```

凭据通过部署秘密配置注入，连接字符串不原样写入日志或管理 API；TLS 根证书等可配置。远程 PostgreSQL 建议验证服务器证书与主机名。[pgx 连接配置](https://github.com/jackc/pgx/blob/master/_autodocs/configuration.md)

SQLite 使用 WAL、`foreign_keys=ON`、`synchronous=FULL`、`busy_timeout=5000`，启动时检查实际 journal mode。参数通过驱动 DSN 初始化每个新物理连接，避免只在某个池连接执行一次 PRAGMA。[SQLite WAL](https://sqlite.org/wal.html)、[PRAGMA](https://sqlite.org/pragma.html)

SQLite 使用最多一个连接的写池和四个连接的只读池。写入口配置 immediate 事务，由 `sql.Tx` 管理提交与回滚；只读池为 `mode=ro`、`query_only=1`，不使用 immediate。WAL 支持读写并行，但同一数据库仍只能有一个写者。已测负载与限制见 [容量记录](capacity.md)，池大小不是任意环境的性能保证。[SQLite 事务](https://sqlite.org/lang_transaction.html)、[Go 连接池](https://go.dev/doc/database/manage-connections)

数据库文件放在本机持久卷，容器持久化整个数据目录。WAL 模式不能用于跨主机网络文件系统；运行中的 WAL/SHM 文件不能当作可随意删除的临时文件。及时关闭查询和事务，观察 WAL 大小与检查点进度。[WAL 约束](https://sqlite.org/wal.html)

启动时检查 `sqlite_version()`，当前实现要求至少 3.51.3，确保包含 WAL-reset bug 修复；不自动接纳其他旧分支的回补版本。[修复说明](https://sqlite.org/wal.html#walresetbug)

PostgreSQL 使用 `database/sql` 管理一个有界连接池，不叠加 pgxpool。`OCTOPULSE_DB_MAX_CONNECTIONS` 默认 10、范围 2–100，其中一个连接持有运行锁；最多保留 4 个空闲连接，连接生命周期 30 分钟。采用短事务、条件更新和事务内监控行 `FOR UPDATE` 处理竞争，普通事务使用默认 Read Committed；不能因此假定多条读取共享固定快照。[Go 连接池](https://go.dev/doc/database/manage-connections)、[PostgreSQL 隔离](https://www.postgresql.org/docs/current/transaction-iso.html)

## 表与数据类型

两库使用相同的逻辑表、关联、唯一约束与索引意图，DDL 按方言编写。当前表为 `documents`、`monitors`、`monitor_runtime`、`rounds`、`attempts`、`state_intervals`、`events`、`deliveries`、`page_slugs`、`page_domains`、`aggregates`、`watermarks`。配置资源和身份以带 kind/id 的文档保存，热路径使用普通列和关系约束：

| 数据组     | 内容                                                                      |
| ---------- | ------------------------------------------------------------------------- |
| 配置与身份 | 账号、会话、监控配置、秘密引用、通知渠道、Beszel 接入                     |
| 采集与状态 | 检查轮次、尝试明细、当前状态与确认计数、状态区间、运行代次/采集有效性记录 |
| 时间统计   | 维护/暂停区间、统计桶、聚合进度                                           |
| 通知       | 状态事件、逐渠道投递任务、投递尝试与租约                                  |
| 公开页面   | 草稿与发布配置、监控关联、分组排序、域名绑定、公告与进展、审计记录        |

应用生成 ID；统计用 UTC Unix 毫秒及整数时长，SQLite `INTEGER` 对应 PostgreSQL `BIGINT`。时区只用于输入解释和界面展示，跨窗口计算使用共用 Go 算法。ID、0/1 布尔、NULL 和错误在适配层映射为统一 Go 类型。

需要筛选、关联或索引的字段使用普通列，例如 monitor_id、类型、启用状态、间隔、due_at。类型专属设置、页面配置和资源文档以 JSON 序列化，两库当前均使用 TEXT；Go 负责相同的验证和业务解释，不依赖数据库专属 JSON 函数。修改文档、运行状态和通知标记仍可放在同一事务中。

索引覆盖监控项与时间、统计桶、通知状态与到期时间。规范化后的 slug/domain 设置唯一约束；顺序查询明确第二排序键，避免结果依赖不同数据库的默认排序。

## 事务与并发

HTTP/DNS/TCP 探测、Shoutrrr 发送及 Beszel 网络请求都在数据库事务外。Go 使用事务对象执行事务内操作，不能混入池上的非事务调用。[Go 事务](https://go.dev/doc/database/execute-transactions)

一轮探测完成后，在一个短事务中校验配置版本和运行代次，写入轮次/尝试，更新确认计数和当前状态，关闭/开启状态区间，创建状态事件及逐渠道通知任务。全部提交或全部回滚；稳定 round_id 和事件/渠道唯一约束使重复提交幂等。编辑、暂停或删除后，过期探测不能覆盖新状态。

被动心跳上报/超期判定同样原子保存最近上报、确认状态、区间和通知任务；超期判断须校验最近上报版本，避免与成功上报竞争后错误改为 Down。证书检查独立保存证书四态、阈值/续期事件及通知，不写入服务 uptime 状态区间。

投递 worker 通过条件 UPDATE 认领到期任务，保存 lease_token/lease_until。发送在事务外，完成或退避更新必须匹配同一 token；崩溃后的过期租约可重新认领。发送前检查事件版本、故障周期和维护规则，取消失效任务。外部发送成功但成功记录尚未提交时崩溃，仍可能重复发送，因此不承诺恰好一次或所有渠道一定送达。

投递 worker 默认并发为 4，单次发送超时为 30 秒；每项任务最多尝试 8 次（含首次），失败后从 5 秒开始指数退避，最长间隔为 15 分钟，达到尝试上限后记录为 `failed`。具体参数见 [投递 worker](../internal/notify/worker.go)。

成功确认事务同时保存该渠道的故障周期投递标记；恢复任务及维护退出后的补发按此标记判断。标记、任务完成和已补发记录原子保存，避免重启后丢失通知顺序依据或重复创建恢复任务。

数据库写入失败属于平台采集异常，不能直接把目标服务改为 Down。观察持久化失败、池等待、锁等待和队列积压；重试数据库事务不重新执行外部探测。遇到提交结果不确定时，先按稳定 ID 核对已落库结果。

状态区间是可用率依据；共用算法计算窗口交集和维护/暂停并集。统计桶按确定键幂等更新，桶与聚合水位一起提交，只分批删除已完成必要聚合的数据，删除范围不越过聚合水位。默认每秒扫描调度并更新采集有效性，异常重启从最后可靠水位切出 Unknown；统计中的已知开放区间也裁到可靠采集边界，不能把最后一次 Up 延续到停机区间。维护变更按来源哈希回退受影响桶并重算，已超过原始轮次保留期的延迟计数保留；默认每分钟聚合/清理，周期可配置。

100 项每 30 秒检查，全天约 28.8 万轮，14 天约 403.2 万轮；这只是输入规模估算。当前实际容量测试同时覆盖重试、公开/后台查询、聚合及历史清理，但只填入代表性历史样本，不能把结果推导为已验证 403.2 万轮完整历史规模，详见 [capacity.md](capacity.md)。

## Schema 升级、备份与验证

每次 Schema 变更同时提供两种方言的同业务版本迁移，嵌入二进制；当前 SchemaVersion 为 2。取得运行锁和连接检查后执行迁移，成功后才启动 API/探测/投递，失败退出。版本 2 为聚合增加成功轮次计数，并从仍保留的原始轮次补齐，升级保留行为有专门用例。goose 用版本表跟踪迁移，但不自动验证两库语义等价。[goose Provider](https://pkg.go.dev/github.com/pressly/goose/v3#NewProvider)

本实现使用覆盖迁移和运行期的 PostgreSQL advisory lock / SQLite 文件锁，不依赖 goose 默认锁。迁移原子边界是单个文件，不是整个升级批次；当前迁移未使用 `NO TRANSACTION`。已发布文件不修改，以新增迁移修正；数据库版本高于程序支持版本时拒绝启动。[迁移事务](https://pressly.github.io/goose/documentation/annotations/#no-transaction)

SQLite CLI `octopulse backup DESTINATION` 使用 `VACUUM INTO` 创建一致备份，输出权限 0600。该 CLI 同样取得运行锁，须先停止服务；不能在运行时只复制 `.db`。PostgreSQL 使用 pg_dump/pg_restore 或部署方已有备份体系，同类型恢复。AES-256-GCM 密钥与上传图片也需要备份，数据库恢复后必须沿用原密钥；具体二进制/Compose 命令见 [运维手册](operations.md#备份与恢复)。[PostgreSQL Dump](https://www.postgresql.org/docs/current/backup-dump.html)

首版双数据库支持包括分别建库、升级及恢复，不自动包含 SQLite→PostgreSQL 的在线切换。以后若增加搬迁工具，应通过停机导出/导入统一模型、保留 ID/关联并校验统计与未完成任务，而不是直接导入 SQLite SQL dump。

同一套业务集成测试在临时文件 SQLite 和真实 PostgreSQL18.6 上运行，完整竞态套件已通过，具体记录见 [验收映射](acceptance.md)。测试使用专用数据库并创建独立 schema，跨包运行要加 `-p 1` 避免运行锁竞争。测试覆盖：

- 事务中途失败不留下半个轮次、状态变化或通知；重复提交不创建重复事件/任务。
- 确认阈值、维护/暂停并集、跨窗口、Unknown、零分母和覆盖率结果一致。
- 编辑/暂停期间的旧探测不会覆盖新配置；并发写入和连接重建正确执行外键/唯一约束。
- 租约过期恢复、旧 token 回写被拒、过时通知取消和外部发送后崩溃符合投递规则。
- 聚合重跑、清理并发和异常重启不重复统计或延续错误状态。
- 空库迁移、旧版本升级、备份恢复及恢复后继续投递/统计分别通过。
- 100 项/30 秒与公开查询、聚合、清理同时运行，记录 API p95、最长轮次、计划迟到、连接池和投递积压、goroutine 与 Go 内存采样；当前记录不包含磁盘 I/O 或进程 RSS 测量。
