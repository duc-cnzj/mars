---
title: 安装
lang: zh-cn
---

# 安装

::: tip
安装 mars 使用 二进制安装/helm 安装(推荐)。
:::

## 二进制安装

直接去 [release page](https://github.com/DuC-cnZj/mars/releases) 下载二进制包，然后执行

```bash
mars init
vim config.yaml
```

需要手动配置的参数

```yaml
app_port: 6000
#grpc_port: 50000
debug: false

# use app show tags
# eg: cron,api,metrics
#exclude_server: ""

# 'logrus' | 'zap'
#log_channel: "logrus"

# 开启 git 请求缓存，默认开启，因为 git 请求比较慢，需要缓存
git_server_cached: true

# 'db' or 'memory'
cache_driver: "memory"

git_server_plugin:
  name: gitlab
  args:
    token: ""
    baseurl: "https://gitlab.com/api/v4"
#  name: github
#  args:
#    token: ""
ws_sender_plugin:
  name: ws_sender_memory
#  name: ws_sender_nsq
#  args:
#    addr: 127.0.0.1:4150
#    lookupd_addr 可选，有就用
#    lookupd_addr: 127.0.0.1:4160
#  name: ws_sender_redis
#  args:
#    addr: 127.0.0.1:6379
#    password: ""
#    db: 1
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
picture_plugin:
  name: picture_cartoon
#  name: picture_bing

# kubeconfig 如果在集群外部，则需要配置，否则不用
kubeconfig: ""

# upload_dir 默认是 /tmp/mars-uploads
upload_dir: ""

# 默认 50m, 使用 MB,Gi,m,g这样的写法
#upload_max_size: "50m"

# 文件上传到 s3
s3_endpoint: ""
s3_access_key_id: ""
s3_secret_access_key: ""
s3_use_ssl: false
s3_bucket: mars

# database
# 'sqlite' or 'mysql', 如果是 'sqlite', 'db_database' 为 db 绝对路径
db_driver: "sqlite"
db_database: /tmp/mars-sqlite.db
# 如果是 'mysql' 以下均为必填项，如果是 'sqlite' 就不用填
db_host: 127.0.0.1
db_port: 3306
db_username: root
db_password: ""
db_slow_log_enabled: true
# "ns", "us" (or "µs"), "ms", "s", "m", "h"
db_slow_log_threshold: 200ms

jaeger_agent_host_port: ""

# 集群外网访问 ip
external_ip: "127.0.0.1"

install_timeout: 90s

# imagepullsecrets: docker 私有镜像仓库需要配置相关的账号密码以及仓库地址
# server default: "https://index.docker.io/v1/"
imagepullsecrets:
  - username: "jack"
    password: "12345"
    email: "jack@example.com"
    # server: ""
  - username: "john"
    password: "12345"
    email: "john@example.com"
    server: "registry.cn-hangzhou.aliyuncs.com"

oidc:
 - name: "sso1"
   enabled: true
   provider_url: "http://127.0.0.1:9001"
   client_id: "sso-xxx"
   client_secret: "xxxx"
   redirect_url: "http://127.0.0.1:3000/auth/callback"

admin_password: "123456"
private_key: ""

```

## helm 安装（推荐）

```bash
helm repo add mars-charts https://duc-cnzj.github.io/mars-charts/
# 这里需要自行配置相关参数和上面一样
helm show values mars-charts/mars > mars-values.yamlhelm upgrade --install mars mars-charts/mars -f mars-values.yaml
```