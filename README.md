# Octopulse

Octopulse 是单组织自托管的 uptime 监控平台。一个 Go 进程负责探测、调度、统计、通知、API 和静态资源；Vue CSR SPA 提供管理后台和可配置公开状态页。运行数据库可选择 SQLite 或 PostgreSQL。

- HTTP(S)、TCP、DNS、被动心跳、证书到期监控，含请求、认证、TLS、代理、断言和重试配置。
- 按确认状态持续时间计算可用率和覆盖率，独立处理 Unknown、暂停、维护和采集断档。
- Shoutrrr 通知渠道、持久投递记录、有限退避重试、故障与恢复顺序。
- 多个状态页的品牌、别名、分组、排序、公告、维护、草稿预览和显式发布；支持路径与独立域名。
- 内置管理员、操作员和只读账号；凭据加密保存，公开接口使用独立字段白名单。
- 独立 Beszel 服务器页，展示摘要、历史和容器；服务器指标不参与 uptime 判定。

首版采用单实例运行，不提供分布式探测。GitHub 主题不在首版范围。

## 快速启动

使用 Docker Compose，在仓库根目录执行：

```sh
docker compose up --build -d
docker compose logs -f octopulse
```

打开 `http://127.0.0.1:8080/app`，首次访问创建管理员，之后由管理员添加其他成员。密码要求为 12–72 字节，公开注册关闭。

默认使用 SQLite，数据库、加密密钥和上传图片保存在 `octopulse-data` 卷；应用端口仅发布到宿主机回环地址。PostgreSQL、二进制部署、域名与备份流程见 [运维手册](docs/operations.md)。本地开发的工具安装和启动步骤见 [开发指南](docs/development.md#本地开发)。

本地开发先安装 mise，然后在仓库根目录执行：

```sh
mise trust
mise install
mise run install
mise run dev
```

根目录的私有 aube workspace 管理 Node 命令、共享工具和 ESLint 配置；`web/` 保留前端代码、运行时依赖、专用工具及构建配置。mise 使用 `mise.toml` 和 `mise.lock` 固定工具链，统一编排依赖安装、开发、检查、测试、构建和契约生成；根 `package.json` 的同名脚本仅处理前端。Node.js 只参与开发与构建，生产仍由 Go 提供页面和 API。

## 文档导航

| 内容                                   | 文档                                                                                     |
| -------------------------------------- | ---------------------------------------------------------------------------------------- |
| 开发环境、前端约定、契约生成与测试     | [开发指南](docs/development.md)                                                          |
| 构建交付、配置、数据库部署、域名与备份 | [运维手册](docs/operations.md)                                                           |
| 产品边界、权限及业务要求               | [首版规格](docs/v1-spec.md)                                                              |
| 各类监控设置与限值                     | [监控配置](docs/monitor-options.md)                                                      |
| 状态页发布、汇总及公开字段             | [状态页规则](docs/status-page-policy.md)                                                 |
| 数据持久化、事务、统计与迁移           | [数据库设计](docs/database-design.md)                                                    |
| 实现差距、验证记录与容量边界           | [验收映射](docs/acceptance.md)、[容量记录](docs/capacity.md)                             |
| 领域术语与架构决策背景                 | [术语表](GLOSSARY.md)、[设计访谈与 ADR 索引](docs/design-tree.md#已记录的术语与架构决策) |
| Beszel 协议、兼容版本与缓存策略        | [适配器说明](internal/beszel/README.md)                                                  |
