---
title: 📦 Install with Helm
lang: en-US
---

# 📦 Install with Helm

The fastest way to run Mars in a cluster is the official Helm chart — it packages the server, its configuration and (optionally) Redis for you. Two flavors are available:

| Mode | Default | What you get |
|---|---|---|
| **standalone** | ✅ | One replica, SQLite database, in-memory queue. Zero external dependencies — perfect for evaluation and small teams. |
| **cluster** | — | Multiple replicas, MySQL database, Redis-backed message queue. For high availability. |

## Prerequisites

- `kubectl` configured against your cluster
- `helm` v3+
- For **cluster** mode: an external MySQL reachable from inside the cluster

## 1. Add the chart repository

```bash
helm repo add mars-charts https://duc-cnzj.github.io/mars-charts/
helm repo update
helm search repo mars-charts        # should list mars-charts/mars
```

## 2. Prepare your values

Start from the chart's defaults and override what you need:

```bash
helm show values mars-charts/mars > mars-values.yaml
```

The chart turns a `standalone_conf` / `cluster_conf` block into the app's `config.yaml` automatically. At minimum, set the mode and your own admin password:

```yaml
image:
  repository: registry.cn-hangzhou.aliyuncs.com/duc-cnzj/mars
  tag: "latest"              # pin a release instead of latest

mode: "standalone"           # or "cluster"

standalone_conf: |
  external_ip: "1.2.3.4"             # your cluster's externally reachable IP
  admin_password: "change-me"        # 🔒 change the default!
  kubeconfig: ""                     # empty = use the in-cluster config
  git_server_plugin:
    name: gitlab
    args:
      token: "your-gitlab-token"     # GitLab → Settings → Access Tokens
      baseurl: "https://gitlab.com/api/v4"
```

## 3. Install

```bash
helm upgrade --install mars mars-charts/mars \
  -n mars --create-namespace \
  -f mars-values.yaml
```

Watch it come up:

```bash
kubectl get pods -n mars -w
```

## 4. Open the web UI

| Access | values | Result |
|---|---|---|
| **Port-forward** (default `ClusterIP`) | `service.type: ClusterIP` | `kubectl port-forward -n mars svc/mars 4000:6000` → open `http://127.0.0.1:4000` |
| **NodePort** | `service.type: NodePort` | `http://<node-ip>:<node-port>` |
| **LoadBalancer** | `service.type: LoadBalancer` | `http://<lb-ip>` |
| **Ingress** | `ingress.enabled: true` + hosts | `http://your-domain` |

Log in with `admin` and the `admin_password` you set.

## Cluster mode (high availability)

Two replicas, the bundled Redis for the message queue, and an external MySQL for storage:

```yaml
mode: "cluster"
replicaCount: 2

# bundled Redis (bitnami)
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
`cluster_conf` is rendered as a Go template — you can reference release metadata with <code v-pre>{{.Release.Name}}</code> and <code v-pre>{{.Release.Namespace}}</code>, exactly like the `ws_sender_redis` address above.
:::

## Storage

By default the chart mounts an `emptyDir` — the SQLite database and uploads are **lost when the pod restarts**. Keep them with a PVC:

```yaml
persistent:
  enabled: true
  storage: 10Gi
  # storageClassName: your-class
```

If uploads go to S3 instead, persistence can stay off.

## Expose gRPC (optional)

For the SDK / CLI over gRPC:

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

## Canary releases (advanced)

The chart can run a canary Deployment/Service/Ingress alongside the stable one — a new image tag receives a slice of traffic based on nginx-ingress rules (header, cookie or weight). Requires an nginx-ingress controller and `ingress.enabled: true`:

```yaml
canary:
  enabled: true
  image:
    tag: "next"              # new version
  replicas: 1
  ingress:
    byHeader: "x-canary"     # route when this header == "always"
    weight: "10"             # or route 10% of traffic
```

## Upgrade

```bash
helm repo update
helm upgrade --install mars mars-charts/mars -n mars -f mars-values.yaml
```

The pod is recreated when the config checksum changes.

## Security notes

- The chart binds its ServiceAccount to **`cluster-admin`** — Mars needs cluster-wide permissions to deploy apps and watch resources for you. Don't run it in a cluster you don't fully trust.
- Change the default `admin_password` (default `123456`).
- Provide a `private_key` (RSA) — Mars uses it to sign access tokens. The example values include a working one; generate your own.
- Prometheus scraping is enabled by default on **port 9091** (`prometheus.io/scrape: "true"`).
