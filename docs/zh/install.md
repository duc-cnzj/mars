---
title: 📦 Helm 部署
lang: zh-CN
---

# 📦 Helm 部署

在集群里跑 Mars，最快的方式是官方 Helm Chart——它把服务端、配置、（可选）Redis 一起打包好了。有两种模式：

| 模式 | 默认 | 你能得到什么 |
|---|---|---|
| **standalone** | ✅ | 单副本，SQLite 数据库，内存队列。零外部依赖——适合试用和小团队。 |
| **cluster** | — | 多副本，MySQL 数据库，Redis 消息队列。适合高可用。 |

## 前置条件

- 已配置好 `kubectl` 的集群
- `helm` v3+
- **cluster 模式**需要一个集群内可达的外部 MySQL

## 1. 添加仓库

```bash
helm repo add mars-charts https://duc-cnzj.github.io/mars-charts/
helm repo update
helm search repo mars-charts        # 应能看到 mars-charts/mars
```

## 2. 准备 values

先从 chart 默认值开始，再覆盖你需要改的部分：

```bash
helm show values mars-charts/mars > mars-values.yaml
```

chart 会把 `standalone_conf` / `cluster_conf` 自动转成应用的 `config.yaml`。至少设置模式和你的 admin 密码：

```yaml
image:
  repository: registry.cn-hangzhou.aliyuncs.com/duc-cnzj/mars
  tag: "latest"              # 建议固定到某个发布版本，别用 latest

mode: "standalone"           # 或 "cluster"

standalone_conf: |
  external_ip: "1.2.3.4"             # 集群对外可达的 IP
  admin_password: "change-me"        # 🔒 改掉默认密码！
  kubeconfig: ""                     # 留空 = 使用集群内配置
  git_server_plugin:
    name: gitlab
    args:
      token: "your-gitlab-token"     # GitLab → Settings → Access Tokens
      baseurl: "https://gitlab.com/api/v4"
```

## 3. 安装

```bash
helm upgrade --install mars mars-charts/mars \
  -n mars --create-namespace \
  -f mars-values.yaml
```

等它起来：

```bash
kubectl get pods -n mars -w
```

## 4. 打开网页

| 方式 | values | 结果 |
|---|---|---|
| **端口转发**（默认 `ClusterIP`）| `service.type: ClusterIP` | `kubectl port-forward -n mars svc/mars 4000:6000` → 打开 `http://127.0.0.1:4000` |
| **NodePort** | `service.type: NodePort` | `http://<节点ip>:<节点端口>` |
| **LoadBalancer** | `service.type: LoadBalancer` | `http://<负载均衡ip>` |
| **Ingress** | `ingress.enabled: true` + hosts | `http://你的域名` |

用 `admin` 和你设置的 `admin_password` 登录。

## 集群模式（高可用）

两个副本，用内置 Redis 做消息队列，外部 MySQL 做存储：

```yaml
mode: "cluster"
replicaCount: 2

# 内置 Redis（bitnami）
redis:
  enabled: true
  auth:
    password: "mars-redis-password"
  master:
    persistence:
      enabled: false
  replica:
    replicaCount: 0

cluster_conf: |
  db_driver: 'mysql'
  db_host: your-mysql-host
  db_port: 3306
  db_username: root
  db_password: ""
  db_database: mars

  ws_sender_plugin:
    name: ws_sender_redis
    args:
      addr: "{{.Release.Name}}-redis-master.{{.Release.Namespace}}.svc.cluster.local:6379"
      password: "mars-redis-password"
      db: 1
```

::: tip
`cluster_conf` 会被当作 Go 模板渲染——可以用 <code v-pre>{{.Release.Name}}</code>、<code v-pre>{{.Release.Namespace}}</code> 引用发布元数据，就像上面 `ws_sender_redis` 的地址一样。
:::

## 存储

默认 chart 挂的是 `emptyDir`——SQLite 数据库和上传文件在 **Pod 重启后会丢**。要保住它们，开 PVC：

```yaml
persistent:
  enabled: true
  storage: 10Gi
  # storageClassName: your-class
```

如果上传文件走 S3，持久化也可以不开。

## 暴露 gRPC（可选）

给 SDK / CLI 走 gRPC 用：

```yaml
grpc:
  enabled: true
  service:
    enabled: true
    type: ClusterIP
  ingress:
    enabled: true
    annotations:
      nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
    hosts:
      - host: grpc.mars.example.com
```

## 金丝雀发布（进阶）

chart 可以在稳定版本旁边再跑一个金丝雀 Deployment/Service/Ingress——新版本镜像按 nginx-ingress 规则（header、cookie 或权重）分到部分流量。依赖 nginx-ingress controller，且需要 `ingress.enabled: true`：

```yaml
canary:
  enabled: true
  image:
    tag: "next"              # 新版本
  replicas: 1
  ingress:
    byHeader: "x-canary"     # 请求头等于 "always" 时走金丝雀
    weight: "10"             # 或者分 10% 流量
```

## 升级

```bash
helm repo update
helm upgrade --install mars mars-charts/mars -n mars -f mars-values.yaml
```

配置 checksum 变化时 Pod 会被重建。

## 安全注意

- chart 会把 ServiceAccount 绑定到 **`cluster-admin`**——Mars 需要集群级权限来替你部署应用、监听资源。不要装在你完全不信赖的集群上。
- 改掉默认 `admin_password`（默认 `123456`）。
- 提供 `private_key`（RSA）——Mars 用它给访问 token 签名。示例 values 里有一个可用的，自己部署建议生成自己的。
- 默认开启 Prometheus 采集，端口 **9091**（`prometheus.io/scrape: "true"`）。
