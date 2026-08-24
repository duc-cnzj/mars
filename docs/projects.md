---
title: 📦 Deploying Apps
lang: en-US
---

# 📦 Deploying Apps

Everything your team deploys lives in a **namespace**, and each application is a **project**. This page explains both and how to manage versions.

## Namespaces

A namespace is a **place to group your applications** — think of it as a folder or a team's space. Mars creates one real Kubernetes namespace for each of these.

- **Create** — click the **+** button in the top-right corner, give it a name, and create it.
- **Public vs private** — public namespaces are visible to anyone logged in; private ones only to people you add (see [Permissions](./access-control.md)).
- **Favorite** — star a namespace so it's easy to find.
- **Transfer** — hand ownership of a namespace to another user (only the owner can do this).

## Projects

A **project** ties a Git repository to a Helm chart so Mars knows how to deploy it.

### Create a project

1. Open your namespace and add a project.
2. Choose the Git repository and branch.
3. **Set the chart directory** — where the Helm chart lives:
   - In this repository: just write the folder path (e.g. `charts`).
   - In another project: use `project-id|branch|path`, e.g. `12|main|charts`.

### Configure the chart values

After you set the chart directory, Mars loads a default `values.yaml` that you can edit. This file controls things like the image tag, replicas and domain.

You can use these built-in variables (surrounded by `<>`, so they don't clash with Helm's own templates):

| Variable | What it becomes |
|---|---|
| `<.Branch>` | the current Git branch |
| `<.Commit>` | the current commit |
| `<.Pipeline>` | the GitLab pipeline number |
| `<.Host1>` … `<.Host10>` | your domain(s), up to 10 |
| `<.TlsSecret1>` … `<.TlsSecret10>` | the matching HTTPS certificate secret |
| `<.ClusterIssuer>` | the certificate issuer (if you use cert-manager) |
| `<.ImagePullSecrets>` | your private-registry credentials |

Example:

```yaml
image:
  repository: myapp
  tag: "<.Branch>-<.Pipeline>"    # e.g. main-1234

ingress:
  enabled: true
  hosts:
    - host: <.Host1>
  tls:
    - secretName: <.TlsSecret1>
      hosts:
        - <.Host1>
```

**Remember to save** your chart directory before configuring values — the values are loaded from that directory.

### Deploy / upgrade

Click **Deploy** to install the app (or to update it — a new deployment becomes a rolling update). While testing, use **debug mode** so the deployment continues even if a step fails.

### Roll back

Every deployment is saved as a **version**. If something goes wrong:

1. Open the project's **version history**.
2. Find the last good version.
3. Click **roll back** — Mars reinstalls that version.

### Resource topology

Open the project's **resource tree** to see how the app's components (deployments, services, ingresses…) depend on each other — a handy overview when something isn't connecting.

## Per-branch settings (.mars.yaml)

You can also put a `.mars.yaml` file in the repository to define settings per branch — chart location, default values, allowed branches, and so on. This keeps deployment config close to the code.

```yaml
# where the chart lives in this repository
local_chart_path: charts
# only show these branches in the UI
branches:
  - main
  - dev
# default values (like a values.yaml)
values_yaml: |
  replicaCount: 1
  image:
    repository: myapp
    tag: "<.Branch>-<.Pipeline>"
```
