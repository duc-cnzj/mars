---
title: 🚀 快速开始
lang: zh-CN
---

# 🚀 快速开始

这一页带你从零开始：装好 Mars，部署第一个应用。

::: tip
Mars 默认**零外部依赖**（SQLite + 内存队列），可以先跑起来体验，之后再接真实集群和 Git 服务。
:::

## 你需要准备

| 条件 | 用途 | 说明 |
|---|---|---|
| 一台 Linux / macOS 机器（Windows 用 WSL） | 运行 Mars | 小服务器甚至笔记本都行 |
| 一个 Kubernetes 集群 | 真正部署应用 | 一开始用测试集群（kind / minikube）也可以 |
| 一个 Git 服务 + token | 拉你的代码和 chart | 比如 GitLab，需要建一个访问 token |

## 第 1 步 — 拿到 Mars

任选一种：

```bash
# 方式 A：源码构建
make build                      # 生成 ./bin/app
```

```bash
# 方式 B：下载二进制
# 去 https://github.com/DuC-cnZj/mars/releases 下载最新版
```

## 第 2 步 — 生成配置

```bash
./bin/app init                  # 生成 config.yaml（已存在则跳过）
```

## 第 3 步 — 连接 Git 服务和集群

打开 `config.yaml`，填两处（详见 [系统配置](./configuration.md)）：

```yaml
git_server_plugin:
  name: gitlab
  args:
    token: "你的-gitlab-token"        # GitLab → Settings → Access Tokens 里创建
    baseurl: "https://gitlab.com/api/v4"

kubeconfig: "/home/you/.kube/config"    # 指向你集群的 kubeconfig
```

::: tip
第一次试跑可以跳过这步——Mars 用默认的内存配置也能跑起来。
:::

## 第 4 步 — 启动 Mars

```bash
./bin/app serve -c config.yaml
```

打开浏览器：

- **地址**：`http://127.0.0.1:4000`
- **默认账号**：`admin` / `123456`

## 第 5 步 — 部署第一个应用

1. **创建命名空间** — 点右上角的 **+**。命名空间就是给应用分组的空间。（先保持公开即可。）
2. **打开项目** — 点进命名空间，为你的 Git 仓库创建项目。
3. **设置 chart 目录** — 告诉 Mars 仓库里的 Helm chart 在哪。chart 在仓库根目录，直接写文件夹名（如 `charts`）；在别的项目里，用 `项目id|分支|路径` 的格式。
4. **配置 values** — Mars 会加载一份 `values.yaml` 给你改。可以用内置变量，比如 `<.Branch>` 代表分支名、`<.Host1>` 代表域名。详见 [部署应用](./projects.md)。
5. **点「部署」** — 看着它跑起来。调试时用 debug 模式，某一步失败也会继续，方便排查。
6. **验证** — 打开应用的日志看看，域名配好后点访问链接。

搞定，你的应用上线了。要发新版本，改项目里的镜像 tag 再部署一次即可。

## 下一步

- [系统配置](./configuration.md) — 配置单点登录、私有镜像仓库、存储等
- [部署应用](./projects.md) — 命名空间、values 变量、升级与回滚
- [容器与日志](./containers.md) — 终端、日志、文件操作
