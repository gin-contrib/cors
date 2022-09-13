# CORS gin's middleware

[![Run Tests](https://github.com/andressg79/cors/actions/workflows/go.yml/badge.svg)](https://github.com/andressg79/cors/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/andressg79/cors/branch/master/graph/badge.svg)](https://codecov.io/gh/andressg79/cors)
[![Go Report Card](https://goreportcard.com/badge/github.com/andressg79/cors)](https://goreportcard.com/report/github.com/andressg79/cors)
[![GoDoc](https://godoc.org/github.com/andressg79/cors?status.svg)](https://godoc.org/github.com/andressg79/cors)

Gin middleware/handler to enable CORS support.

## Usage

### Start using it

Download and install it:

```sh
go get github.com/andressg79/cors
```

Import it in your code:

```go
import "github.com/andressg79/cors"
```

### Canonical example

```go
package main

import (
  "time"

  "github.com/andressg79/cors"
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

### Using DefaultConfig as start point

```go
func main() {
  router := gin.Default()
  // - No origin allowed by default
  // - GET,POST, PUT, HEAD methods
  // - Credentials share disabled
  // - Preflight requests cached for 12 hours
  config := cors.DefaultConfig()
  config.AllowOrigins = []string{"http://google.com"}
  // config.AllowOrigins = []string{"http://google.com", "http://facebook.com"}
  // config.AllowAllOrigins = true

  router.Use(cors.New(config))
  router.Run()
}
```
note: while Default() allows all origins, DefaultConfig() does not and you will still have to use AllowAllOrigins

### Default() allows all origins

```go
func main() {
  router := gin.Default()
  // same as
  // config := cors.DefaultConfig()
  // config.AllowAllOrigins = true
  // router.Use(cors.New(config))
  router.Use(cors.Default())
  router.Run()
}
```
