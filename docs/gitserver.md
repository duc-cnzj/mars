---
title: Git 仓库
lang: zh-cn
---

# Git 仓库

## gitlab(推荐)

```yaml
git_server_plugin:
  name: gitlab
  args:
    token: ""
    baseurl: "https://gitlab.com/api/v4"
```

## ~~github~~ v5+ 废弃

> github 接口不太好用，集成较差，无法使用 `pipeline` 变量

```yaml
git_server_plugin:
  name: github
  args:
    token: ""
```