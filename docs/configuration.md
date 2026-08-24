---
title: ⚙️ Configuration
lang: en-US
---

# ⚙️ Configuration

Mars reads a single YAML file for all its settings: `config.yaml` in the directory you run it from. Generate one with `./bin/app init`, then edit it and restart Mars for changes to take effect.

This page explains the settings you're most likely to touch. A full commented template lives at [`config_example.yaml`](https://github.com/duc-cnzj/mars/blob/master/config_example.yaml) in the repository.

## Core settings

| Setting | Default | What it does |
|---|---|---|
| `app_port` | `4000` | The web address users open: `http://<host>:4000` |
| `grpc_port` | `50000` | Internal API port (most users don't touch this) |
| `admin_password` | `123456` | Password for the built-in `admin` account — **change it!** |

## Connecting your Git server

Set `git_server_plugin` so Mars can read your repositories and charts:

```yaml
git_server_plugin:
  name: gitlab
  args:
    token: "your-gitlab-token"
    baseurl: "https://gitlab.com/api/v4"
```

To create a GitLab token: **GitLab → Settings → Access Tokens → Add new token**, and give it at least `read_api` scope. Use your self-hosted GitLab address in `baseurl` if you have one.

## Connecting your Kubernetes cluster

Point Mars at your cluster's kubeconfig. If Mars runs on the same machine as `kubectl`, you usually already have this file:

```yaml
kubeconfig: "/home/you/.kube/config"
```

- Running Mars **inside** the cluster? Leave it empty — Mars uses the in-cluster credentials automatically.
- Running **outside**? Point it at your kubeconfig file.

Mars adds a prefix to every namespace it creates — `ns_prefix` (default `devops-`). You can change it if that prefix clashes with your cluster.

## Sign-in

### Password login

Works out of the box with `admin` and `admin_password`. When users log in with the admin account, they get full permissions (see [Permissions](./access-control.md)).

### Single sign-on (OIDC)

To let people log in with their company account, add one or more OIDC providers:

```yaml
oidc:
  - name: "company-sso"
    enabled: true
    provider_url: "https://sso.company.com"
    client_id: "mars"
    client_secret: "xxxx"
    redirect_url: "http://127.0.0.1:3000/auth/callback"
```

You can configure several providers — users can pick which one to sign in with.

## Private image registry

To pull images from a private Docker registry, list the credentials:

```yaml
imagepullsecrets:
  - username: "registry-user"
    password: "registry-password"
    email: "you@example.com"
    server: "registry.example.com"        # default: https://index.docker.io/v1/
```

Mars attaches these credentials to deployed applications so they can pull private images.

## Storage

Mars stores its own data (namespaces, projects, deploy history) in a database.

| Setting | Default | Notes |
|---|---|---|
| `db_driver` | `sqlite` | `sqlite` needs no setup; `mysql` for production |
| `db_database` | `/tmp/mars-sqlite.db` | file path for sqlite; database name for mysql |
| `db_host` / `db_port` / `db_username` / `db_password` | — | required only for `mysql` |

## File uploads

| Setting | Default | Notes |
|---|---|---|
| `upload_dir` | `/tmp/mars-uploads` | where uploaded files are stored |
| `upload_max_size` | `50m` | maximum upload size (e.g. `100m`, `1g`) |
| `s3_enabled` | `false` | store uploads on S3 / MinIO instead of disk |

## Small niceties

| Setting | Default | What it does |
|---|---|---|
| `picture_plugin` | `picture_bing` | background image on the login page: `picture_bing` or `picture_cartoon` |
| `external_ip` | `127.0.0.1` | externally reachable IP of your cluster |
| `install_timeout` | `90s` | how long a deployment may take before timing out |
| `debug` | `false` | verbose logs for troubleshooting |

## Command line

| Command | What it does |
|---|---|
| `./bin/app init` | generate a `config.yaml` if one doesn't exist |
| `./bin/app serve -c config.yaml` | start the server |
| `./bin/app inspect` | show runtime info (config, plugins, jobs…) |

Common flags for `serve`:

```bash
./bin/app serve -c config.yaml \
  --app_port 4000 \
  --grpc_port 50000 \
  --metrics_port 9091 \
  --kubeconfig ~/.kube/config
```
