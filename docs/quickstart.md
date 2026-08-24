---
title: 🚀 Quickstart
lang: en-US
---

# 🚀 Quickstart

This page gets you from zero to a running Mars instance with your first application deployed.

::: tip
Mars starts with **zero external dependencies** by default (SQLite + in-memory queue). You can try it out right away, and connect your real cluster and Git server when you're ready.
:::

## What you need

| Requirement | Needed for | Notes |
|---|---|---|
| A Linux / macOS machine (or WSL on Windows) | running Mars | any small server or even a laptop |
| A Kubernetes cluster | actually deploying apps | a test cluster (e.g. kind / minikube) is fine to start |
| A Git server with a token | pulling your code and charts | e.g. GitLab — you'll need an access token |

## Step 1 — Get Mars

Pick one:

```bash
# Option A: build from source
make build                      # produces ./bin/app
```

```bash
# Option B: download a binary
# grab the latest release from https://github.com/DuC-cnZj/mars/releases
```

## Step 2 — Generate the config

```bash
./bin/app init                  # creates config.yaml (skipped if it already exists)
```

## Step 3 — Connect your Git server and cluster

Open `config.yaml` and fill in two things (see [Configuration](./configuration.md) for details):

```yaml
git_server_plugin:
  name: gitlab
  args:
    token: "your-gitlab-token"          # from GitLab → Settings → Access Tokens
    baseurl: "https://gitlab.com/api/v4"

kubeconfig: "/home/you/.kube/config"    # point to your cluster's kubeconfig
```

::: tip
Skip this step for a first try — Mars runs with the default in-memory settings.
:::

## Step 4 — Start Mars

```bash
./bin/app serve -c config.yaml
```

Open your browser:

- **Address**: `http://127.0.0.1:4000`
- **Default account**: `admin` / `123456`

## Step 5 — Deploy your first app

1. **Create a namespace** — click the **+** button in the top-right corner. A namespace is just a place to group your apps. (Leave it public for now.)
2. **Open the project** — click your namespace, then create a project for your Git repository.
3. **Set the chart directory** — tell Mars where your Helm chart lives in the repo. If your chart is at the repo root, enter the folder name (e.g. `charts`). If it's in another project, use `project-id|branch|path`.
4. **Fill in the values** — Mars shows a ready-made `values.yaml` you can edit. You can use built-in variables like `<.Branch>` for the branch name and `<.Host1>` for your domain. See [Deploying Apps](./projects.md).
5. **Click Deploy** — watch it roll out. Debug mode keeps going even if a step fails, which is handy while testing.
6. **Verify** — open the app's logs, or click the access link once the domain is ready.

That's it — your app is live. To release a new version, change the image tag in the project and deploy again.

## Next steps

- [Configuration](./configuration.md) — set up SSO, private image registry, storage and more
- [Deploying Apps](./projects.md) — namespaces, values variables, upgrades and rollbacks
- [Containers & Logs](./containers.md) — terminals, logs and file operations
