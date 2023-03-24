---
title: ingress 域名插件
lang: zh-cn
---

# ingress 域名插件

## 同步指定 namespace secret(推荐)

```yaml
domain_manager_plugin:
 name: sync_secret_domain_manager
 args:
   wildcard_domain: "*.mars.local"
   secret_name: ""
   secret_namespace: ""
```

## 手动配置注入证书

```yaml
domain_manager_plugin:
 name: manual_domain_manager
 args:
   wildcard_domain: "*.mars.local"
   tls_key: ""
   tls_crt: ""
```

## cert-manager issue 插件(不推荐)

```yaml
domain_manager_plugin:
 name: cert-manager_domain_manager
 args:
   wildcard_domain: "*.mars.local"
   cluster_issuer: "letsencrypt-mars"
```

## default_domain_manager (fake domain)

```yaml
domain_manager_plugin:
  name: default_domain_manager
```
