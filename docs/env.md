---
title: 环境变量
lang: zh-cn
---

# 环境变量


| 变量名称            | 介绍                                 |
| ------------------- | ------------------------------------ |
| <.Branch>           | 分支名称，动态: dev/master...        |
| <.Commit>           | commit sha, 6位                      |
| <.Pipeline>         | gitlab 流水线id, github 目前获取不到 |
| <.ClusterIssuer>    | Cert-manager 的issuer                |
| <.Host1-10>            | 域名                                 |
| <.TlsSecret1-10>       | 域名对应的secret                     |
| <.ImagePullSecrets> | 拉取镜像的秘钥                       |



```yaml
imagePullSecrets: <.ImagePullSecrets>
image:
  tag: "<.Branch>-<.Pipeline>"

ingress:
  enabled: false
  annotations: 
    cert-manager.io/cluster-issuer: <.ClusterIssuer>
  hosts:
    - host: <.Host1>
      paths:
        - path: /
          pathType: Prefix
  tls: 
   - secretName: <.Secret1>
     hosts:
       - <.Host1>
```