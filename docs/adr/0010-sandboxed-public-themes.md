---
status: deprecated
---

# 用公共页面外壳承载隔离的完整主题

2026-10-01：该提案未被接受，随 GitHub 主题移出首版而撤回，见 [ADR 0011](./0011-built-in-configurable-status-pages.md)。当前首版直接使用内置 Vue CSR 页面，以下仅保留历史讨论。

为同时支持同域路径、独立域名和完整 JavaScript 主题，建议由平台控制公开页面薄外壳，主题 UI 在 sandbox iframe 中运行，CSS/JS 不装载到管理后台或外壳 DOM。主题文档响应强制 CSP sandbox，公开数据通过匿名 API 与受限消息桥提供，后台接口保持独立鉴权与 CSRF 校验。

该方案原拟保留主题的完整视觉与交互定制，也限制主题访问后台 DOM、会话和存储；公开导航与展示偏好原拟由外壳接口支持。原 Q27 已失效，本方案未实现也不进入首版验收。
