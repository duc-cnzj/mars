---
title: ⚙️ 配置参考
lang: zh-CN
---

# ⚙️ 配置参考

Mars 通过 `-c, --config` 读取 YAML 配置，默认读取当前目录下的 `config.yaml`（由 `./bin/app init` 生成；完整示例见 [config_example.yaml](https://github.com/duc-cnzj/mars/blob/master/config_example.yaml)）。

## CLI 命令

根命令为 `app`：

| 命令 | 说明 |
|---|---|
| `app serve` | 启动全部服务（api / metrics / cron / profile）|
| `app inspect` | 查看运行时信息：`all` / `tags` / `cronjobs` / `events` / `plugins` / `config` |
| `app init` | 生成默认 `config.yaml` |

`app serve` 常用参数：

```bash
app serve -c config.yaml \
  --app_port 4000 \
  --grpc_port 50000 \
  --metrics_port 9091 \
  --kubeconfig ~/.kube/config \
  --exclude_server metrics,cron,profile
```

> 配置还支持环境变量覆盖，前缀为 `MARS`。

## 配置字段总表

对齐 `config_example.yaml`，未注释字段均为默认值。

### 服务与通用

| 字段 | 默认值 | 说明 |
|---|---|---|
| `app_port` | `4000` | HTTP/JSON（gateway）端口 |
| `grpc_port` | `50000` | gRPC 端口 |
| `metrics_port` | `9090` | metrics 端口 |
| `debug` | `false` | 调试模式 |
| `exclude_server` | `""` | 需要排除的 server tag，如 `cron,api,metrics` |
| `log_channel` | `zap` | 日志通道：`logrus` / `zap` |
| `git_server_cached` | `true` | 开启 git 请求缓存（git 请求较慢，默认开启）|
| `db_auto_migrate` | `false` | 是否自动迁移数据库表结构 |
| `cache_driver` | `memory` | 缓存驱动：`memory` / `db` |
| `external_ip` | `127.0.0.1` | 集群外网访问 IP |
| `ns_prefix` | `devops-` | mars 管理的命名空间统一加此前缀 |
| `install_timeout` | `90s` | helm 安装超时 |
| `tracing_endpoint` | `""` | OTLP 链路追踪端点（如 Jaeger 4317）|
| `admin_password` | `123456` | admin 账号密码 |
| `private_key` | `""` | RSA 私钥（签发访问 token 用）|

### 存储

| 字段 | 默认值 | 说明 |
|---|---|---|
| `db_driver` | `sqlite` | 存储驱动：`sqlite` / `mysql` |
| `db_database` | `/tmp/mars-sqlite.db` | sqlite 时为 db 绝对路径；mysql 时为库名 |
| `db_host` / `db_port` / `db_username` / `db_password` | `127.0.0.1` / `13306` / `root` / `""` | mysql 时必填，sqlite 忽略 |
| `db_slow_log_enabled` | `true` | 开启慢查询日志 |
| `db_slow_log_threshold` | `200ms` | 慢查询阈值：`ns` / `us` / `ms` / `s` / `m` / `h` |
| `db_debug` | `false` | 开启 ent Debug 模式（打印 SQL）|

### 文件与 S3

| 字段 | 默认值 | 说明 |
|---|---|---|
| `upload_dir` | `/tmp/mars-uploads` | 上传文件保存目录 |
| `upload_max_size` | `50m` | 上传大小上限（支持 `MB` / `Gi` / `m` / `g`）|
| `s3_enabled` | `false` | 是否上传文件到 S3 |
| `s3_endpoint` | `""` | MinIO / S3 端点 |
| `s3_access_key_id` / `s3_secret_access_key` | `minioadmin` | S3 凭据 |
| `s3_use_ssl` | `false` | 是否使用 SSL |
| `s3_bucket` | `mars` | S3 bucket 名 |

### 插件

#### git_server_plugin（Git 服务端）

```yaml
git_server_plugin:
  name: gitlab
  args:
    token: ""
    baseurl: "https://gitlab.com/api/v4"
```

#### ws_sender_plugin（实时消息队列）

```yaml
ws_sender_plugin:
  name: ws_sender_memory
#  name: ws_sender_nsq
#  args:
#    addr: 127.0.0.1:4150
#    # lookupd_addr 可选；必须是 nsqlookupd HTTP 端口 4161（4160 是 TCP 端口）
#    lookupd_addr: 127.0.0.1:4161
#    msg_timeout: 120        # 消息处理超时，默认 60s
#    dial_timeout: 5         # 连接建立超时，默认 1s
#    read_timeout: 120       # 读取超时，默认 60s
#    write_timeout: 5        # 写入超时，默认 1s
#    heartbeat_interval: 30  # 心跳间隔，默认 30s
#  name: ws_sender_redis
#  args:
#    addr: 127.0.0.1:6379
#    password: ""
#    db: 1
```

::: warning
使用 NSQ 时注意：`lookupd_addr` 必须填 **nsqlookupd 的 HTTP 端口 4161**，4160 是 TCP 端口。
:::

#### domain_manager_plugin（域名/证书管理）

```yaml
domain_manager_plugin:
  name: default_domain_manager
#  name: manual_domain_manager
#  args:
#    wildcard_domain: "*.mars.local"
#    tls_key: ""
#    tls_crt: ""
#  name: sync_secret_domain_manager
#  args:
#    wildcard_domain: "*.mars.local"
#    secret_name: ""
#    secret_namespace: ""
#  name: cert-manager_domain_manager
#  args:
#    wildcard_domain: "*.mars.local"
#    cluster_issuer: "letsencrypt-mars"
```

#### picture_plugin（登录页背景图）

```yaml
picture_plugin:
  name: picture_bing
#  name: picture_cartoon
```

### 私有镜像仓库

```yaml
imagepullsecrets:
  - username: "jack"
    password: "12345"
    email: "jack@example.com"
    # server 默认 "https://index.docker.io/v1/"
  - username: "john"
    password: "12345"
    email: "john@example.com"
    server: "registry.cn-hangzhou.aliyuncs.com"
```

### OIDC 单点登录（多 provider）

```yaml
oidc:
  - name: "sso1"
    enabled: true
    provider_url: "http://127.0.0.1:9001"
    client_id: "sso-xxx"
    client_secret: "xxxx"
    redirect_url: "http://127.0.0.1:3000/auth/callback"
  - name: "sso2"
    enabled: true
    provider_url: "http://127.0.0.1:9001"
    client_id: "sso-xxx"
    client_secret: "xxxx"
    redirect_url: "http://127.0.0.1:3000/auth/callback"
```
