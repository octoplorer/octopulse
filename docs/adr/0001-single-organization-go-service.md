---
status: accepted
---

# 单组织自托管，Go 承担常驻服务

Octopulse 面向单组织自托管。API、探测、调度、通知和 Beszel 集成由 Go 常驻服务承担，Node.js 用于前端和契约构建工具链。该选择明确了服务运行与构建的职责边界；首版不采用多租户 SaaS 或 Go、Node.js 双后端运行时。

首版不需要分布式探测，只运行一个 Go 服务实例。前端由本地或 CI 构建。

部署方式见 [运维手册](../operations.md#构建和二进制运行)。
