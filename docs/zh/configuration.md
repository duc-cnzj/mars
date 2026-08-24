---
title: ⚙️ 系统配置
lang: zh-CN
---

# ⚙️ 系统配置

Mars 的所有设置都在一个 YAML 文件里：运行目录下的 `config.yaml`。用 `./bin/app init` 生成，改完**重启**生效。

这一页只讲你最可能用到的配置。完整带注释的模板在仓库的 [`config_example.yaml`](https://github.com/duc-cnzj/mars/blob/master/config_example.yaml)。

## 核心设置

| 配置 | 默认值 | 作用 |
|---|---|---|
| `app_port` | `4000` | 用户打开网页的地址：`http://<主机>:4000` |
| `grpc_port` | `50000` | 内部 API 端口（一般不用动）|
| `admin_password` | `123456` | 内置 `admin` 账号的密码——**记得改掉！** |

## 连接你的 Git 服务

配置 `git_server_plugin`，让 Mars 能读你的仓库和 chart：

```yaml
git_server_plugin:
  name: gitlab
  args:
    token: "你的-gitlab-token"
    baseurl: "https://gitlab.com/api/v4"
```

建 GitLab token：**GitLab → Settings → Access Tokens → Add new token**，至少勾 `read_api`。用自建 GitLab 就把 `baseurl` 换成你自己的地址。

## 连接你的 Kubernetes 集群

把 Mars 指向集群的 kubeconfig。如果 Mars 和 `kubectl` 跑在同一台机器，这个文件通常已经有了：

```yaml
kubeconfig: "/home/you/.kube/config"
```

- Mars **在集群内**运行？留空即可，自动用集群内凭据。
- **在集群外**运行？填你的 kubeconfig 文件路径。

Mars 会给它创建的每个命名空间加前缀 `ns_prefix`（默认 `devops-`）。如果和你的集群冲突，可以改掉。

## 登录方式

### 账号密码登录

开箱即用，账号 `admin`、密码是 `admin_password`。用 admin 登录的用户拥有全部权限（见 [权限管理](./access-control.md)）。

### 单点登录（OIDC）

想让员工用公司账号登录，加一个或多个 OIDC 配置：

```yaml
oidc:
  - name: "company-sso"
    enabled: true
    provider_url: "https://sso.company.com"
    client_id: "mars"
    client_secret: "xxxx"
    redirect_url: "http://127.0.0.1:3000/auth/callback"
```

可以配多个 provider，用户登录时自己选。

## 私有镜像仓库

要拉私有 Docker 仓库里的镜像，填上凭据：

```yaml
imagepullsecrets:
  - username: "registry-user"
    password: "registry-password"
    email: "you@example.com"
    server: "registry.example.com"        # 默认 https://index.docker.io/v1/
```

Mars 会把这份凭据挂给部署的应用，让它们能拉私有镜像。

## 存储

Mars 自己的数据（命名空间、项目、部署历史）存在数据库里。

| 配置 | 默认值 | 说明 |
|---|---|---|
| `db_driver` | `sqlite` | `sqlite` 免配置；生产用 `mysql` |
| `db_database` | `/tmp/mars-sqlite.db` | sqlite 是文件路径；mysql 是库名 |
| `db_host` / `db_port` / `db_username` / `db_password` | — | 只有 `mysql` 才要填 |

## 文件上传

| 配置 | 默认值 | 说明 |
|---|---|---|
| `upload_dir` | `/tmp/mars-uploads` | 上传文件存放目录 |
| `upload_max_size` | `50m` | 上传大小上限（如 `100m`、`1g`）|
| `s3_enabled` | `false` | 上传文件存 S3 / MinIO 而不是磁盘 |

## 一些小设置

| 配置 | 默认值 | 作用 |
|---|---|---|
| `picture_plugin` | `picture_bing` | 登录页背景图：`picture_bing` 或 `picture_cartoon` |
| `external_ip` | `127.0.0.1` | 集群对外可达的 IP |
| `install_timeout` | `90s` | 一次部署允许的最长时间，超时算失败 |
| `debug` | `false` | 打详细日志，排查问题用 |

## 命令行

| 命令 | 作用 |
|---|---|
| `./bin/app init` | 生成 `config.yaml`（不存在时）|
| `./bin/app serve -c config.yaml` | 启动服务 |
| `./bin/app inspect` | 查看运行时信息（配置、插件、任务…）|

`serve` 常用参数：

```bash
./bin/app serve -c config.yaml \
  --app_port 4000 \
  --grpc_port 50000 \
  --metrics_port 9091 \
  --kubeconfig ~/.kube/config
```
