---
title: RBAC
lang: zh-cn
---

# RBAC

## 本地账户

目前只有管理员账户默认有超级管理员权限

## sso

> version: v4.26+

可以在 sso 返回的 Claims 中增加一个 `roles []string` 字段，并写入 `mars_admin` 即可获得超级管理员权限。
