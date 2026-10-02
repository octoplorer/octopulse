---
status: accepted
---

# Go API 定义作为契约来源

以 Go 类型和路由定义作为 API 契约的唯一来源，通过 Huma 导出 OpenAPI，再由 HeyAPI 生成 TypeScript 类型、SDK 和 Pinia Colada 集成。该路线避免并行维护手写规范与 Go 接口定义；规范、生成代码及实现之间的一致性仍需验证。

管理接口与匿名公共接口采用独立数据投影，在服务端限制公开字段，避免依赖客户端隐藏秘密。

契约生成与验证流程见 [开发指南](../development.md#契约生成)。
