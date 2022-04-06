<h1 align="center">Mars</h1>
<p align="center">专为devops而生，30秒内部署一个应用。</p>
<br><br>

[查看文档](https://youngduc.gitbook.io/mars/)

## 💡 简介

[Mars](https://github.com/duc-cnzj/mars) 是一款专门为devops服务的一款应用，基于 kubernetes 之上，可以在短短几秒内部署一个和生产环境一模一样的应用。它打通了 git、kubernetes、helm，通过 git ci 构建镜像，然后通过kubernetes 部署高可用应用，一气呵成。

## 🗺️ 背景

随着 devops 概念的兴起，现在软件开发不仅要求开发效率高，而且还要求部署便捷，最好能做到流水线开发打包测试上线一条龙服务。
[Mars](https://github.com/duc-cnzj/mars) 由此而生，它打通了打包、测试、部署，基于 git ci/cd 做到任何人不管是开发大牛，还是不懂代码的产品小白，都能在30秒部署一个生产级别的应用。真真做到一教即会，高效生产。

## ✨  特性

* 支持基于 helm charts 开发的任何应用。
* 支持自动配置 https 域名。
* 支持高可用，弹性部署。
* 支持命令行操作。
* 支持查看容器日志。
* 支持查看容器cpu和内存使用情况。
* 插件化
  * 队列驱动: ws_sender_nsq, ws_sender_redis, ws_sender_memory
  * 证书驱动: manual_domain_manager, cert-manager_domain_manager
  * 代码仓库支持: gitlab, github
  * 背景图: picture_cartoon，picture_bing
* sdk 接入: [go](https://github.com/duc-cnzj/mars-client)。

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

## TODO

- [ ] 国际化
- [ ] ratelimiter
- [ ] 外部接口调用优化
- [ ] namespace all -> list
- [ ] grpc 可配置使用 tls
- ~~[ ] 增加 basic? or CA? auth，参考 k8s 的做法~~
- [x] ~~redis 不想强依赖 redis~~ db cache, 最后使用了 DB cache
- [x] gitlab 接口缓存优化，commit 接口有些值都是固定的可以做缓存
- [x] git server cache
- [x] 所有 gitlab 都改成 git
- [x] 通过 ci 发布客户端
- [x] export/import 配置文件
- [x] 重构所有表单，增加表单验证
- [x] 自定义额外字段，组合模式(前端就算了，连类都不用了，都是 FC)
- [x] client 集成 copy to pod & uploader
- [x] rpc 增加远程执行容器命令接口
- [x] 接口验证
- [x] 前端 namespace 页面 margin-bottom
- [x] 缺一个 project list
- [x] 打开modal无法下滑页面的问题 `ant-scrolling-effect` overflow: hidden 引起的，从 modal click 给 body 加 class 入手解决
- [x] c.GitServer().ProjectList 不应该叫list，因为拿到的是全部，要叫 all
- [x] add current metrics
- [x] 增加修改记录，能清楚的记录谁在什么时候修改了什么
- [x] ui 美化
- [x] 引入 values 字段替换掉之前的 DockerRepository、DockerTagFormat、IngressOverwriteValues。
- [x] ws 部分也是用 proto 定义，input 和 response 都通过 proto。
- [x] 插件化 ingress/tls 证书的注入方式
- [x] opentracing
- [x] shell 自适应高度
- [x] gin -> grpc ？
- [x] grpc-gateway 替换 gin, branch: `grpc`
- [x] 重构 ui 创建项目部分代码
- [x] 重构后端部署部分代码
- [x] 多容器还没写
- [x] sessionId 还是要用起来
- [x] 前端shell退出后，后端对应的 goroutine 也要退出，防止泄漏
- [x] 高可用化
- ~~socket install 方法剥离出来~~
- ~~分离 ResponseMetadata 中的 Data~~
- ~~断开连接使用图标的方式~~
- ~~配置可启动之后再添加~~