# Make the Beego Route Code Navigatable

## Migration

### old source file:
```go
package routers

import (
	"controllers"
	"github.com/astaxie/beego"
)

func register() {
	beego.NewNamespace("/api",
		beego.NSNamespace("/v1",
			beego.NSRouter("/books", &controllers.BookController{}, "post:Create"),
			beego.NSRouter("/books/:id", &controllers.BookController{}, "get:Get;put:Update;delete:Delete"), // you can't navigate the code in editors!!!
		),
	)
}
```


### run cmd

```
./beegorouter -input=<the input file>
```

To overwrite the source file, add `-o` flag.

```go
package routers

import (
        "controllers"
        "github.com/astaxie/beego"
        . "github.com/icattlecoder/beegoroutable"
)

func register() {
        beego.NewNamespace("/api",
                beego.NSNamespace("/v1",
                        beego.NSRouter("/books", &controllers.BookController{}, MappingMethods(POST(controllers.DefaultBookController.Create))),
                        beego.NSRouter("/books/:id", &controllers.BookController{}, MappingMethods(GET(controllers.DefaultBookController.Get), PUT(controllers.DefaultBookController.Update), DELETE(controllers.DefaultBookController.Delete))),
                ),
        )
}
```


Then you need declare a `DefaultBookController` in your controller file, have fun!