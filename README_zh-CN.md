<div align="center">

[English](README.md) | 简体中文

</div>

<h1 align="center">Mars</h1>
<div align="center"><img style="width: 100px;height: 100px" src="./frontend/public/logo192.png" /></div>
<p align="center">专为devops而生，30秒内部署一个应用。</p>
<br><br>

<div align="center">

[![codecov](https://codecov.io/gh/duc-cnzj/mars/branch/master/graph/badge.svg?token=EUSLRBT6NN)](https://codecov.io/gh/duc-cnzj/mars)
[![unittest](https://github.com/duc-cnzj/mars/actions/workflows/test.yaml/badge.svg)](https://github.com/duc-cnzj/mars/actions/workflows/test.yaml)
[![Release](https://img.shields.io/github/release/duc-cnzj/mars.svg)](https://github.com/duc-cnzj/mars/releases/latest)
[![GitHub license](https://img.shields.io/github/license/duc-cnzj/mars)](https://github.com/duc-cnzj/mars/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/duc-cnzj/mars/v6)](https://goreportcard.com/report/github.com/duc-cnzj/mars/v6)
[![Documentation](https://pkg.go.dev/badge/github.com/duc-cnzj/mars/api/v6/grpc.svg)](https://pkg.go.dev/github.com/duc-cnzj/mars/api/v6/grpc)

</div>

[查看文档](https://duc-cnzj.github.io/mars/)

## 💡 简介

[Mars](https://github.com/duc-cnzj/mars) 是一款专门为 devops 服务的一款应用，基于 kubernetes 之上，可以在短短几秒内部署一个和生产环境一模一样的应用。它打通了 git、kubernetes、helm，通过 git ci 构建镜像，然后通过 kubernetes 部署高可用应用，一气呵成。

## 🗺️ 背景

随着 devops 概念的兴起，现在软件开发不仅要求开发效率高，而且还要求部署便捷，最好能做到流水线开发打包测试上线一条龙服务。
[Mars](https://github.com/duc-cnzj/mars) 由此而生，它打通了打包、测试、部署，基于 git ci/cd 做到任何人不管是开发大牛，还是不懂代码的产品小白，都能在 30 秒部署一个生产级别的应用。真真做到一教即会，高效生产。

## ✨ 特性

- 支持基于 helm charts 开发的任何应用。
- 支持自动配置 https 域名。
- 支持高可用，弹性部署。
- 支持命令行操作。
- 支持查看容器日志。
- 支持查看容器 cpu 和内存使用情况。
- 插件化
  - 队列驱动: ws_sender_nsq, ws_sender_redis, ws_sender_memory
  - 证书驱动: manual_domain_manager, cert-manager_domain_manager, sync_secret_domain_manager, default_domain_manager
  - 代码仓库支持: gitlab ~~github~~
  - 背景图: picture_cartoon，picture_bing
- sdk 接入: [api/](api/)

## 🚀 快速开始

> 默认配置（sqlite 存储 + 内存消息队列）**零外部依赖**，先启动体验，再接入真实集群。

```bash
# 1. 编译
make build                 # 产物 ./bin/app

# 2. 生成默认配置 config.yaml（已存在则跳过）
./bin/app init

# 3. 启动服务
./bin/app serve -c config.yaml

# 4. 打开 Web 界面
#    http://127.0.0.1:4000   默认账号 admin / 123456
```

没有 Go 环境？直接用源码跑：`go run main.go serve -c config.yaml`。

真正把应用部署进 Kubernetes 集群，还需在 `config.yaml` 里配置两处：

- `git_server_plugin`：填入你的 GitLab `token` 与 `baseurl`；
- `kubeconfig`：集群外运行 mars 时，指向你的 kubeconfig 文件。

需要 NSQ/Redis 做消息队列、或 MySQL 做存储时，[dev/docker-compose.yml](dev/docker-compose.yml) 提供这些基础设施（只含依赖，不含 mars 本体）：

```bash
make dev-up    # 等价 docker compose -f dev/docker-compose.yml up -d
# 单服务：make dc-up SVC=redis；关闭：make dev-down / make dc-down SVC=redis
```

## ⚙️ 配置

mars 通过 `-c, --config` 读取 YAML 配置，默认使用当前目录 `config.yaml`（`./bin/app init` 可生成，完整示例见 [config_example.yaml](config_example.yaml)）。

| 配置项                      | 默认                             | 说明                                                    |
| --------------------------- | -------------------------------- | ------------------------------------------------------- |
| `app_port`                  | `4000`                           | HTTP/JSON（gateway）端口；`--app_port` flag 默认 `6000` |
| `grpc_port`                 | `50000`                          | gRPC 端口                                               |
| `db_driver` / `db_database` | `sqlite` / `/tmp/mars-sqlite.db` | 存储驱动：`sqlite` / `mysql`                            |
| `git_server_plugin`         | gitlab                           | 代码仓库，需 `token` + `baseurl`                        |
| `ws_sender_plugin`          | `ws_sender_memory`               | 实时消息：memory / nsq / redis                          |
| `domain_manager_plugin`     | `default_domain_manager`         | HTTPS 域名：manual / sync_secret / cert-manager         |
| `picture_plugin`            | `picture_bing`                   | 背景图：bing / cartoon                                  |
| `cache_driver`              | `memory`                         | 缓存：`memory` / `db`                                   |
| `admin_password`            | `123456`                         | 管理员密码                                              |
| `kubeconfig`                | 空                               | 集群外运行需配置                                        |

## 🖥️ 命令行

二进制根命令为 `app`：

| 命令          | 说明                                                                        |
| ------------- | --------------------------------------------------------------------------- |
| `app serve`   | 启动服务（api / metrics / cron / profile）                                  |
| `app inspect` | 查看运行信息：`all` / `tags` / `cronjobs` / `events` / `plugins` / `config` |
| `app init`    | 生成默认 `config.yaml`                                                      |

`app serve` 常用 flag：

```bash
app serve -c config.yaml \
  --app_port 4000 \
  --grpc_port 50000 \
  --metrics_port 9091 \
  --kubeconfig ~/.kube/config \
  --exclude_server metrics,cron,profile
```

## 🔨 本地开发

```bash
make build      # 编译 ./bin/app
make serve      # go run main.go serve
make test       # 全量单测（-race -cover）
make cover-web  # 覆盖率报告
make api        # protoc 重新生成 proto
make gen        # go generate ./...
make fmt        # gofmt 格式化
make lint       # golangci-lint
make sec        # gosec 安全扫描
```

## 🏗️ 架构

```text
  开发者 ──git push──▶ GitLab ──CI 构建──▶ 镜像仓库
                                              │ 部署时拉取
                                              ▼
  Web / CLI / SDK ──▶ mars ──helm 渲染─────▶ Kubernetes 集群
                        │                      （高可用应用 + HTTPS 域名）
                        ├─ 插件：gitlab · domain_manager · ws_sender · picture
                        └─ 存储：SQLite | MySQL（ent）· 可选 S3
```

## 🧰 技术栈

- **语言/运行时**：Go
- **API**：gRPC + grpc-gateway（HTTP/JSON）
- **编排**：Kubernetes · Helm v3
- **数据**：ent（ORM）· SQLite / MySQL
- **CLI/配置**：cobra · viper
- **可选依赖**：NSQ / Redis（消息队列）· MinIO / S3（文件存储）

## 🤝 贡献

欢迎 PR。提交前请确保：

- 新增代码补齐单测（项目标准：手写生产代码 100% 覆盖、零死代码）；
- `make test` + `make lint` + `make sec` 全绿；
- proto 有变更时跑 `make api` 重新生成并提交产物。

文档与示例见 [doc/](doc/)（OpenAPI）与 [examples/](examples/)。

## 📄 License

[AGPL-3.0](LICENSE)
