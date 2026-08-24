---
title: 🔌 API Reference
lang: en-US
---

# 🔌 API Reference

## Ports

| Service | Port | Description |
|---|---|---|
| gRPC | `50000` | binary gRPC |
| HTTP/JSON | `4000` | grpc-gateway |
| metrics | `9090` | Prometheus metrics |

> `app_port` defaults to `4000`; the `--app_port` flag defaults to `6000`.

## OpenAPI

The generated OpenAPI spec lives at [doc/openapi.yaml](https://github.com/duc-cnzj/mars/blob/master/doc/openapi.yaml) in the repository, produced by protoc-gen-openapi. The frontend generates type definitions (`schema.d.ts`) from it via openapi-typescript.

## Services

mars defines the following services via proto (see [api/proto](https://github.com/duc-cnzj/mars/tree/master/api/proto)):

| Service | Description |
|---|---|
| `auth` | login / OIDC auth / user info |
| `changelog` | deployment changelog |
| `cluster` | cluster info |
| `container` | container terminal / logs / file copy |
| `endpoint` | access endpoints |
| `event` | audit events |
| `file` | file upload / download / session playback |
| `git` | GitLab integration: repos / branches / commits / pipelines |
| `mars` | meta info |
| `metrics` | CPU / memory / TopPod |
| `namespace` | namespace management |
| `picture` | background pictures |
| `project` | project deploy / rollback / resource topology |
| `repo` | repo templates |
| `token` | access tokens |
| `types` | common types |
| `version` | version info |
| `websocket` | WebSocket messaging |
