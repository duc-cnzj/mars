---
title: 🔌 API 参考
lang: zh-CN
---

# 🔌 API 参考

## 端口约定

| 服务 | 端口 | 说明 |
|---|---|---|
| gRPC | `50000` | 二进制 gRPC |
| HTTP/JSON | `4000` | grpc-gateway 网关 |
| metrics | `9090` | Prometheus metrics |

> `app_port` 默认 `4000`；`--app_port` flag 默认 `6000`。

## OpenAPI

生成的 OpenAPI 规范在仓库内 [doc/openapi.yaml](https://github.com/duc-cnzj/mars/blob/master/doc/openapi.yaml)，由 protoc-gen-openapi 生成。前端通过 openapi-typescript 生成类型定义（`schema.d.ts`）。

## 服务清单

mars 基于 proto 定义了以下服务（见 [api/proto](https://github.com/duc-cnzj/mars/tree/master/api/proto)）：

| 服务 | 说明 |
|---|---|
| `auth` | 登录 / OIDC 认证 / 用户信息 |
| `changelog` | 部署变更记录 |
| `cluster` | 集群信息 |
| `container` | 容器终端 / 日志 / 文件拷贝 |
| `endpoint` | 访问端点 |
| `event` | 审计事件 |
| `file` | 文件上传 / 下载 / 会话回放 |
| `git` | GitLab 集成：仓库 / 分支 / commit / pipeline |
| `mars` | 元信息 |
| `metrics` | CPU / 内存 / TopPod |
| `namespace` | 命名空间管理 |
| `picture` | 背景图 |
| `project` | 项目部署 / 回滚 / 资源拓扑 |
| `repo` | 仓库模板 |
| `token` | 访问 token |
| `types` | 通用类型 |
| `version` | 版本信息 |
| `websocket` | WebSocket 消息 |
