---
title: 配置项目
lang: zh-cn
---

# 配置项目

## 全局配置（推荐）

登录到页面，点击 配置项目->开启项目->启用全局配置
![配置项目](./images/config1.png)
![开启项目](./images/config2.png)

首先配置 charts 目录

> 如果charts 就在项目目录下可以直接写相对路径
> 如果是引用别的的项目的charts，可以按照这个格式写 "项目id|项目分支|相对路径"

![首先配置 charts 路径，然后保存](./images/config3.png)

::: warning
配置完记得保存
:::

配好 charts 后保存，会自动加载默认 `values.yaml` 文件，这个只是给你参考用的，然后按照提示配置玩其他字段，其中 `values.yaml` , 有内置变量可以使用，配置完后大概长下面这样

![配置完其他字段](./images/config4.png)