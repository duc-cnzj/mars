---
title: Websocket 插件
lang: zh-cn
---

# Websocket 插件

## memory (单机)

```yaml
ws_sender_plugin:
  name: ws_sender_memory
```

## redis

```yaml
ws_sender_plugin:
 name: ws_sender_redis
 args:
   addr: 127.0.0.1:6379
   password: ""
   db: 1
```

## nsq (大规模)

```yaml
ws_sender_plugin:
 name: ws_sender_nsq
 args:
   addr: 127.0.0.1:4150
   lookupd_addr: 127.0.0.1:4160
```
