# gin-contrib/cors

[![Run Tests](https://github.com/gin-contrib/cors/actions/workflows/go.yml/badge.svg)](https://github.com/gin-contrib/cors/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/gin-contrib/cors/branch/master/graph/badge.svg)](https://codecov.io/gh/gin-contrib/cors)
[![Go Report Card](https://goreportcard.com/badge/github.com/gin-contrib/cors)](https://goreportcard.com/report/github.com/gin-contrib/cors)
[![GoDoc](https://godoc.org/github.com/gin-contrib/cors?status.svg)](https://godoc.org/github.com/gin-contrib/cors)

CORS (Cross-Origin Resource Sharing) middleware for [Gin](https://github.com/gin-gonic/gin). Enables flexible CORS handling for your Gin-based APIs.

- [gin-contrib/cors](#gin-contribcors)
  - [Installation](#installation)
  - [Quick Start Example](#quick-start-example)
  - [Advanced Usage](#advanced-usage)
    - [Custom Configuration](#custom-configuration)
    - [DefaultConfig Reference](#defaultconfig-reference)
    - [Default() Convenience](#default-convenience)
  - [Important Notes](#important-notes)

---

## Installation

Install with:

```sh
go get github.com/gin-contrib/cors
```

Import the package in your Go code:

```go
import "github.com/gin-contrib/cors"
```

---

## Quick Start Example

Basic usage with default (all origins allowed):

```go
import (
  "github.com/gin-contrib/cors"
  "github.com/gin-gonic/gin"
)

func main() {
  router := gin.Default()
  router.Use(cors.Default()) // All origins allowed by default
  router.Run()
}
```

> **Warning:** Allowing all origins disables the ability for Gin to set cookies for clients. For credentialed requests, DO NOT allow all origins.

---

## Advanced Usage

### Custom Configuration

Configure allowed origins, methods, and more:

```go
import (
  "time"

  "github.com/gin-contrib/cors"
  "github.com/gin-gonic/gin"
)

func main() {
  router := gin.Default()
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

### DefaultConfig Reference

Start with the library defaults and customize as desired.

```go
import (
  "github.com/gin-contrib/cors"
  "github.com/gin-gonic/gin"
)

func main() {
  router := gin.Default()
  // By default:
  //   - No origins are allowed
  //   - Methods GET, POST, PUT, HEAD are allowed
  //   - Credentials are NOT allowed
  //   - Preflight requests are cached for 12 hours
  config := cors.DefaultConfig()
  config.AllowOrigins = []string{"http://google.com"}
  // config.AllowOrigins = []string{"http://google.com", "http://facebook.com"}
  // config.AllowAllOrigins = true

  router.Use(cors.New(config))
  router.Run()
}
```

> **Note:** `Default()` allows all origins, but `DefaultConfig()` does **not**. To allow all origins, set `AllowAllOrigins = true` explicitly.

### Default() Convenience

A simple method to enable all origins:

```go
router.Use(cors.Default()) // Equivalent to AllowAllOrigins = true
```

---

## Important Notes

- **Enabling all origins disables cookies:** When `AllowAllOrigins` is enabled, Gin cannot set cookies for clients. If you need credential sharing (cookies, authentication headers), **do not** allow all origins.
- For detailed documentation and configuration options, see the [GoDoc](https://godoc.org/github.com/gin-contrib/cors).
