---
status: accepted
---

# 首版采用内置可配置状态页

用户决定从首版移除 GitHub 完整主题功能，集中做好状态页自定义。状态页使用平台内置的 Vue CSR 界面，通过每页配置设置名称、描述、Logo、品牌色、深浅色、公开链接、监控项及别名、分组和排序，并继续展示统计、公告、维护与证书风险。

首版不包含主题仓库、主题包、安装更新、任意主题脚本、sandbox iframe 或消息桥；Node.js 仅承担前端与契约构建。页面路径及独立域名入口、公开字段白名单和固定后台界面保持既定要求。此决策取代 [ADR 0002](./0002-versioned-status-page-themes.md)、[ADR 0007](./0007-prebuilt-theme-packages.md)，并撤回 [ADR 0010](./0010-sandboxed-public-themes.md) 的未接受提案。
