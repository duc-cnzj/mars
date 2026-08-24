---
title: 📦 Projects
lang: en-US
---

# 📦 Projects

## Namespaces

A namespace is mars' **resource isolation unit** — every project belongs to a namespace. All namespaces managed by mars carry the `ns_prefix` prefix (default `devops-`).

Supported operations:

| Operation | Description |
|---|---|
| Create / List | click `+` on the page to create; private namespaces require membership |
| Transfer | transfer namespace ownership (owner only) |
| UpdatePrivate | toggle public / private (owner only) |
| SyncMembers | sync members (owner only) |
| Favorite | favorite a namespace |
| IsExists | check if a namespace exists (private ones appear as non-existent) |

> See [Access Control](./access-control.md) for permission details.

## Projects

A project maps to a git repository + a set of Helm chart configs. Deploying renders the project into the Kubernetes cluster via Helm.

### Core Operations

| Operation | Description |
|---|---|
| WebApply / Apply | one-click deploy / upgrade an app via Helm |
| Version / Rollback | list versions and roll back to a historical one |
| AllContainers | list all containers of the project |
| ResourceTree | real-time resource topology showing the resource dependency tree |
| MemoryCpuAndEndpoints | container resource usage and access endpoints |

### Values Variables

The `values.yaml` you configure supports built-in variables (using `<>` as delimiters to avoid clashing with Helm template syntax):

| Variable | Meaning |
|---|---|
| `<.ImagePullSecrets>` | image pull secrets |
| `<.Branch>` | current branch |
| `<.Commit>` | current commit |
| `<.Pipeline>` | GitLab pipeline |
| `<.ClusterIssuer>` | cluster issuer |
| `<.Host1>` … `<.Host10>` | hostnames, up to 10 |
| `<.TlsSecret1>` … `<.TlsSecret10>` | matching TLS secrets |

Example:

```yaml
image:
  repository: xxx
  tag: "<.Branch>-<.Pipeline>"

ingress:
  enabled: true
  annotations:
    cert-manager.io/cluster-issuer: "<.ClusterIssuer>"
  hosts:
    - host: <.Host1>
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: <.TlsSecret1>
      hosts:
        - <.Host1>
```

## Configuration

### Global Configuration (recommended)

In "Configure project", enable the project and switch to **global configuration** mode:

1. First configure the **charts directory**
   - Charts live in the project repo: write the relative path directly
   - Charts come from another project: use the `project-id|branch|relative-path` format, e.g. `12|main|charts`
2. After saving the charts path, the default `values.yaml` is loaded for reference; fill in the remaining fields
3. **Remember to save when done**

### Per-branch Configuration (.mars.yaml)

Modeled after `.gitlab-ci.yml`; just create a `.mars.yaml` in the project:

```yaml
# default config file of the project (optional)
config_file: config.yaml
# default config, must use '|'; used when config_file is not set
config_file_values: |
  env: dev
  port: 8000
# config file type (required when config_file is set)
config_file_type: yaml
# which field in helm values.yaml config_field maps to (required when config_file is set)
# '->' points to a nested level, e.g. 'config->app_name' produces:
#   config:
#     app_name: xxxx
config_field: conf
# directory holding the charts in the project (required), same format as global config
local_chart_path: charts
# whether it is a single-field config (required when config_file is set)
is_simple_env: false
# if set, only these branches are shown, default "*" (optional)
branches:
  - dev
  - master
# behaves like helm values.yaml but supports built-in variables (see above)
values_yaml: |
  replicaCount: 1
  image:
    repository: xxx
    pullPolicy: IfNotPresent
    tag: "<.Branch>-<.Pipeline>"
  imagePullSecrets: []
  ingress:
    enabled: true
    hosts:
      - host: <.Host1>
```

#### `is_simple_env` / `config_file` explained

Using a plain Helm values.yaml as an example:

```yaml
# your app's config: these are individual variables → is_simple_env: false, config_field: conf
conf:
  APP_PORT: 8080
  DB_HOST: mysql
  DB_PORT: 3306

# this is one whole block → is_simple_env: true, config_field: conf_two
conf_two: |
  APP_PORT: 8080
  DB_HOST: mysql
  DB_PORT: 3306
```

- `conf` holds individual variables: `is_simple_env` should be `false`
- `conf_two` is a single block: `is_simple_env` should be `true`
- `config_field` decides which field of the Helm values.yaml these configs land under
