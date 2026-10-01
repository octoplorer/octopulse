# SQLite 与 PostgreSQL 持久层方案

状态：2026-10-02，根据用户追问细化的工程建议。双数据库支持已确认；本文的驱动、生成器、连接池及目录安排尚未实现，也不代表性能已经验证。

## 共享业务模型与方言边界

一个部署只连接一个数据库，由启动配置选择 `sqlite` 或 `postgres`。同一 Go 二进制包含两种驱动，默认建议 SQLite；切换连接配置只会打开另一个数据库，不会自动迁移已有数据。

首版仍只运行一个 Go 服务实例。建议加入运行期间的启动互斥：SQLite 使用数据目录/数据库的进程锁，PostgreSQL 使用独立会话持有应用 advisory lock，或等效的单实例保护。迁移锁不代替运行锁；运行锁失去后停止调度并退出，不将 PostgreSQL 支持解释为多副本调度支持。

监控状态机、阈值、维护规则、时间统计、通知策略和权限只实现一次。持久层提供统一的业务数据类型和事务入口，两种适配器处理 SQL、列类型及错误转换，不把数据库类型判断散落在业务服务中。

| 层次 | 建议选择 | 作用 |
| --- | --- | --- |
| 连接与事务 | Go `database/sql` | 共用连接池、context 和事务编程接口 |
| SQLite 驱动 | `modernc.org/sqlite` | CGO-free，便于交付 Go 二进制 |
| PostgreSQL 驱动 | `github.com/jackc/pgx/v5/stdlib` | 将 pgx 接入 `database/sql` |
| 查询代码 | sqlc | 分别校验和生成两种方言的带类型查询 |
| Schema 迁移 | goose | 按数据库版本执行嵌入二进制的迁移 |

