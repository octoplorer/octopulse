# Web 组件库

组件位于 `web/src/components/ui/`，以 Cloudflare Kumo 的设计和行为为参考，使用 Vue 3、Ark UI 与 UnoCSS Wind4 实现。上游基线为 [Kumo 543dc8a](https://github.com/cloudflare/kumo/tree/543dc8a0ca1b6f7678eb9c27be6750965eaf2564)。派生代码和素材的许可见 `web/src/components/ui/LICENSE.kumo`。

## 完整组件清单

下表覆盖该基线的 48 个组件目录；按项目要求不提供 CloudflareLogo 与 PoweredByCloudflare 品牌组件。Vue 使用命名导出与命名子组件；模型和事件遵循 Vue 约定。

| 上游组件 | 本地入口 | 主要导出 |
| --- | --- | --- |
| Autocomplete | `autocomplete` | Autocomplete |
| Badge | `badge` | Badge |
| Banner | `banner` | Banner、BannerAction、BannerTitle、BannerDescription |
| Breadcrumbs | `breadcrumbs` | Breadcrumbs |
| Button | `button` | Button、RefreshButton、LinkButton |
| ButtonGroup | `button-group` | ButtonGroup |
| Chart | `chart` | Chart、ChartPalette、SankeyChart、TimeseriesChart、ChartLegend、BubbleMap、ChoroplethMap、GlobeMap |
| Checkbox | `checkbox` | Checkbox、CheckboxGroup、CheckboxItem |
| ClipboardText | `clipboard-text` | ClipboardText |
| Code | `code` | Code、CodeBlock |
| Collapsible | `collapsible` | Collapsible、CollapsibleTrigger、CollapsibleContent |
| Combobox | `combobox` | Combobox |
| CommandPalette | `command-palette` | CommandPalette、CommandPalettePanel |
| DatePicker | `date-picker` | DatePicker |
| DateRangePicker | `date-range-picker` | DateRangePicker |
| Dialog | `dialog` | Dialog、DialogRoot、DialogTrigger、DialogTitle、DialogDescription、DialogClose |
| DropdownMenu | `dropdown-menu` | DropdownMenu 及 Trigger、Content、Item、CheckboxItem、RadioGroup、Submenu 等 |
| Empty | `empty` | Empty、EmptyMedia、EmptyTitle、EmptyDescription |
| Field | `field` | Field、FieldLabel、FieldDescription、FieldError、FieldGroup、FieldSection、FieldActions |
| Flow | `flow` | Flow、FlowRoot、FlowNode、FlowParallel、FlowList、FlowAnchor |
| Grid | `grid` | Grid、GridItem |
| InlineCopyText | `inline-copy-text` | InlineCopyText |
| Input | `input` | Input、InputArea、Textarea |
| InputGroup | `input-group` | InputGroup、InputGroupInput、InputGroupAddon、InputGroupSuffix、InputGroupButton |
| Label | `label` | Label |
| LayerCard | `layer-card` | LayerCard、LayerCardPrimary、LayerCardSecondary |
| LayerDialog | `layer-dialog` | LayerDialog 及 Content、Body、Actions、Action 等 |
| Link | `link` | Link |
| Loader | `loader` | Loader、SkeletonLine |
| MenuBar | `menubar` | MenuBar、MenuBarItem |
| Meter | `meter` | Meter |
| Pagination | `pagination` | Pagination、PaginationControls、PaginationInfo、PaginationPageSize |
| Popover | `popover` | Popover 及 Trigger、Content、Title、Description、Close 等 |
| Radio | `radio` | Radio、RadioGroup、RadioItem |
| Select | `select` | Select、SelectTrigger、SelectContent、SelectItem、SelectValue 等 |
| SensitiveInput | `sensitive-input` | SensitiveInput |
| Sidebar | `sidebar` | Sidebar、SidebarProvider 及组、菜单、折叠、滑动视图、调整大小等 |
| Slider | `slider` | Slider |
| Surface | `surface` | Surface |
| Switch | `switch` | Switch |
| Table | `table` | Table、TableContainer、TableHeader、TableBody、TableRow、TableHead、TableCell 等 |
| TableOfContents | `table-of-contents` | TableOfContents、TableOfContentsList、TableOfContentsItem、TableOfContentsGroup |
| Tabs | `tabs` | Tabs、TabsRoot、TabsList、TabsTrigger、TabsContent |
| TagInput | `tag-input` | TagInput |
| Text | `text` | Text |
| Toast | `toast` | Toasty、ToastProvider、createKumoToastManager、useKumoToastManager |
| Toolbar | `toolbar` | Toolbar、ToolbarButton、ToolbarLink、ToolbarInput、ToolbarInputGroup |
| Tooltip | `tooltip` | Tooltip、TooltipProvider |

复合区块：`delete-resource` 提供 DeleteResource，`blocks/page-header` 提供 PageHeader，`blocks/resource-list` 提供 ResourceListPage。Surface、MenuBar 和 DateRangePicker 虽已在上游弃用，仍提供等价的组合实现。

## 使用约定

- 从具体目录导入，避免在业务页面引用整库；`ui/index.ts` 用于统一发现和组件预览。
- 表单控件使用 `v-model`；浮层使用 `v-model:open`。Select 单选为字符串，多选为字符串数组；日期采用 ISO 日期字符串数组，范围为两个日期。
- Field 管理 label、description、error 与控件的 ARIA 关联；可组合 Input、InputArea 和其它 Ark 控件。
- 多变体使用 `cva@1.0.0-beta.12` 的对象配置。所有变体保留完整静态类名，UnoCSS 扫描 `.vue` 和 `.ts`；条件变体保持互斥，CVA 只组合类名。
- Wind4 的 `base` 是背景色名称；字号使用 `text-size-base`。状态文字使用 `text-fg-danger` 等前景 token，背景使用 `bg-danger`、`bg-danger-tint`。
- 焦点轮廓需显式声明 `focus-visible:outline-solid`；Wind4 中 `outline-none` 后仅增加 `outline-2` 不会恢复轮廓样式。全局基础样式位于 `@layer base`，避免覆盖组件。焦点按组件对齐上游：普通输入为无间距的 1.5px 半透明 ring，按钮键盘焦点为 2px brand ring，日历格为内侧 brand ring。
- Ark 的状态使用 `data-[state=checked]`、`data-[selected]`、`data-[disabled]` 等实际属性；不直接复制 Base UI 的状态选择器或 Tailwind 动画插件类。
- Dialog 内的菜单、选择器、日历与提示自动保留在模态内容内，避免 body portal 被隐藏或阻挡。浮层使用 fixed 定位，层级声明在 Content 上，供 Ark 的 Positioner 样式镜像读取。
- 主题由根节点 `data-mode="light|dark"` 与语义 token 控制。图表使用本地上下文解析颜色，并尊重减少动态效果设置。
- Label、closeLabel、copyLabel、emptyLabel 等用户可见文案由业务页面通过 vue-i18n 传入。通用组件不依赖应用状态或 API 客户端。

## 预览与验证

启动 `mise run dev:web` 后访问 `/ui.html` 可独立预览全部 48 组组件与复合区块，切换主题和禁用、只读、错误状态。分类锚点提供基础输入、展示布局、选择日期、导航浮层和图表流程。菜单、通知、标签、日历、侧栏、表格选择和列宽、异步删除、图表数据选择都可以实际操作。此入口仅用于开发，未加入生产构建或应用路由。

测试参考上游 Kumo 的交互场景，使用真实 Vue 组件与 Ark 行为，覆盖受控模型、禁用状态、键盘导航、焦点、分页边界、异步删除、标签提交、复制反馈、通知更新、日期范围以及图表和 Flow 的数据转换。

执行 `mise run check:web` 和 `mise run build:web`。不要以源文件字符串或类名存在性测试代替用户交互验证。
