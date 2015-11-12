# CORS gin's middleware
Gin middleware/handler to enable CORS support.

## Usage

###Canonical example:

```go
package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://foo.com"},
		AllowMethods:     []string{"PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
	router.Run()
}
```

###Using DefaultConfig as start point

```go
func main() {
    router := gin.Default()
    // - No origin allowed by default
    // - GET,POST, PUT, HEAD methods
    // - Credentials share disabled
    // - Preflight requests cached for 12 hours
    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://google.com"}
    config.AddAllowOrigins("http://facebook.com")
    // config.AllowOrigins == []string{"http://google.com", "http://facebook.com"}

    router.Use(cors.New(config))
    router.Run()
}
```

###Default() allows all origins

```go
    router := gin.Default()
    // same as
    // config := cors.DefaultConfig()
    // config.AllowAllOrigins = true
    // router.Use(cors.Default())
    router.Use(cors.Default())
    router.Run()
```



