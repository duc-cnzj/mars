---
title: GO-SDK
lang: zh-cn
---

# GO-SDK

## 🍀 go-sdk 接入

```bash
go get -u github.com/duc-cnzj/mars-client/v4
```

```go
package main

import (
  client "github.com/duc-cnzj/mars-client/v3"
)

func main()  {
  c, err := client.NewClient("127.0.0.1:50000",
    client.WithAuth("admin", "123456"),
    client.WithTokenAutoRefresh(),
  )
  defer c.Close()
}
```