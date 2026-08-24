---
title: 🔐 权限模型
lang: zh-CN
---

# 🔐 权限模型

mars 对全部传输接口（gRPC + HTTP）采用 **6 级权限判定模型**。权威清单见仓库内 [doc/access_control.md](https://github.com/duc-cnzj/mars/blob/master/doc/access_control.md)。

## 6 级判定模型

| 等级 | 判定 | 说明 |
|---|---|---|
| 🆓 公开 | 命中 `biz.IsPublicMethod` 白名单 | 无需任何凭证即可调用 |
| 🔑 登录即可 | gRPC 登录拦截器校验 Bearer token | 任意有效 token 用户 |
| 🛡️ 命名空间级 | `RequireNamespaceAccessByName/ByID` / `RequireProjectAccess` | 公开空间任意登录；私有空间仅 admin / 创建者 / 成员 |
| 🏠 owner 专属 | `RequireNamespaceOwner` | 仅命名空间创建者 |
| ⭐ admin 专属 | `RequireAdmin` | 仅 admin 角色；fullMethodName 精确命中 allowlist 时豁免 |
| 📄 文件所有者/admin | `RequireFileAccess` | `文件上传者 == 当前用户` 或 admin |

## 三层鉴权链

每个 gRPC 方法按序经过：

1. **登录拦截器**：命中 `IsPublicMethod` 白名单直接放行；其余要求 Bearer token，校验通过后把用户注入上下文
2. **Authorize 门禁**：服务实现 `Authorize` 接口 → 自动调用 `RequireAdmin`（如 file / repo 服务）
3. **方法内访问控制**：方法体开头调用 AccessBiz 的 `Require*` / `Can*` 方法（命名空间 / 项目 / owner 级）

## 常见场景

| 场景 | 权限 |
|---|---|
| 登录 / 设置 / 集群信息 / 背景图 / 版本 | 🆓 公开 |
| 项目管理、部署 | 🛡️ 命名空间级（私有空间成员即可）|
| 命名空间 Transfer / 删除 / 改私有 / 同步成员 | 🏠 owner |
| 文件上传、仓库管理 | 🔑 / ⭐ |
| 会话回放、文件下载 | 📄 所有者或 admin |
| 审计事件 | 🔑 普通用户仅见本人，admin 全量 |

## 安全审计观察

- `namespace.UpdateDesc` 目前无 owner/访问校验，若描述承载敏感信息需收紧（建议统一走 owner 判定）
- 部署（Apply/WebApply）门槛 = 命名空间可访问性，私有空间**任何成员**都可发起部署；如需收紧为 owner/admin，可在 `deploy/apply.go` 将 `RequireNamespaceAccessByID` 换成 owner 判定
- access token 采用「持有即有权」语义（`Revoke`/`Lease` 不校验归属），属合理设计，勿误加归属校验
