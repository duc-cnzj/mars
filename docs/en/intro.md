---
title: 💡 Introduction
lang: en-US
---

# 💡 Introduction

[Mars](https://github.com/duc-cnzj/mars) is an application built for DevOps, running on top of Kubernetes. It deploys an application identical to your production environment within seconds. Mars connects Git, Kubernetes and Helm: images are built through Git CI, then deployed as highly-available applications on Kubernetes — all in one seamless flow.

> In one sentence: **You write the code, we ship it live. Production in 30 seconds.**

## 🗺️ Background

As DevOps takes hold, software development now demands not only efficiency but also effortless deployment — ideally a one-stop pipeline from coding, building, testing to shipping. Mars was born for this. It ties together building, testing and deployment with Git CI/CD, so that anyone — from senior developers to product people who don't write code — can deploy a production-grade application in 30 seconds. Teach once, ship anywhere.

## ✨ Features

- **Binary distribution**: a single binary + `config.yaml` is enough to run, zero external dependencies by default (SQLite + in-memory queue)
- **One-click deploy**: supports any application built on Helm charts
- **Automatic HTTPS domains**: automatic domain and certificate configuration
- **High availability / elastic scaling**: powered by Kubernetes replica scaling
- **Full CLI support**
- **Container terminal / logs / files**: open a terminal, stream logs, and copy files from the browser
- **Resource topology**: real-time resource dependency tree in the cluster
- **Monitoring**: container CPU / memory usage and TopPod ranking
- **Security & audit**: 6-level access control, OIDC SSO, full operation records with playback
- **Pluggable**:
  - Git server: `gitlab`
  - Message queue drivers: `ws_sender_nsq` / `ws_sender_redis` / `ws_sender_memory`
  - Domain / certificate managers: `manual_domain_manager` / `cert-manager_domain_manager` / `sync_secret_domain_manager` / `default_domain_manager`
  - Background pictures: `picture_cartoon` / `picture_bing`
- **Gray release**: cookie-based gray release
- **SDK**: in-repo [api/](https://github.com/duc-cnzj/mars/tree/master/api) module with both gRPC and HTTP/JSON clients

## 🧰 Tech Stack

| Area | Choice |
|---|---|
| Language / Runtime | Go |
| API | gRPC + grpc-gateway (HTTP/JSON) |
| Orchestration | Kubernetes · Helm v3 |
| Data | ent (ORM) · SQLite / MySQL |
| CLI / Config | cobra · viper |
| Frontend | React 18 + Vite 6 + TypeScript + Tailwind CSS v4 (in-repo `frontend/`) |
| Optional deps | NSQ / Redis (messaging) · MinIO / S3 (storage) · Jaeger (tracing) |
