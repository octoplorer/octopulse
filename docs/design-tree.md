# Octopulse 设计访谈

日期：2026-10-01。状态：Q1–Q26 已回答并记录，最新范围调整移除首版 GitHub 主题；原 Q27 主题运行问题失效。产品分支已收敛，整体规格待最终确认，尚未开始实现。

本文件记录用户已给出的要求、设计决策之间的依赖和待回答问题。已确定的领域术语写入根目录 `GLOSSARY.md`，原 `CONTEXT.md` 保留链接入口；有真实取舍且难以逆转的已确定决策写入 `docs/adr/`。

## 用户已给出的要求

- 使用 mise 管理多语言开发环境。
- 使用 Go 和 Node.js；npm 包管理器使用 aube，已核实为 [aubepkg/aube](https://github.com/aubepkg/aube)。
- 数据库支持 SQLite 和 PostgreSQL。
- 前端使用 Vite + Vue，纯 CSR SPA；引入 Pinia Colada、Vue Router、VueUse。
- 样式使用 UnoCSS、preset-wind4、preset-attributify，尽可能采用属性模式。
- 前端组件使用 @ark-ui/vue。
- 使用 HeyAPI 集成前后端契约。
- 监控项提供全面设置，包括显示名称、监控类型、检查间隔、重试次数、请求方法、请求编码和高级选项。
- 通知使用用户指定的 nicholas-fedor/shoutrrr。
- 状态页支持充分的内置自定义；用户最新决定从首版移除以 GitHub 仓库提供完整主题的功能。
- 集成 henrygd/beszel，额外显示服务器监控信息。
- 未获用户要求前，不提交 Git commit，不推送远端；如之后获授权，commit 和 PR 标题使用 Conventional 风格。

以上是要求记录，不代表已解决其具体语义和架构取舍。

## 第 1 轮：根节点

用户已在本轮作出如下决定。

| 编号 | 问题 | 已确认决定 |
| --- | --- | --- |
| Q1 | 产品面向单组织自托管，还是多租户托管服务？ | 单组织自托管；Q6 已确认多人管理 |
| Q2 | Go 与 Node.js 分别承担哪些运行职责？ | Go 承担常驻服务；Node.js 用于前端和契约构建，原主题构建职责随最新范围调整撤销 |
| Q3 | 首版目标监控数量和最小检查间隔是多少？ | 100 个监控项、最小 30 秒；这是设计与压测目标，尚未验证 |
| Q4 | 首版要一次覆盖哪些监控类型与功能？ | 首个可用版本覆盖 HTTP(S)、TCP、DNS、被动心跳、证书到期，以及通知、内置状态页自定义、Beszel 展示 |
| Q5 | GitHub 主题的自由度是否包含任意浏览器脚本？ | 原已讨论完整界面替换；最新决定从首版撤销整个 GitHub 主题分支，采用内置可配置状态页，见 ADR 0011 |

## 第 2 轮：产品与监控语义

用户已在本轮作出如下决定。

| 编号 | 问题 | 已确认决定 |
| --- | --- | --- |
| Q6 | 单管理员，还是多人账号及权限？ | 多人账号；管理员、操作员、只读三种角色，Q21 已确认具体边界 |
| Q7 | 首版是否需要多个运行实例或远程探测节点？ | 不需要分布式探测；Q11 已确认单 Go 实例及重启调度策略 |
| Q8 | 主动可用性检查的轮次内重试、连续失败阈值和恢复阈值是否独立？ | 独立配置；默认追加重试 2 次、连续失败轮次阈值 1、连续成功轮次恢复阈值 1；被动心跳与证书告警另定语义 |
| Q9 | 状态页的发布边界是什么？ | 首版多页面、匿名公开、独立选择及分组监控项；支持人工事件公告和计划维护；受保护页面后续支持 |
| Q10 | Beszel 指标展示位置及对服务状态的影响是什么？ | 网站与服务器指标无关；Beszel 作为独立的数据展示接入。Q18 已确认后台独立服务器页及首版不公开发布 |

本轮确定产品模型；后续已收敛状态页公共契约、通知触发、数据库与运维、类型专属设置、数据留存和前端范围。原主题安装分支随最新范围调整撤销。

## 第 3 轮：时间、统计与交付

用户已在本轮接受全部建议。

| 编号 | 问题 | 已确认决定 |
| --- | --- | --- |
| Q11 | 检查耗时、重叠执行及进程重启如何处理？ | 首版单 Go 运行实例；固定计划间隔，同项不重叠；每轮总时间预算不超过间隔，重试次数为上限；重启不补跑历史检查，立即开始新轮次，采集断档单独标记 |
| Q12 | 可用率如何计算，维护与暂停如何影响统计和探测？ | 按确认后的 Up/Down 持续时间计算，不回溯确认时间；Unknown、暂停、计划维护排除分母并展示覆盖率；维护继续探测且抑制故障/恢复通知，暂停停止探测；平台漏检不能冒充目标故障 |
| Q13 | 被动心跳的缺失、首次上报和显式失败如何判断？ | 每项独立密钥；预期周期+宽限时间；首次等待为 Unknown、过期为 Down；成功上报为 Up，允许主动上报失败；具体 API 与限值后续确定 |
| Q14 | 证书到期如何展示、提醒及参与可用率？ | 独立证书状态：有效、即将到期、已过期、检查失败；默认到期提醒阈值 30/14/7/1 天，默认每日检查；不并入服务可用率 |
| Q15 | 通知需要怎样的重启恢复与投递保证？ | 数据库持久任务、按渠道记录结果和有限退避重试；接受极端场景重复投递；抑制已失去时效的故障通知；Q22 已确认触发与维护退出；投递参数及结果查询由工程细化 |
| Q16 | GitHub 主题在哪里构建、如何安装？ | 原已讨论预构建主题包；最新决定随整个 GitHub 主题功能移出首版，不安装或执行外部主题产物 |
| Q17 | OpenAPI 规范还是 Go API 定义作为契约来源？ | Go 类型/路由作为唯一源，通过 Huma 导出 OpenAPI，再由 HeyAPI 生成 TypeScript、SDK 和 Pinia Colada 集成；生成与验证流程后续确定 |
| Q18 | Beszel 独立展示首版包含哪些能力、展示在哪里？ | 管理后台独立服务器页，服务器摘要、历史图表、容器列表；首版一个 Hub，公开状态页暂不发布服务器指标 |

Q11–Q18 已回答，Q16 随最新范围调整撤销。当前可用率、调度、投递和页面配置的工程边界进一步明确如下。

### 由已确认决定推导的工程规则

- 每轮预算覆盖尝试、重试等待与收尾；默认重试 2 次是上限，预算不足时允许更少尝试，并记录实际尝试数和终止原因。
- 同一监控项不重叠执行；平台停机不补跑历史探测，也不能延长最后一次 Up 来填满整个停机区间。
- 可用率为 `Up / (Up + Down)`；采集覆盖率为 `(Up + Down) / (窗口时长 - 暂停及维护的并集时长)`。Unknown 降低覆盖率；可用率分母为零时展示暂无统计。
- 暂停与多个维护窗口按时间并集处理，避免重复扣除统计时长；维护继续采集，暂停停止采集。
- 被动心跳具备可轮换的独立密钥；密钥和鉴权配置不进入公开状态页契约。
- 证书四态独立于服务可用率；Q23 已确认风险独立展示，不参与服务总状态自动汇总。
- 状态页自定义通过结构化配置实现，各页独立作用域；预览与发布复用内置组件，公共接口保持字段白名单。
- SQLite 与 PostgreSQL 保持相同领域语义并验证关键流程；双数据库支持不自动包含在线跨数据库搬迁。
- Beszel 标记采集时间、来源与过期信息，不重复实现其监控告警，也不默认复制全部历史数据。

## 第 4 轮：设置、权限与发布边界

用户已回答本轮：Q19/Q21/Q22/Q23/Q25/Q26 接受建议，Q20/Q24 依其修正执行。

| 编号 | 问题 | 已确认决定 |
| --- | --- | --- |
| Q19 | 全面监控设置的首版范围 | 按 [配置清单](./monitor-options.md) 提供类型专属设置；请求格式、字符编码与传输压缩分别配置；复杂自动认证及额外协议后续支持 |
| Q20 | HTTP 请求是否按方法采用不同追加重试策略 | 所有方法采用同一配置和默认值（追加重试 2 次），不从 method 推断幂等性、不自动改变次数；配置者负责选择可重复探测的请求 |
| Q21 | 首版登录及三角色权限矩阵 | 内置账号、关闭公开注册；管理员管理账号、秘密、渠道、Beszel 及系统设置；操作员管理监控、页面内容及自定义、公告和维护，使用已有渠道；只读查看脱敏信息；OIDC 后续支持。主题权限随最新范围调整移除 |
| Q22 | 通知触发、持续提醒和维护退出 | 确认故障与恢复各通知一次，初始 Up 不通知，持续提醒可配置且默认关闭；维护退出完成一次新的有效评估（主动检查新轮次、心跳重新评估期限），仍故障则通知，若此前已发故障且在维护内恢复则补发恢复；Unknown 不补发故障/恢复；证书阈值与续期通知独立配置 |
| Q23 | 状态页总体状态与基本自定义 | 按 [发布与汇总规则](./status-page-policy.md) 汇总：全部 Down 为全面故障、部分 Down 为部分故障、无 Down 但有 Unknown 为数据不足、其次维护、其次全 Up 正常；空集合/全暂停为数据不足；公告只能显式提升影响等级，证书风险独立；内置页面提供品牌、内容与分组配置 |
| Q24 | 状态页自定义与路径/域名发布 | 自定义只影响所发布状态页，管理后台不受影响；每页首版支持默认 `example.com/status1` 及可选独立域名 `status1.example.com/`。公共字段白名单保留，直接使用内置 Vue CSR 页面；主题包及 iframe 方案撤销 |
| Q25 | 默认历史留存与诊断保存范围 | 原始轮次 14 天、尝试明细 3 天、5 分钟聚合 90 天、小时聚合及状态区间 13 个月，可配置；不默认保存完整请求/响应正文与秘密头 |
| Q26 | 前端语言、时区和浏览器支持 | 简体中文/英文、浅色/深色、组织时区与个人显示时区、桌面优先响应式、现代浏览器；不承担旧浏览器兼容 |

每页默认路径与可选独立域名均为首版要求，最新范围调整只移除 GitHub 主题，不移除发布入口或页面自定义。两种入口复用同一内置页面组件和已发布配置；不引入 iframe 或主题消息桥。

拟采用的工程默认方案是支持内网目标、自定义代理、二进制与 Docker 交付，以及由运维配置 DNS/HTTPS/反代；这些不是新增的用户确认结果，均在 [首版规格](./v1-spec.md) 中明示供最终评审。现有 Hub 特殊认证与具体工具版本作为工程核实事项，不再增加常规选型问卷。

## 最新范围调整与整体定稿

用户决定首版移除 GitHub 主题、集中做好内置状态页自定义；[ADR 0011](./adr/0011-built-in-configurable-status-pages.md) 已接受。原 Q27 的 iframe 运行问题失效，外部主题安装、任意脚本、产物兼容和消息桥不进入首版。

没有剩余产品决策问卷。当前 [首版规格与验收](./v1-spec.md) 已按最新决定修订，整体定稿仍待用户确认；确认后开始实现。后续事实核实、版本、限值和工程组织按规格自主完成，不再要求接受 iframe。

## 已记录的术语与架构决策

- [领域术语](../GLOSSARY.md)
- [ADR 0001：单组织自托管，Go 承担常驻服务](./adr/0001-single-organization-go-service.md)
- [ADR 0002：GitHub 主题完整定义状态页界面（历史，已被取代）](./adr/0002-versioned-status-page-themes.md)
- [ADR 0003：分开轮次内重试与故障、恢复确认阈值](./adr/0003-check-rounds-and-state-thresholds.md)
- [ADR 0004：Beszel 接入作为独立的数据展示](./adr/0004-independent-server-metrics.md)
- [ADR 0005：可用率按确认状态时长计算，并展示采集覆盖率](./adr/0005-duration-availability-and-coverage.md)
- [ADR 0006：通知由平台管理持久投递任务](./adr/0006-durable-notification-delivery.md)
- [ADR 0007：安装预构建的状态页主题包（历史，已被取代）](./adr/0007-prebuilt-theme-packages.md)
- [ADR 0008：Go API 定义作为契约来源](./adr/0008-go-source-openapi-contract.md)
- [ADR 0009：每个状态页同时支持路径入口和独立域名](./adr/0009-status-page-paths-and-domains.md)
- [ADR 0010：用公共页面外壳承载隔离的完整主题（未接受，已撤回）](./adr/0010-sandboxed-public-themes.md)
- [ADR 0011：首版采用内置可配置状态页](./adr/0011-built-in-configurable-status-pages.md)
- [首版规格与验收](./v1-spec.md)

## 依赖树

- Q1 产品与用户边界
  - 身份认证、角色、资源归属、审计与密钥访问
  - 公共/私有状态页与订阅者
  - 出站探测的授权范围
- Q2 运行职责 + Q3 容量目标
  - 进程与部署形式、检查调度、重启恢复、并发控制
  - 单节点/多节点探测、选主与去重
  - SQLite/PostgreSQL 语义一致性、迁移与备份
  - 实时数据刷新、历史保留与聚合
- Q4 功能范围
  - 监控配置：类型专属选项、请求编码、认证、TLS、代理、断言
  - 领域语义：检查、尝试、故障、恢复、暂停、维护、未知状态
  - 重试与调度：间隔起点、超时、抖动、重叠检查
  - 可用率：分母、缺失数据、维护时间、跨窗口事件
  - 通知：触发、恢复、抑制、升级、重发、失败处理
  - Beszel：展示范围、凭据、兼容版本、刷新与失联表现
- Q23/Q24 内置状态页自定义与发布
  - 每页品牌、内容、别名、分组与排序
  - 已发布公共数据、公开接口与管理权限边界
  - 自定义预览、保存与发布，两种访问入口
- 原 Q5/Q16 GitHub 主题分支
  - 已移出首版，相关历史 ADR 保留但不参与实现和验收
- 各根节点确定后
  - OpenAPI 契约来源与 HeyAPI 生成流程
  - CSR 状态页、路由、分享元信息和搜索可见性
  - 前端组件、属性样式、浏览器支持、无障碍与国际化
  - 开发工作流、发布、验收和运维文档

资料调查中的事实是尚未完成的前置条件；依赖这些事实的问题留待后续轮次，不让用户猜测事实。

## 已核实的资料约束

- aube 是用户所指的包管理器项目，后续按其自身能力设计工具链。[官方仓库](https://github.com/aubepkg/aube)
- HeyAPI 已提供 Pinia Colada 集成，当前文档标明支持 Pinia Colada v0；生成流程仍需由契约来源和兼容版本决定。[官方文档](https://heyapi.dev/docs/openapi/typescript/plugins/pinia-colada)
- HeyAPI 消费 OpenAPI 输入，不替后端决定契约来源；Q17 已选择由 Go/Huma 导出规范。[输入文档](https://heyapi.dev/docs/openapi/typescript/configuration/input)
- mise 的工具版本锁文件与 aube 的依赖锁文件承担不同职责。aube 新项目默认采用 `aube-lock.yaml`，冻结安装与依赖安装脚本信任策略需要在工程初始化时验证。[mise lockfile](https://mise.jdx.dev/dev-tools/mise-lock.html)、[aube lockfiles](https://aube.sh/package-manager/lockfiles)
- Beszel 官方提醒 REST API 的数据结构和内容可能在次版本中变化，集成需要明确兼容版本及升级策略。[REST API 文档](https://beszel.dev/guide/rest-api)
- Beszel 支持普通 readonly 用户；首版独立后台展示已由 Q18 确认，专用身份与凭据管理作为工程兼容验证事项。[用户账号文档](https://beszel.dev/guide/user-accounts)
- UnoCSS Attributify 官方建议使用 `un-` 前缀避免与组件属性冲突；属性模式的具体规范留待前端设计阶段确定。[Attributify 文档](https://unocss.dev/presets/attributify)
- Ark UI 提供 headless 组件，因此平台仍需建立统一的视觉规范。[样式文档](https://ark-ui.com/docs/guides/styling)
- Vue 多根组件不会自动透传样式属性；属性模式需要明确组件的 DOM 落点。[属性透传](https://vuejs.org/guide/components/attrs)
- Wind4 使用现代 CSS 能力，不支持 presetLegacyCompat；Q26 已选择现代浏览器范围，精确版本底线在实现时核实。[Wind4 文档](https://unocss.dev/presets/wind4)
- Vue Router 的 HTML5 history 需要服务端回退；API、静态资源与未知路由的行为须明确。[history mode](https://router.vuejs.org/guide/essentials/history-mode)
- Shoutrrr 的 Router 内置队列是内存集合，Flush 发送后清空且不保留发送错误；它不提供持久投递任务或统一自动重试。渠道内部重试能力并不一致；平台持久任务及投递规则已由 Q15/Q22 确认。核验时固定源码为 `afb6e75af1260ca122d902e239390112bee7f195`，这不代表已经选择该版本作为依赖。[队列源码](https://github.com/nicholas-fedor/shoutrrr/blob/afb6e75af1260ca122d902e239390112bee7f195/pkg/router/router.go#L100-L146)、[发送源码](https://github.com/nicholas-fedor/shoutrrr/blob/afb6e75af1260ca122d902e239390112bee7f195/pkg/router/router.go#L212-L266)
- Shoutrrr Context7/README 的部分 API 描述与所核验源码不一致，接入时应固定依赖版本并以对应源码及实际接口验证；不能假定全渠道去重、恰好一次投递或所有发送都能被超时取消。
- Huma 可以从 Go 路由和输入/输出类型生成 OpenAPI，默认提供请求验证，已由 Q17 选择。oapi-codegen 作为调查过的替代路线，可以从规范生成 Go 类型和 server interfaces，但 strict-server 不等于完整请求验证，也不提供完整响应验证。[Huma 规范生成](https://huma.rocks/features/openapi-generation/)、[Huma 请求验证](https://huma.rocks/features/request-validation/)、[oapi-codegen 验证责任](https://github.com/oapi-codegen/oapi-codegen/blob/v2.8.0/README.md#requestresponse-validation-middleware)
- oapi-codegen v2.8.0 发布说明已有初步 OpenAPI 3.1 支持；不能沿用“完全不支持 3.1”的旧资料。具体依赖和规范版本尚未选择。[发布说明](https://github.com/oapi-codegen/oapi-codegen/releases/tag/v2.8.0)
- Beszel 可提供服务器摘要、历史图表和容器摘要，字段依赖对应 Hub 版本。所核验源码的默认历史数据按分辨率分层，最长保留 30 天；不代表 Octopulse 已接受这一留存或展示策略。[指标列表](https://github.com/henrygd/beszel#supported-metrics)、[保留策略源码](https://github.com/henrygd/beszel/blob/5b0952ffcf0b3cf0a8f975e2fa396db679a0dab7/internal/records/records_deletion.go#L60)
- Beszel readonly 普通用户仍可创建自身告警，不能把该角色描述为绝对无写权限。PocketBase auth-refresh 基于尚有效的当前 token，没有独立 refresh token；过期后需重新认证。Hub 配置可能禁用密码登录或启用 MFA，具体支持版本与认证作为实现阶段兼容验证。[用户角色](https://beszel.dev/guide/user-accounts)、[PocketBase 认证](https://pocketbase.io/docs/authentication/)、[Beszel 环境配置](https://beszel.dev/guide/environment-variables)
- 多 Hub 读取是 Octopulse 可选择实现的集成能力，Beszel 没有提供跨 Hub 联邦查询契约；该能力与 uptime 分布式探测无关。
- Go 的 Transport 透明网络重试与平台轮次追加重试不同；重定向可能改变方法，gzip 自动处理随请求头设置变化。完整 HTTP 设置需统一这些语义。[Go Client](https://pkg.go.dev/net/http#Client)、[Transport](https://pkg.go.dev/net/http#Transport)
- 管理会话使用 host-only cookie 和写操作来源/CSRF 校验，匿名公共接口不因登录态增加秘密字段；页面公开不取消管理权限边界。[Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie)、[CSRF](https://developer.mozilla.org/en-US/docs/Web/Security/Attacks/CSRF)
- 独立域名的 Host 绑定、DNS 解析和 HTTPS 是不同部署步骤，不能把保存页面绑定描述为自动建立 DNS/申请证书。[Host](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Host)、[域名和 DNS](https://developer.mozilla.org/en-US/docs/Learn_web_development/Howto/Web_mechanics/What_is_a_domain_name)
