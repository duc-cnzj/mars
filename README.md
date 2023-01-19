<h1 align="center">Mars</h1>
<div align="center"><img style="width: 100px;height: 100px" src="./frontend/public/logo192.png" /></div>
<p align="center">专为devops而生，30秒内部署一个应用。</p>
<br><br>

<div align="center">

[![codecov](https://codecov.io/gh/duc-cnzj/mars/branch/master/graph/badge.svg?token=EUSLRBT6NN)](https://codecov.io/gh/duc-cnzj/mars) [![unittest](https://github.com/duc-cnzj/mars/actions/workflows/test.yaml/badge.svg)](https://github.com/duc-cnzj/mars/actions/workflows/test.yaml) [![Release](https://img.shields.io/github/release/duc-cnzj/mars.svg)](https://github.com/duc-cnzj/mars/releases/latest) [![GitHub license](https://img.shields.io/github/license/duc-cnzj/mars)](https://github.com/duc-cnzj/mars/blob/master/LICENSE)

</div>

[查看文档](https://youngduc.gitbook.io/mars/)

## 💡 简介

[Mars](https://github.com/duc-cnzj/mars) 是一款专门为 devops 服务的一款应用，基于 kubernetes 之上，可以在短短几秒内部署一个和生产环境一模一样的应用。它打通了 git、kubernetes、helm，通过 git ci 构建镜像，然后通过 kubernetes 部署高可用应用，一气呵成。

## 🗺️ 背景

随着 devops 概念的兴起，现在软件开发不仅要求开发效率高，而且还要求部署便捷，最好能做到流水线开发打包测试上线一条龙服务。
[Mars](https://github.com/duc-cnzj/mars) 由此而生，它打通了打包、测试、部署，基于 git ci/cd 做到任何人不管是开发大牛，还是不懂代码的产品小白，都能在 30 秒部署一个生产级别的应用。真真做到一教即会，高效生产。

## ✨ 特性

- 支持基于 helm charts 开发的任何应用。
- 支持自动配置 https 域名。
- 支持高可用，弹性部署。
- 支持命令行操作。
- 支持查看容器日志。
- 支持查看容器 cpu 和内存使用情况。
- 插件化
  - 队列驱动: ws_sender_nsq, ws_sender_redis, ws_sender_memory
  - 证书驱动: manual_domain_manager, cert-manager_domain_manager, sync_secret_domain_manager
  - 代码仓库支持: gitlab, github
  - 背景图: picture_cartoon，picture_bing
- sdk 接入: [go](https://github.com/duc-cnzj/mars-client)。

## 🛠️ 使用文档

1. 直接去 [release page](https://github.com/duc-cnzj/mars/releases) 下载二进制包

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

web 页面配置项目，开启全局配置。

## 🏗 preview

> [demo source code](https://gitlab.com/duc-cnzj/mars-demo)

[视频教程](https://www.bilibili.com/video/BV19b4y1r7iY/)

## 🍀 go-sdk 接入

```
go get -u github.com/duc-cnzj/mars-client/v4
```

```golang
package main

import (
  client "github.com/duc-cnzj/mars-client/v4"
)

func main()  {
  c, _ := client.NewClient("127.0.0.1:50000",
    client.WithAuth("admin", "123456"),
    client.WithTokenAutoRefresh(),
  )
  defer c.Close()

  // ...
}
```
