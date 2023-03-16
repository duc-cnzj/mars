---
title: 💡 简介
lang: zh-cn
---

# 💡 简介

[Mars](https://github.com/duc-cnzj/mars) 是一款专门为devops服务的一款应用，基于 kubernetes 之上，可以在短短几秒内部署一个和生产环境一模一样的应用。它打通了 gitlab、kubernetes、helm，通过 gitlab ci 构建镜像，然后通过kubernetes 部署高可用应用，一气呵成。

## 🗺️ 背景
随着 devops 概念的兴起，现在软件开发不仅要求开发效率高，而且还要求部署便捷，最好能做到流水线开发打包测试上线一条龙服务。 Mars 由此而生，它打通了打包、测试、部署，基于 gitlab ci/cd 做到任何人不管是开发大牛，还是不懂代码的产品小白，都能在30秒部署一个生产级别的应用。真真做到一教即会，高效生产。

## ✨ 特性

- 二进制部署
- 支持基于 helm charts 开发的任何应用。
- 支持自动配置 https 域名。
- 支持高可用，弹性部署。
- 支持命令行操作。
- 支持查看容器日志。
- 支持查看容器cpu和内存使用情况。
- 插件化
- 队列驱动: nsq, redis, memory。
- sdk 接入: go-sdk。