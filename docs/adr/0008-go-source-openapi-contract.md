---
status: accepted
---

# Go API 定义作为契约来源

以 Go 类型和路由定义作为 API 契约的唯一来源，通过 Huma 导出 OpenAPI，再由 HeyAPI 生成 TypeScript 类型、SDK 和 Pinia Colada 集成。该路线避免并行维护手写规范与 Go 接口定义；规范、生成代码及实现之间的一致性仍需验证。

具体工具版本、底层路由适配和生成文件管理属于后续工程工作；管理接口和公共状态页的数据权限边界需分别定义。
