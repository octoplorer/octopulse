---
status: accepted
---

# 每个状态页同时支持路径入口和独立域名

每个公开状态页首版支持默认 `/{slug}` 路径，并可绑定独立主机名，使 `example.com/status1` 与 `status1.example.com/` 指向同一个页面、自定义配置和公开数据。页面自定义只作用于公开状态页，管理后台保持固定界面，两种入口均为首版能力。

平台保存精确域名绑定并按可信 Host 路由。DNS、HTTPS 和反向代理由自托管部署配置，首版提供指引；页面使用 [ADR 0011](./0011-built-in-configurable-status-pages.md) 定义的内置 Vue CSR 界面。
