---
title: 🐳 Container Terminal & Logs
lang: en-US
---

# 🐳 Container Terminal & Logs

mars provides container interaction over WebSocket; the frontend is built on `xterm`, so you can troubleshoot right in the browser.

## Terminal (Exec / ExecOnce)

Open an interactive terminal for the selected container:

| API | Description |
|---|---|
| `Exec` | establish a persistent WebSocket terminal session |
| `ExecOnce` | run a single command and return its output |

Typical use: shell into a container to debug, run a one-shot diagnostic command.

## Logs (StreamContainerLog)

Stream container stdout/stderr in real time, filtered by namespace, project or container.

```bash
# stream logs via the HTTP/JSON SDK (SSE/NDJSON)
go run ./examples/http -action logs
```

## File Operations (CopyToPod)

| API | Description |
|---|---|
| `CopyToPod` | copy a local file into a container |
| `StreamCopyToPod` | stream file copy |
| `POST /api/copy_from_pod` | copy a file out of a container |

## Metrics

| API | Description |
|---|---|
| `TopPod` / `StreamTopPod` | pod resource ranking in a namespace (real-time stream) |
| `CpuMemoryInProject` | CPU / memory usage per project |
| `CpuMemoryInNamespace` | CPU / memory usage per namespace |

## Permissions

All container APIs require **namespace-level access** (`RequireNamespaceAccessByName`); see [Access Control](./access-control.md).
