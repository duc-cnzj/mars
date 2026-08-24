---
title: 💡 Introduction
lang: en-US
---

# 💡 Introduction

**Mars** is a platform that deploys and runs your applications for you. You write the code and push it to your Git server; Mars turns it into a running application on Kubernetes — including a domain with HTTPS — in about 30 seconds.

> **You write the code, we ship it live. Production in 30 seconds.**

## Who is it for?

- **Developers** who want to ship without babysitting Kubernetes and Helm details
- **Ops engineers** who manage many applications and environments
- **Product people** who need a quick staging or demo environment — no coding required

## What can you do with it?

| You can… | How |
|---|---|
| Deploy an application | Link a Git repository, click **Deploy** — done |
| Give it a domain | Mars configures the address and HTTPS certificate automatically |
| Debug in the browser | Open a container terminal, watch live logs, copy files |
| Watch resource usage | See CPU and memory per container, plus pod ranking |
| Release a new version | Change the image tag and deploy again — rolling update |
| Go back | Roll back to any previous deployment with one click |
| Manage a team | Namespaces can be public or private, with admin / owner / member roles |
| Stay compliant | Every deployment and operation is recorded; sessions can be replayed |
| Sign in with SSO | Use your company's OIDC single sign-on instead of passwords |

## How it works (in plain words)

- Your app is described by a **Helm chart** — a standard, portable way to define an application.
- Mars stores your **namespace** (a collection of apps) and **project** (your Git repo + chart configuration).
- When you click **Deploy**, Mars builds the app from your chart and starts it in your **Kubernetes** cluster.

You don't need to master Kubernetes or Helm to use it — the web interface handles that for you.

## Not sure where to start?

Head over to the [Quickstart](./quickstart.md) — it walks you through installing Mars and deploying your first app.
