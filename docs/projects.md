---
title: 📦 项目管理
lang: zh-CN
---

# 📦 项目管理

## 命名空间（Namespace）

命名空间是 mars 的**资源隔离单元**，每个项目都归属某个命名空间。mars 管理的命名空间统一带 `ns_prefix` 前缀（默认 `devops-`）。

支持的操作：

| 操作 | 说明 |
|---|---|
| 创建 / 列表 | 在页面点击 `+` 创建；私有空间需成员身份 |
| Transfer | 转移命名空间所有权（仅 owner）|
| UpdatePrivate | 切换公开 / 私有（仅 owner）|
| SyncMembers | 同步成员（仅 owner）|
| Favorite | 收藏命名空间 |
| IsExists | 查询命名空间是否存在（私有空间视同不存在）|

> 权限细节见 [权限模型](./access-control.md)。

## 项目（Project）

项目对应一个 git 仓库 + 一套 helm charts 配置，部署即通过 helm 把项目渲染到 Kubernetes 集群。

### 核心操作

| 操作 | 说明 |
|---|---|
| WebApply / Apply | 基于 helm 一键部署 / 升级应用 |
| Version / 回滚 | 查看版本列表并回滚到历史版本 |
| AllContainers | 列出项目全部容器 |
| ResourceTree | 实时资源拓扑，展示资源依赖树 |
| MemoryCpuAndEndpoints | 查看容器资源占用与访问端点 |

### values 变量

配置 `values.yaml` 时支持内置变量（使用 `<>` 作为定界符，避免和 helm 模板语法冲突）：

| 变量 | 含义 |
|---|---|
| `<.ImagePullSecrets>` | 镜像拉取凭据 |
| `<.Branch>` | 当前分支 |
| `<.Commit>` | 当前 commit |
| `<.Pipeline>` | gitlab pipeline |
| `<.ClusterIssuer>` | 集群签发器 |
| `<.Host1>` … `<.Host10>` | 域名，最多 10 个 |
| `<.TlsSecret1>` … `<.TlsSecret10>` | 对应 TLS 证书 secret |

示例：

```yaml
image:
  repository: xxx
  tag: "<.Branch>-<.Pipeline>"

ingress:
  enabled: true
  annotations:
    cert-manager.io/cluster-issuer: "<.ClusterIssuer>"
  hosts:
    - host: <.Host1>
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: <.TlsSecret1>
      hosts:
        - <.Host1>
```

## 配置方式

### 全局配置（推荐）

在页面的「配置项目」中开启项目，进入**启用全局配置**模式：

1. 首先配置 **charts 目录**
   - charts 就在项目目录下：直接写相对路径
   - 引用其他项目的 charts：按 `项目id|项目分支|相对路径` 格式，如 `12|main|charts`
2. 保存 charts 路径后，会自动加载默认 `values.yaml` 供参考，按提示配置其他字段
3. **配置完记得保存**

### 按分支单独配置（.mars.yaml）

用法借鉴 `.gitlab-ci.yml`，在项目下创建 `.mars.yaml` 即可：

```yaml
# 项目默认的配置文件(可选)
config_file: config.yaml
# 默认配置，必须用 '|'；未设置 config_file 时使用
config_file_values: |
  env: dev
  port: 8000
# 配置文件的类型(有 config_file 时必填)
config_file_type: yaml
# config_field 对应到 helm values.yaml 中的哪个字段(有 config_file 时必填)
# 支持 '->' 指向下一级，如 'config->app_name' 会生成：
#   config:
#     app_name: xxxx
config_field: conf
# charts 文件在项目中存放的目录(必填)，格式同全局配置
local_chart_path: charts
# 是否单字段配置(有 config_file 时必填)
is_simple_env: false
# 若配置则只会显示配置的分支，默认 "*"(可选)
branches:
  - dev
  - master
# 与 helm 的 values.yaml 用法一致，但支持内置变量(见上文)
values_yaml: |
  replicaCount: 1
  image:
    repository: xxx
    pullPolicy: IfNotPresent
    tag: "<.Branch>-<.Pipeline>"
  imagePullSecrets: []
  ingress:
    enabled: true
    hosts:
      - host: <.Host1>
```

#### `is_simple_env` / `config_file` 说明

以一份普通的 helm values.yaml 为例：

```yaml
# 你的 app 的 config 配置：这些都是独立变量 → is_simple_env: false，config_field: conf
conf:
  APP_PORT: 8080
  DB_HOST: mysql
  DB_PORT: 3306

# 这是一个整体 → is_simple_env: true，config_field: conf_two
conf_two: |
  APP_PORT: 8080
  DB_HOST: mysql
  DB_PORT: 3306
```

- `conf` 下是一组独立变量：`is_simple_env` 应为 `false`
- `conf_two` 是整块内容：`is_simple_env` 应为 `true`
- `config_field` 指定这些内容挂在 helm values.yaml 的哪个字段下
