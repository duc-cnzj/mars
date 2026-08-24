---
title: 🚀 Quickstart
lang: en-US
---

# 🚀 Quickstart

::: tip
The default configuration (SQLite storage + in-memory message queue) has **zero external dependencies** — start it up first, then connect a real cluster.
:::

## Prerequisites

- A Linux / macOS host (use WSL on Windows)
- Optional: a Kubernetes cluster (required to actually deploy apps) + a GitLab

## Install & Run

### 1. Build

```bash
make build        # produces ./bin/app
```

No Go toolchain? Run straight from source: `go run main.go serve -c config.yaml`.

Or download a binary from the [Release page](https://github.com/DuC-cnZj/mars/releases).

### 2. Generate the default config

```bash
./bin/app init    # generates config.yaml (skipped if it already exists)
```

### 3. Start the server

```bash
./bin/app serve -c config.yaml
```

Then open the web UI:

- **Address**: `http://127.0.0.1:4000`
- **Default account**: `admin` / `123456`

### 4. Connect a real cluster

Zero-dependency mode is only for trying things out. To actually deploy applications into a Kubernetes cluster, configure two things (see [Configuration](./configuration.md)):

- `git_server_plugin`: fill in your GitLab `token` and `baseurl`
- `kubeconfig`: point to your kubeconfig file when running mars outside the cluster

### 5. (Optional) Infrastructure

When you need NSQ/Redis for messaging, MySQL for storage, or MinIO, [dev/docker-compose.yml](https://github.com/duc-cnzj/mars/blob/master/dev/docker-compose.yml) provides these dependencies (infrastructure only, not mars itself):

```bash
make dev-up                              # docker compose -f dev/docker-compose.yml up -d
# single service: make dc-up SVC=redis
# teardown:       make dev-down / make dc-down SVC=redis
```

## Deploy Your First App

The steps below walk through the web UI (all text — the actual UI may vary by version):

1. **Log in**: open `http://127.0.0.1:4000`, sign in with `admin` / `123456`
2. **Create a namespace**: click the `+` button in the top-right, fill in the namespace config and create it
3. **Configure the project**: open the project and configure the charts directory
   - Charts live in the project repo: write the relative path directly
   - Charts come from another project: use the `project-id|branch|relative-path` format
4. **Configure values**: after saving the charts path, the default `values.yaml` is loaded for reference; fill in the remaining fields (see [Projects](./projects.md))
5. **Deploy**: click deploy. `debug` mode means `atomic = false` in Helm
6. **Verify**: after deployment, check container status, open logs, and visit the app

Once deployed, try **overriding config**: modify the image tag and deploy again to roll out a new version.
