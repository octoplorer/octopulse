# 开发指南

本文中的命令默认在仓库根目录执行；需要进入其他目录的步骤会显式使用 `cd`。构建产物的交付与运行方式见 [运维手册](operations.md#构建和二进制运行)。

## 本地开发

先安装 [mise](https://mise.jdx.dev/)，在仓库根目录执行：

```sh
mise trust
mise install
mise exec -- go mod download
mise exec -- aube install --frozen-lockfile
```

工具版本固定在 [mise.toml](../mise.toml)：Go 1.27.1、Node.js 24.19.0、aube 2.6.1、sqlc 1.30.0。Go 命令和整项目任务编排由 mise 管理，Node 命令定义在根目录私有 [package.json](../package.json)。[aube-workspace.yaml](../aube-workspace.yaml) 纳入 `web` 包，依赖由根目录 `aube-lock.yaml` 锁定。Node.js 只用于开发与构建，生产运行不需要 Node.js。

根目录维护共享工具 `@antfu/eslint-config`、`@types/node`、`eslint`、`eslint-plugin-format` 和 `typescript`。Vue 等运行时依赖，以及 Vite、Vitest、vue-tsc、UnoCSS、HeyAPI 和图标包仍在 `web/package.json`；Vite、TypeScript、UnoCSS 和 OpenAPI 配置保留在 `web/`，ESLint 配置位于根目录。前端工具使用根 workspace 的共享依赖，`web` 包需随 workspace 安装。

同时启动 API 与前端开发服务器：

```sh
mise run dev
```

`mise run dev` 并行执行 `dev:api` 与 `dev:web`；也可用两个终端分别执行 `mise run dev:api` 和 `mise run dev:web`。依赖安装也可用 `mise run install`，使用根锁文件的 frozen 模式。

整项目任务和 Node 脚本的作用不同：

| 任务 | 整项目 | 仅 Node/前端 |
| --- | --- | --- |
| 开发 | `mise run dev` | `mise exec -- aube run dev` |
| 检查 | `mise run check` | `mise exec -- aube run check` |
| 测试 | `mise run test` | `mise exec -- aube run test` |
| 构建 | `mise run build` | `mise exec -- aube run build` |
| 契约生成 | `mise run generate:web`（导出 OpenAPI 后生成客户端） | `mise exec -- aube run generate`（基于现有契约生成客户端） |

根 Node 脚本还提供 `typecheck`、`lint`、`lint:fix` 和 `test:watch`；Go 的开发、测试和构建可分别执行 `mise run dev:api`、`mise run test:go` 和 `mise run build:go`。只需直接调用前端包时，使用 workspace 过滤：

```sh
mise exec -- aube --filter @octopulse/web run <script>
```

打开 `http://127.0.0.1:5173/app`，首次访问创建管理员；之后由管理员添加其他成员。默认 API 地址为 `127.0.0.1:8080`，Vite 将 `/api` 和上传图片请求代理到该地址。密码要求为 12–72 字节，公开注册关闭。

开发代理保留浏览器的 `Host` 和 `Origin`，以通过后台的同源校验。调整代理时保持 `changeOrigin: false`；修改 Vite 配置后，开发服务器会自动重启。

## 前端约定

前端使用 Vite、Vue Router、Vue I18n、VueUse、TanStack Form、Pinia Colada 和 Ark UI；UnoCSS 配置 `preset-wind4` 与 `preset-attributify`，属性样式采用 `un-` 前缀。

代码检查与格式化统一由 ESLint 和 `@antfu/eslint-config` 管理，配置位于根目录 `eslint.config.js`；CSS 和 HTML 通过 `eslint-plugin-format` 格式化。根脚本 `lint` 检查 `web/`、根 `package.json` 和 ESLint 配置；生成的 `web/src/client`、`web/typed-router.d.ts` 和构建产物 `web/dist` 不参与检查。执行 `mise exec -- aube run lint:fix` 自动修复。前端包的 `lint`、`lint:fix`、`format` 和 `format:check` 委托根 workspace 对应命令，统一使用根工具依赖。

### 页面路由

页面路由由 Vue Router 5 官方的 `vue-router/vite` 插件从 `web/src/pages` 自动生成。`app/login.vue` 是独立登录页，`app/(admin).vue` 提供后台布局，`app/(admin)/` 内的页面通过 `definePage()` 声明标题的翻译 key 和角色权限；`[id]` 目录用于动态参数。公开状态页使用 `[[slug]]/[[...rest]]+.vue`，兼容路径入口和独立域名。新增页面只需创建对应 `.vue` 文件；开发服务器会更新路由，插件会在开发和构建时生成 `web/typed-router.d.ts`，页面变更应连同更新后的路由类型声明一起提交。

### 共享状态与工具

前端共享响应式状态放在 `web/src/composables/`（会话、国际化、偏好和通知）；`web/src/lib/` 保留无状态工具和类型。依赖当前语言或时区的格式化函数随偏好模块放在 `composables/preferences.ts`。

页面和展示组件直接引用 `web/src/client/types.gen.ts` 中生成的 API 类型。监控编辑器使用 `web/src/lib/monitor-form.ts` 中的表单类型，描述补齐默认值后的配置；API 数据通过 `toMonitorForm()` 转为表单状态。

编辑表单由 `@tanstack/vue-form` 的 `useForm` 管理，字段通过 `form.Field` 的 `handleChange` / `handleBlur` 更新，显示值和提交状态使用 `form.useSelector` 订阅。打开编辑或加载服务端数据时使用 `reset(values)`；保存后的重置需等异步提交步骤完成，避免提前清空 `isSubmitting`。列表筛选、标签页、弹窗开关和立即生效的主题偏好保留为界面状态。

字段的标签、提示与错误关联使用 Ark UI Field，文本控件通过 `FieldInput` / `FieldTextarea` 绑定 `modelValue`。标签使用 Ark UI TagsInput 直接绑定字符串数组，保存时提交尚未确认的标签。嵌套编辑组件需发出新的对象或数组，不能直接修改表单 store 中的引用。数组操作使用表单 API；动态数组字段以完整字段路径作为 key，确保重排后字段实例绑定正确的位置。列表和 JSON 输入组件直接接收结构化值，在组件内维护文本编辑状态；表单使用 API 所需的数据类型，提交监控时移除非当前监控类型的配置。

### 文档 head

文档 head 使用 Unhead 管理，`App.vue` 通过 `useHead` 声明默认标题、HTML 语言和 metadata；公开路由直接根据已发布配置覆盖标题、描述、品牌色和配色方案，数据清空或组件卸载时自动清理。后台内嵌草稿预览不修改文档 head，字符集与 viewport 保留在 `web/index.html`。

### 国际化与显示偏好

前端国际化使用 Vue I18n Composition API，入口为 `web/src/composables/i18n.ts`，中英文文案分别维护在 `web/src/locales/zh-CN.json` 和 `web/src/locales/en.json`。组件使用 `useI18n({ useScope: 'global' })` 获取 `t`、`n`、`d` 和响应式 `locale`；普通 TypeScript 模块使用共享 composer。新增文案应为两个语言包添加相同的语义 key，变量使用命名插值（如 `t('errors.invalidJSON', { label })`），数量使用完整复数消息（如 `t('counts.monitors', { count }, count)`），避免拼接文案。消息中的字面量 `@`、花括号和 `|` 使用 Vue I18n 的字面量插值语法转义。

语言沿用 API 的 `zh-CN` / `en`，登录后使用个人设置，访客使用 VueUse `useStorage` 保存的 `octopulse.locale` 偏好，默认简体中文；缺失翻译回退到简体中文。日期与数字通过 composer 格式化，日期沿用所选显示时区；表单中的机器日期格式保持固定。增加语言时同时更新语言包、`i18n.ts` 的语言及格式配置和后端允许的语言值。`mise exec -- aube run test` 验证语言包一致性、消息编译、复数和日期/数字格式。用户填写的状态页和事件内容、服务端返回的诊断文案按原内容显示。

## 契约生成

Go 路由与输入/输出类型是 API 契约的来源，Huma 导出 OpenAPI，HeyAPI 生成 TypeScript、SDK 和 Colada 查询选项。mise 的 `generate:web` 先执行 `generate:api`，将契约写入 `api/openapi.json`，再调用 Node 的 `generate`，按 `web/openapi-ts.config.ts` 在 `web/src/client` 生成客户端。Node 的 `generate` 只读取现有契约；单独导出 OpenAPI 可运行 `mise run generate:api`。`mise run generate:db` 调用 sqlc 生成数据库查询。任务定义见 [mise.toml](../mise.toml)，Node 命令见根 [package.json](../package.json)。

修改相关契约后，在仓库根目录执行对应生成任务：

```sh
mise run generate:db
mise run generate:web
```

## 验证

```sh
mise run check
mise exec -- go vet ./...
mise run build
```

前端单元测试使用 Vitest，复用 `web/vite.config.ts`，在 Node 环境中递归发现 `src` 下的 `*.test.ts` 和 `*.spec.ts`。根 Node 脚本 `test` 只运行前端测试，`mise run test` 同时运行 Go 与前端测试；开发时执行 `mise exec -- aube run test:watch`，监听前端文件变化并重跑相关测试。测试文件显式从 `vitest` 导入 `it` 和 `expect`，由 `mise exec -- aube run typecheck` 检查类型。

在仓库根目录运行 `mise exec -- aube run lint` 检查前端与根工具配置；根 Node 脚本 `check` 依次运行 lint、前端类型检查和单元测试，`mise run check` 还运行 Go 测试。`mise run build` 构建 Go 二进制与前端资源，根 Node 脚本 `build` 只构建前端。lint 检查以零警告为通过条件。需要竞态检测时执行 `mise exec -- go test -race -p 1 ./...`。

真实 PostgreSQL 的同一业务套件，在仓库根目录执行：

```sh
OCTOPULSE_TEST_DB_DRIVER=postgres \
OCTOPULSE_TEST_POSTGRES_DSN='postgres://USER@HOST:PORT/octopulse_test?sslmode=disable' \
mise exec -- go test -race -p 1 -count=1 ./...
```

测试数据库必须专用；测试创建并清理独立 schema，部分测试会终止自身锁连接。数据库实例锁跨 schema，`-p 1` 防止不同包并行争抢同一数据库。安装匹配的 `pg_dump` 和 `pg_restore`，才能运行真实 PostgreSQL 备份恢复用例。

[CI](../.github/workflows/verify.yml) 安装 mise 固定工具链和根 workspace 锁定依赖，使用根 Node 脚本执行 lint、类型检查、前端单元测试和前端构建，并打包 Linux amd64 二进制与 `web/dist`；契约生成由 mise 编排，同时对 SQLite 及真实 PostgreSQL 运行竞态检测，检查 sqlc/OpenAPI/HeyAPI 生成文件漂移。工作流手动触发时可选择额外容量测试；完整负载命令与限制见 [容量记录](capacity.md)。

首版交付的本地验收记录覆盖 Darwin arm64 与原生 Linux arm64 容器；后续改动的验证记录和当前实现差距见 [验收映射](acceptance.md)。历史结果不代表当前提交已重新完成整体验收，远程 CI 和 Linux amd64 产物的验证状态以对应提交的工作流结果为准。
