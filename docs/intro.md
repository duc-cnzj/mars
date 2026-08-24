---
title: 💡 简介
lang: zh-CN
---

# 💡 简介

[Mars](https://github.com/duc-cnzj/mars) 是一款专门为 devops 服务的应用，基于 Kubernetes 之上，可以在短短几十秒内部署一个和生产环境一模一样的应用。它打通了 gitlab、kubernetes、helm：通过 gitlab ci 构建镜像，再由 Kubernetes 高可用部署，一气呵成。

> 一句话：**You write the code, we ship it live. Production in 30 seconds.**

## 🗺️ 背景

随着 devops 概念的兴起，软件开发不仅要求开发效率高，还要求部署便捷，最好能做到「流水线开发、打包、测试、上线」一条龙。Mars 由此而生：它把打包、测试、部署打通，基于 gitlab ci/cd，无论是开发大牛还是不懂代码的产品小白，都能在 30 秒内部署一个生产级别的应用，真正做到「一教即会」。

## ✨ 特性

- **二进制部署**：单二进制 + `config.yaml` 即可运行，默认零外部依赖（SQLite + 内存队列）
- **一键部署**：支持基于 helm charts 开发的任何应用
- **自动 HTTPS 域名**：支持自动配置域名与证书
- **高可用 / 弹性部署**：基于 Kubernetes 副本弹性伸缩
- **命令行操作**：完整的 CLI 支持
- **容器终端 / 日志 / 文件**：网页内直接打开终端、查看日志、拷贝文件
- **资源拓扑**：实时展示项目在集群中的资源依赖树
- **监控**：查看容器 CPU、内存使用，TopPod 排行
- **安全与审计**：6 级权限模型、OIDC 单点登录、操作全程留痕可回放
- **插件化**：
  - Git 服务端：`gitlab`
  - 消息队列驱动：`ws_sender_nsq` / `ws_sender_redis` / `ws_sender_memory`
  - 域名/证书管理：`manual_domain_manager` / `cert-manager_domain_manager` / `sync_secret_domain_manager` / `default_domain_manager`
  - 背景图：`picture_cartoon` / `picture_bing`
- **灰度发布**：基于 cookie 的灰度发布能力
- **SDK**：仓内 [api/](https://github.com/duc-cnzj/mars/tree/master/api) 模块，同时提供 gRPC 与 HTTP/JSON 客户端

## 🧰 技术栈

| 领域 | 选型 |
|---|---|
| 语言 / 运行时 | Go |
| API | gRPC + grpc-gateway（HTTP/JSON）|
| 编排 | Kubernetes · Helm v3 |
| 数据 | ent（ORM）· SQLite / MySQL |
| CLI / 配置 | cobra · viper |
| 前端 | React 18 + Vite 6 + TypeScript + Tailwind CSS v4（仓内 `frontend/`）|
| 可选依赖 | NSQ / Redis（消息）· MinIO / S3（文件存储）· Jaeger（链路追踪）|
