---
title: 快速体验
lang: zh-cn
---

# 安装

::: tip
本页教程为了让用户快速入门，采用使用二进制安装的方法，单机部署，如果要高可用部署，以及的更多功能，请看文档。
:::

## 二进制安装

直接去 [release page](https://github.com/DuC-cnZj/mars/releases) 下载二进制包，然后执行

```bash
./mars init
vim config.yaml

# Output:
# 2023/03/24 16:08:54 创建成功！
```

参数复制这份, 其中 `kubeconfig`、gitlab `token`、`external_ip` 这3项需要配置

```yaml
kubeconfig: "" # TODO 配置集群 kubeconfig 的绝对路径
app_port: 6000
git_server_plugin:
  name: gitlab
  args:
    token: "" # TODO 需要自行创建 gitlab 仓库 read 权限的 token
    baseurl: "https://gitlab.com/api/v4"
ws_sender_plugin:
  name: ws_sender_memory
domain_manager_plugin:
  name: default_domain_manager
picture_plugin:
  name: picture_cartoon
db_driver: "sqlite"
db_database: /tmp/mars-sqlite.db
# TODO 写 nodePort 能访问到的 ip
external_ip: "127.0.0.1"
imagepullsecrets: []
admin_password: "123456"
private_key: |-
  -----BEGIN RSA PRIVATE KEY-----
  MIICWwIBAAKBgQCdx5ZBeL3P3lH2fU/8yd4E1L880DjaKCnnnQkya+kOE7kkJNtP
  xW4WIKsBgXUPtXUYk/uA5AkklJ/1ssiTbkM/G5J54ThsACarhiNijUznD81c7g0Q
  6pbHYGAHU91wQgpcIv39cOKZVpFkEfIwgBMIKUvupBpGyXMU4YALVV23CQIDAQAB
  AoGARo+kzeDumlDlvONr6zRoOybd45eHZWEC5JchLtB9qJL/gH+PKQy1X+X6NDEu
  JflTxcsgdhMFV7u0EdCDzRNJtPKP/cU8hww0J2l3ZKTGzbbQnLIBFD3In8sEc9xe
  3ikEjqs0EgSh3uY5XEq8qzuX3cI+FNlGyOwzM+ZcN7nWfPUCQQDOURX82COQIfAT
  RjTshDQ55J/DUPPHyzpTER9OZNXYKp0IBBNzYyhJ6SHQHSuxHfL8W1FVHhmIsIBW
  GQWo0y7zAkEAw8ZPJ4QH5otMsIgIfwMuPX0rO+QxwmJ6eg9ADuFr5zv6HizjAVVP
  dKXuUU0gnemD4DncgiV2jZ0v2RzHK1aZEwJAR6G7gpgAcPB3jBmaEmwsPdV06rlW
  io2y6FhPiEZWQME62CeiITPSLyc0SC94lfwR+zAxYt4ae2zcgggaAO2hpQJAecA5
  d7S3iRu2XM6sofijaCAQpBV9EItX6dLUHqz4Av0cxmlZ33ljiYKr3CngD/SqS+cQ
  CGwt91H68MXh40TeuwJARxz1VMLq7hKo8J4scAW/YrBTE4N6malYjYoR2HFs+YwL
  cSE/4A4yfzTjN2r5GuJr8rTU7gU4Su9C8dLC0htWCA==
  -----END RSA PRIVATE KEY-----
```

## 启动 mars

```bash
./mars serve

# 打开 localhost:6000
```

账号: `admin`, 密码: `123456`

## 配置

1. [fork 到你自己的 gitlab 仓库](https://gitlab.com/DuC-cnZj/mars-demo), 然后回到 mars
2. mars 页面右上角点击`管理员`
3. 点击`项目配置`进入项目管理页面
4. 搜索 `mars-demo`, 点击`开启`

## 部署 demo

1. 回到主页，点击右上角 + 创建空间
2. 在空间里部署 demo 项目

## 故障排除

> 任何问题请联系 qq: 1025434218