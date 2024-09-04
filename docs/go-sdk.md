---
title: GO-SDK
lang: zh-cn
---

# GO-SDK

## 🍀 go-sdk 接入

```bash
go get -u github.com/duc-cnzj/mars/api/v5
```

```go
package main

import (
  api "github.com/duc-cnzj/mars/api/v5"
)

func main()  {
  c, _ := api.NewClient("127.0.0.1:50000",
    api.WithAuth("admin", "123456"),
    api.WithTokenAutoRefresh(),
  )
  defer c.Close()

  // ...
}
```