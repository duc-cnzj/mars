# 🍀 go-sdk 接入

```
go get -u github.com/duc-cnzj/mars/api/v5
```

```golang
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


更多使用方法请参考 [examples](https://github.com/duc-cnzj/mars/tree/master/examples)