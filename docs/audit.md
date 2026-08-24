---
title: 📋 审计与变更记录
lang: zh-CN
---

# 📋 审计与变更记录

mars 对关键操作全程留痕，支持审计追踪与会话回放。

## 变更记录（Changelog）

记录项目每一次部署变更，`FindLastChangelogsByProjectID` 可查看项目最近的部署历史（携带完整 Config / EnvValues），用于**回滚**到历史版本。

## 审计事件（Event）

记录登录、部署等关键事件，按**操作人邮箱**（`operator_email`）归属：

| 角色 | 可见范围 |
|---|---|
| admin | 全部事件 |
| 普通用户 | 仅本人操作的事件 |

- 越权访问他人事件返回 404（视同不存在），防止审计日志 ID 枚举
- 登录认证失败会打 `[auth audit]` Warning 审计日志

## 会话回放（操作录屏）

- 容器终端操作会被**录制**，通过文件服务（`ShowRecords`）回放
- 普通用户仅能回放自己的会话，admin 可回放全部
- HTTP 下载走 `RequireFileAccess`（文件所有者 / admin）

## 相关接口

| 接口 | 说明 |
|---|---|
| `changelog.FindLastChangelogsByProjectID` | 项目最近部署记录 |
| `event.List` / `event.Show` | 审计事件列表 / 详情 |
| `file.ShowRecords` | 会话录屏回放 |

> 所有审计相关操作均受权限模型约束，详见 [权限模型](./access-control.md)。
