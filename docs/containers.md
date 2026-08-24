---
title: 🐳 容器终端与日志
lang: zh-CN
---

# 🐳 容器终端与日志

mars 通过 WebSocket 提供容器交互能力，前端基于 `xterm` 实现，浏览器里即可完成排障。

## 终端（Exec / ExecOnce）

在网页上直接打开所选容器的交互终端：

| 接口 | 说明 |
|---|---|
| `Exec` | 建立持久 WebSocket 终端会话 |
| `ExecOnce` | 执行一次命令并返回输出 |

典型场景：进入容器调试、执行一次性诊断命令。

## 日志（StreamContainerLog）

实时拉取容器标准输出/错误日志，支持按命名空间、项目、容器维度查看。

```bash
# 通过 HTTP/JSON SDK 拉取日志（SSE/NDJSON 流式）
go run ./examples/http -action logs
```

## 文件操作（CopyToPod）

| 接口 | 说明 |
|---|---|
| `CopyToPod` | 把本地文件拷贝进容器 |
| `StreamCopyToPod` | 流式拷贝 |
| `POST /api/copy_from_pod` | 从容器拷出文件 |

## 资源监控（metrics）

| 接口 | 说明 |
|---|---|
| `TopPod` / `StreamTopPod` | 命名空间内 Pod 资源排行（实时流）|
| `CpuMemoryInProject` | 项目维度 CPU / 内存使用 |
| `CpuMemoryInNamespace` | 命名空间维度 CPU / 内存使用 |

## 权限

所有容器接口均要求**命名空间级访问权限**（`RequireNamespaceAccessByName`），详见 [权限模型](./access-control.md)。
