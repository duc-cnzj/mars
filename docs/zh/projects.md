---
title: 📦 部署应用
lang: zh-CN
---

# 📦 部署应用

你团队部署的一切都放在 **命名空间** 里，每个应用是一个 **项目**。这一页讲清楚这两者，以及怎么管理版本。

## 命名空间

命名空间是 **给应用分组的地方**——可以理解成一个文件夹，或者某个团队的专属空间。Mars 会为每一个在集群里创建对应的 Kubernetes 命名空间。

- **创建** — 点右上角的 **+**，起个名字，创建。
- **公开 / 私有** — 公开空间任何登录用户都看得到；私有空间只对你添加的人开放（见 [权限管理](./access-control.md)）。
- **收藏** — 把常用空间打个星，好找。
- **转让** — 把命名空间的所有权移给另一个用户（只有创建者能操作）。

## 项目

**项目** 把 Git 仓库和 Helm chart 关联起来，Mars 就知道怎么部署它。

### 创建项目

1. 进命名空间，添加项目。
2. 选 Git 仓库和分支。
3. **设置 chart 目录** — Helm chart 在哪里：
   - 就在本仓库里：直接写文件夹路径（如 `charts`）。
   - 在别的项目里：用 `项目id|分支|路径` 格式，如 `12|main|charts`。

### 配置 chart values

设置好 chart 目录后，Mars 会加载一份默认 `values.yaml` 给你改。它控制镜像 tag、副本数、域名这些。

可以用这些内置变量（用 `<>` 包起来，避免和 Helm 模板语法冲突）：

| 变量 | 会变成什么 |
|---|---|
| `<.Branch>` | 当前 Git 分支 |
| `<.Commit>` | 当前 commit |
| `<.Pipeline>` | GitLab pipeline 号 |
| `<.Host1>` … `<.Host10>` | 你的域名，最多 10 个 |
| `<.TlsSecret1>` … `<.TlsSecret10>` | 对应的 HTTPS 证书 secret |
| `<.ClusterIssuer>` | 证书签发器（用 cert-manager 时）|
| `<.ImagePullSecrets>` | 你私有仓库的凭据 |

示例：

```yaml
image:
  repository: myapp
  tag: "<.Branch>-<.Pipeline>"    # 比如 main-1234

ingress:
  enabled: true
  hosts:
    - host: <.Host1>
  tls:
    - secretName: <.TlsSecret1>
      hosts:
        - <.Host1>
```

**记得先保存 chart 目录再配置 values** —— values 是从那个目录加载的。

### 部署 / 升级

点 **部署** 安装应用（再次部署就是升级，会变成滚动更新）。测试时用 **debug 模式**，某一步失败也继续，方便排查。

### 回滚

每次部署都会存成一个 **版本**。出问题了：

1. 打开项目的 **版本历史**。
2. 找到上一个正常版本。
3. 点 **回滚** —— Mars 会重新部署那个版本。

### 资源拓扑

打开项目的 **资源树**，能看到应用的各个组件（Deployment、Service、Ingress…）之间的依赖关系——排查连接问题时很好用。

## 按分支配置（.mars.yaml）

你也可以在仓库里放一个 `.mars.yaml`，按分支定义设置——chart 位置、默认 values、允许的分支等。这样部署配置就和代码放一起了。

```yaml
# 本仓库里 chart 的位置
local_chart_path: charts
# 界面只显示这些分支
branches:
  - main
  - dev
# 默认 values（和 values.yaml 一样）
values_yaml: |
  replicaCount: 1
  image:
    repository: myapp
    tag: "<.Branch>-<.Pipeline>"
```