sqlc 支持在同一配置中设置多个 engine/schema/query/output，Go 可使用 `database/sql`。它不承担把任意 PostgreSQL SQL 翻译成 SQLite SQL 的职责；项目维护两套方言查询，使用同一组业务验收证明等价。[sqlc 配置](https://docs.sqlc.dev/en/latest/reference/config.html)、[查询示例](https://docs.sqlc.dev/en/latest/howto/select.html)

pgx 的 stdlib 是 `database/sql` 兼容层，modernc 的驱动不依赖 CGO。具体版本在实现时核实、固定并记录，迁移和生成工具由 mise 固定。[pgx stdlib](https://pkg.go.dev/github.com/jackc/pgx/v5/stdlib)、[modernc SQLite](https://pkg.go.dev/modernc.org/sqlite)

拟采用的结构：

```text
internal/store/              # 统一业务类型、事务入口
  sqlite/                   # SQLite 适配器及 sqlc 生成包
  postgres/                 # PostgreSQL 适配器及 sqlc 生成包
db/
  sqlite/migrations/        # goose 迁移，也是 sqlc schema 输入
  sqlite/queries/
  postgres/migrations/
  postgres/queries/
sqlc.yaml                   # 两组 engine，分别生成
```

业务服务接收 Store；sqlc 生成的行类型不直接暴露到 Huma API。API 的 Go DTO 仍是前后端契约来源。

## 配置与运行方式

以下环境变量名称为拟定接口：

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

SQLite 建议开启 WAL，设置 `foreign_keys=ON`、`synchronous=FULL`、`busy_timeout=5000`。启动时检查实际 journal mode。外键、同步等级和等待时间等连接设置必须作用于每个新物理连接，不能只通过池中的某个连接执行一次 PRAGMA。[SQLite WAL](https://sqlite.org/wal.html)、[PRAGMA](https://sqlite.org/pragma.html)、[驱动连接初始化](https://pkg.go.dev/modernc.org/sqlite#RegisterConnectionHook)

SQLite 建议使用最多一个连接的写池，以及初始上限为四个连接的小型只读池。写入口使用驱动的 immediate 事务配置，仍由 `sql.Tx` 管理提交与回滚；只读池不使用 immediate。WAL 支持读写并行，但同一数据库仍只能有一个写者。池大小和等待时间需要压测调整；写入、查询和清理都设置 context 截止时间。[SQLite 事务](https://sqlite.org/lang_transaction.html)、[Go 连接池](https://go.dev/doc/database/manage-connections)

数据库文件放在本机持久卷，容器持久化整个数据目录。WAL 模式不能用于跨主机网络文件系统；运行中的 WAL/SHM 文件不能当作可随意删除的临时文件。及时关闭查询和事务，观察 WAL 大小与检查点进度。[WAL 约束](https://sqlite.org/wal.html)

实现时还需检查驱动内嵌的 `sqlite_version()` 已包含 WAL-reset bug 修复；官方说明 SQLite 3.51.3 及之后已修复，旧分支也有指定回补版本。[修复说明](https://sqlite.org/wal.html#walresetbug)

PostgreSQL 使用 `database/sql` 管理一个可配置的有界连接池，初始建议最多十个连接，不再叠加另一套 pgxpool。采用短事务、行版本条件更新及必要锁定处理竞争；普通事务可从 Read Committed 起步，不能因其隔离级别名称而假定多条读取共享固定快照。[Go 连接池](https://go.dev/doc/database/manage-connections)、[PostgreSQL 隔离](https://www.postgresql.org/docs/current/transaction-iso.html)

## 表与数据类型

两库使用相同的逻辑表、关联、唯一约束与索引意图，具体 DDL 按方言编写。初步分组如下，表名在实现时固定：

| 数据组 | 内容 |
| --- | --- |
| 配置与身份 | 账号、会话、监控配置、秘密引用、通知渠道、Beszel 接入 |
| 采集与状态 | 检查轮次、尝试明细、当前状态与确认计数、状态区间、运行代次/采集有效性记录 |
| 时间统计 | 维护/暂停区间、统计桶、聚合进度 |
| 通知 | 状态事件、逐渠道投递任务、投递尝试与租约 |
| 公开页面 | 草稿与发布配置、监控关联、分组排序、域名绑定、公告与进展、审计记录 |

建议应用生成 ID；统计用 UTC Unix 毫秒及整数时长，SQLite `INTEGER` 对应 PostgreSQL `BIGINT`。时区只用于输入解释和界面展示，跨窗口计算使用共用 Go 算法。ID、布尔、NULL 和错误均在适配层映射为统一 Go 类型。

需要筛选、关联或索引的字段使用普通列，例如 monitor_id、类型、启用状态、间隔、due_at。类型专属复杂设置和页面品牌配置可存为带版本 JSON，SQLite 用 TEXT，PostgreSQL 可用 JSONB；Go 负责相同的验证和业务解释，首版不让数据库专属 JSON 函数决定产品行为。

索引覆盖监控项与时间、统计桶、通知状态与到期时间。规范化后的 slug/domain 设置唯一约束；顺序查询明确第二排序键，避免结果依赖不同数据库的默认排序。

## 事务与并发

HTTP/DNS/TCP 探测、Shoutrrr 发送及 Beszel 网络请求都在数据库事务外。Go 使用事务对象执行事务内操作，不能混入池上的非事务调用。[Go 事务](https://go.dev/doc/database/execute-transactions)

一轮探测完成后，在一个短事务中校验配置版本和运行代次，写入轮次/尝试，更新确认计数和当前状态，关闭/开启状态区间，创建状态事件及逐渠道通知任务。全部提交或全部回滚；稳定 round_id 和事件/渠道唯一约束使重复提交幂等。编辑、暂停或删除后，过期探测不能覆盖新状态。

被动心跳上报/超期判定同样原子保存最近上报、确认状态、区间和通知任务；超期判断须校验最近上报版本，避免与成功上报竞争后错误改为 Down。证书检查独立保存证书四态、阈值/续期事件及通知，不写入服务 uptime 状态区间。

投递 worker 通过条件 UPDATE 认领到期任务，保存 lease_token/lease_until。发送在事务外，完成或退避更新必须匹配同一 token；崩溃后的过期租约可重新认领。发送前检查事件版本、故障周期和维护规则，取消失效任务。外部发送成功但成功记录尚未提交时崩溃，仍可能重复发送，因此不承诺恰好一次或所有渠道一定送达。

成功确认事务同时保存该渠道的故障周期投递标记；恢复任务及维护退出后的补发按此标记判断。标记、任务完成和已补发记录原子保存，避免重启后丢失通知顺序依据或重复创建恢复任务。

数据库写入失败属于平台采集异常，不能直接把目标服务改为 Down。观察持久化失败、池等待、锁等待和队列积压；重试数据库事务不重新执行外部探测。遇到提交结果不确定时，先按稳定 ID 核对已落库结果。

状态区间是可用率依据；共用算法计算窗口交集和维护/暂停并集。统计桶按确定键幂等更新，桶与聚合水位一起提交，只分批删除已完成必要聚合的数据，删除范围不能越过聚合水位。异常重启从最后可靠采集水位切出 Unknown，不能把最后一次 Up 延续到重启后；水位周期与保守误差在实现时明确。聚合不越过已确认的采集有效性边界；若重启修正或维护/暂停记录变更影响已有桶，则失效并重算受影响范围。

100 项每 30 秒检查，全天约 28.8 万轮，14 天约 403.2 万轮；这只是输入规模估算。需要同时验证重试、公开查询、聚合及历史清理的负载，不能仅凭平均每秒 3.33 轮宣称性能达标。

## Schema 升级、备份与验证

每次 Schema 变更同时提供两种方言的同业务版本迁移文件，嵌入二进制。启动顺序为连接检查、取得迁移互斥、执行迁移，成功后才启动 API/探测/投递；失败退出。goose 默认使用版本表跟踪迁移，支持选择不同 dialect 和嵌入目录，但不自动验证两种 Schema 语义等价。[goose Provider](https://pkg.go.dev/github.com/pressly/goose/v3#NewProvider)

迁移互斥须显式配置：PostgreSQL 可用 advisory lock，SQLite 部署执行单实例/进程锁。goose 默认没有开启迁移锁。默认原子边界是单个迁移文件，不是整个升级批次；包含 `NO TRANSACTION` 的文件须单独设计失败恢复。已发布的迁移文件不修改，以新增迁移修正；数据库版本高于程序支持版本时拒绝启动。[goose 锁](https://pkg.go.dev/github.com/pressly/goose/v3#WithSessionLocker)、[迁移事务](https://pressly.github.io/goose/documentation/annotations/#no-transaction)

SQLite 使用一致性在线备份或停机、正常关闭后的备份，不在运行时仅复制 `.db`。PostgreSQL 使用 pg_dump/pg_restore 或部署方已有备份体系；原生数据库备份分别用于同类型恢复。页面图片等文件资源也纳入部署备份，若凭据使用应用层加密，密钥另行安全保存。[SQLite Backup](https://sqlite.org/backup.html)、[PostgreSQL Dump](https://www.postgresql.org/docs/current/backup-dump.html)

首版双数据库支持包括分别建库、升级及恢复，不自动包含 SQLite→PostgreSQL 的在线切换。以后若增加搬迁工具，应通过停机导出/导入统一模型、保留 ID/关联并校验统计与未完成任务，而不是直接导入 SQLite SQL dump。

同一套业务集成测试在临时文件 SQLite 和真实 PostgreSQL 上运行：

- 事务中途失败不留下半个轮次、状态变化或通知；重复提交不创建重复事件/任务。
- 确认阈值、维护/暂停并集、跨窗口、Unknown、零分母和覆盖率结果一致。
- 编辑/暂停期间的旧探测不会覆盖新配置；并发写入和连接重建正确执行外键/唯一约束。
- 租约过期恢复、旧 token 回写被拒、过时通知取消和外部发送后崩溃符合投递规则。
- 聚合重跑、清理并发和异常重启不重复统计或延续错误状态。
- 空库迁移、旧版本升级、备份恢复及恢复后继续投递/统计分别通过。
- 100 项/30 秒与公开查询、聚合、清理同时运行，测量持久化延迟、锁/池等待、磁盘与积压。
