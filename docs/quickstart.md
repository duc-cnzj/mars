---
title: 🚀 快速开始
lang: zh-CN
---

# 🚀 快速开始

::: tip
默认配置（SQLite 存储 + 内存消息队列）**零外部依赖** —— 可以先启动起来，再接真实集群。
:::

## 前置条件

- 一台 Linux / macOS 主机（Windows 可用 WSL）
- 可选：一个 Kubernetes 集群（要真正部署应用时需要）+ 一个 GitLab

## 安装与启动

### 1. 构建

```bash
make build        # 生成 ./bin/app
```

没有 Go 工具链？可以直接用源码运行：`go run main.go serve -c config.yaml`。

也可以直接去 [Release 页面](https://github.com/DuC-cnZj/mars/releases) 下载二进制包。

### 2. 生成默认配置

```bash
./bin/app init    # 生成 config.yaml（已存在则跳过）
```

### 3. 启动服务

```bash
./bin/app serve -c config.yaml
```

启动完成后打开 Web 界面：

- **地址**：`http://127.0.0.1:4000`
- **默认账号**：`admin` / `123456`

### 4. 连接真实集群

零依赖模式只适合体验，要真正把应用部署到 Kubernetes，需要配置两处（见 [配置参考](./configuration.md)）：

- `git_server_plugin`：填写 GitLab 的 `token` 和 `baseurl`
- `kubeconfig`：在集群外部运行时，指向你的 kubeconfig 文件

### 5. （可选）基础设施

需要 NSQ/Redis 消息队列、MySQL 存储、MinIO 时，[dev/docker-compose.yml](https://github.com/duc-cnzj/mars/blob/master/dev/docker-compose.yml) 提供了这些依赖（仅基础设施，不含 mars 本身）：

```bash
make dev-up                              # docker compose -f dev/docker-compose.yml up -d
# 单服务：make dc-up SVC=redis
# 拆除：  make dev-down / make dc-down SVC=redis
```

## 部署第一个应用

以 Web 界面操作为例（全部文字步骤，界面以实际版本为准）：

1. **登录**：打开 `http://127.0.0.1:4000`，用 `admin` / `123456` 登录
2. **创建命名空间**：点击右上角 `+`，填写命名空间配置并创建
3. **配置项目**：进入项目，配置 charts 目录
   - charts 就在项目目录下：直接写相对路径
   - 引用其他项目的 charts：按 `项目id|分支|相对路径` 格式填写
4. **配置 values**：保存 charts 路径后，会自动加载默认 `values.yaml` 供参考，按提示配置其他字段（详见 [项目管理](./projects.md)）
5. **点击部署**：点击部署按钮，`debug` 模式即 helm 中的 `atomic = false`
6. **验证**：部署完成后，查看容器状态、打开日志、点击访问即可

部署完成后，可以试试**覆盖配置**：修改镜像 tag 再次部署，即可滚动升级到新版本。
