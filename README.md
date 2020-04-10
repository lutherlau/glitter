# Glitter Web Framework

### 快速开始
```go
package main
import "github.com/lutherlau/glitter"

func main(){
    app := glitter.Default()
    app.GET("/hello", func(ctx *glitter.Context){
        ctx.JSON(200, app.JSON{
            "message":"world",
        })
    })
}
```

### 参考
* [Gee](https://geektutu.com/post/gee.html)
* [Gin](https://github.com/gin-gonic/gin)