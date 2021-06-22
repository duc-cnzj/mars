<h1 align="center">Mars</h1>
<p align="center">专为devops而生，10秒内部署一个应用。</p>
<br><br>

## 💡 简介

[Mars](https://github.com/DuC-cnZj/mars) 是一款专门为devops服务的一款应用，基于 kubernetes 之上，可以在短短几秒内部署一个和生产环境一模一样的应用。它打通了 gitlab、kubernetes、helm，通过 gitlab ci 构建镜像，然后通过kubernetes 部署高可用应用，一气呵成。

## 🗺️ 背景

随着 devops 概念的兴起，现在软件开发不仅要求开发效率高，而且还要求部署便捷，最好能做到流水线开发打包测试上线一条龙服务。
[Mars](https://github.com/DuC-cnZj/mars) 由此而生，它打通了打包、测试、部署，基于 gitlab ci/cd 做到任何人不管是开发大牛，还是不懂代码的产品小白，都能在10秒部署一个生产级别的应用。真真做到一教即会，高效生产。

## ✨  特性

* 支持基于 helm charts 开发的任何应用。
* 支持自动配置 https 域名。
* 支持高可用，弹性部署。
* 支持命令行操作。
* 支持查看容器日志。
* 支持查看容器cpu和内存使用情况。

## 🛠️ 使用文档

1. 直接去 [release page](https://github.com/DuC-cnZj/mars/releases) 下载二进制包

初始化配置
```bash
mars init
```

2. 在 kubernetes 内部署（推荐）

```bash
helm repo add mars-charts https://duc-cnzj.github.io/mars-charts/
# 这里需要自行配置相关参数
helm show values mars-charts/mars > mars-values.yaml
helm upgrade --install mars mars-charts/mars -f mars-values.yaml
```

## 🔍 configuration

用法借鉴 `.gitlab.yml`, 使用时只需要在项目下面创建一个 `.mars.yaml` 就可以了。 

`.mars.yaml` 配置参考：

```yaml
# 项目默认的配置文件
config_file: config.yaml
# 配置文件的类型
config_file_type: yaml
# config_file 对应到 helm values.yaml 中的哪个字段
config_field: conf
# 镜像仓库
docker_repository: nginx
# tag 可以使用的变量有 {{.Commit}} {{.Branch}} {{.Pipeline}}
docker_tag_format: "{{.Branch}}-{{.Pipeline}}"
# charts 文件在项目中存放的目录
local_chart_path: charts
# 是不是单字段的配置
is_simple_env: false
# values.yaml 会合并其他配置
default_values:
  redis:
    enabled: true
    cluster: 
      slaveCount: 0
    usePassword: false
  service:
    type: ClusterIP
  ingess:
    enabled: false
# 若配置则只会显示配置的分支, 默认 "*", 可选 
branches:
- dev
- master
```

### 📒 `is_simple_env`, `config_file` 解释

这是一份普通的 helm charts values.yaml 文件
```yaml
# Default values for charts.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1

image:
  repository: nginx
  pullPolicy: IfNotPresent
  tag: ""

# ... 省略

# 你的 app 的 config 配置应该是这样的, 这个 `conf` 字段会被你用到其他地方比如 configmap、secret 等等
# 下面这个你的 is_simple_env 应该是 false，因为他们都是单独的变量
# config_file 字段的值是 `conf`
conf:
  APP_PORT: 8080
  DB_HOST: mysql
  DB_PORT: 3306
#...

# 下面这个你的 is_simple_env 应该是 true，因为这部分配置是一个整体, config_file 字段的值是 `conf_two` 
conf_two: |
  APP_PORT: 8080
  DB_HOST: mysql
  DB_PORT: 3306
```

## 🏗 preview

> xuanji golang 版本。

https://github.com/Lick-Dog-Club/xuanji-k8s-all-in-one


## TODO

- ui 美化
- 配置管理界面可以直接查看 `.mars.yaml` 配置文件
- grpc-gateway 替换 gin？