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
  - [Configuration Reference](#configuration-reference)
    - [Notes on Configuration](#notes-on-configuration)
      - [Example: Using advanced options](#example-using-advanced-options)
      - [Using custom origin validation](#using-custom-origin-validation)
      - [With Gin context](#with-gin-context)
  - [Helper Methods](#helper-methods)
  - [Validation \& Error Handling](#validation--error-handling)
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

## Configuration Reference

This middleware is controlled via the `cors.Config` struct. All fields are optional unless otherwise stated.

| Field                         | Type                        | Default                                                   | Description                                                                                                                                                    |
|-------------------------------|-----------------------------|-----------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `AllowAllOrigins`             | `bool`                      | `false`                                                   | If true, allows all origins, ignoring `AllowOrigins` and origin checking functions. If true, credentials **cannot** be used.                                   |
| `AllowOrigins`                | `[]string`                  | `[]`                                                      | List of allowed origins. Supports exact match, `*` for all, and wildcards (see below). Example: `[]string{"https://foo.com"}`                                  |
| `AllowOriginFunc`             | `func(string) bool`         | `nil`                                                     | Custom function to validate origin. If set, `AllowOrigins` is ignored.                                                                                         |
| `AllowOriginWithContextFunc`  | `func(*gin.Context,string)bool` | `nil`                                               | Like `AllowOriginFunc`, but allows access to request context. (Read-only: no mutation/side-effects on the request.)                                            |
| `AllowMethods`                | `[]string`                  | `[]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}` | Allowed HTTP methods for cross-domain requests.                                                                         |
| `AllowPrivateNetwork`         | `bool`                      | `false`                                                   | Adds the [Private Network Access](https://wicg.github.io/private-network-access/) CORS header (Chrome/Edge feature).                                           |
| `AllowHeaders`                | `[]string`                  | `[]`                                                      | List of non-simple headers permitted in requests. E.g. `[]string{"Origin"}`                                                                                    |
| `AllowCredentials`            | `bool`                      | `false`                                                   | Allow cookies, HTTP auth, or client certs in CORS requests. Only if precise origins (not `*`) are used.                                                        |
| `ExposeHeaders`               | `[]string`                  | `[]`                                                      | Headers exposed to the browser. E.g. `[]string{"Content-Length"}`                                                                                              |
| `MaxAge`                      | `time.Duration`             | `12 * time.Hour`                                          | Cache time for preflight requests.                                                                                                                             |
| `AllowWildcard`               | `bool`                      | `false`                                                   | Enables support for wildcards in origins (e.g. `https://*.example.com`). Only one `*` per origin string is allowed.                                            |
| `AllowBrowserExtensions`      | `bool`                      | `false`                                                   | Allow standard browser extension schemes as origins (e.g. `chrome-extension://`).                                                                              |
| `CustomSchemas`               | `[]string`                  | `nil`                                                     | Additional allowed URI schemes. Example: `[]string{"tauri://"}`                                                                                                |
| `AllowWebSockets`             | `bool`                      | `false`                                                   | Allow `ws://` and `wss://` schemas.                                                                                                                            |
| `AllowFiles`                  | `bool`                      | `false`                                                   | Allow `file://` origins (dangerous; only use when absolutely necessary).                                                                                       |
| `OptionsResponseStatusCode`   | `int`                       | `204`                                                     | Custom status code for `OPTIONS` responses for legacy browsers/clients.                                                                                        |

### Notes on Configuration

- Only one of `AllowAllOrigins`, `AllowOrigins`, `AllowOriginFunc` or `AllowOriginWithContextFunc` should be set.
- If `AllowAllOrigins` is true, other origin settings are ignored and credentialed requests are not allowed.
- If `AllowWildcard` is enabled, only one `*` is allowed per origin string.
- Use `AllowBrowserExtensions`, `AllowWebSockets`, or `AllowFiles` to permit non-HTTP(s) protocols as origins.
- Custom schemas allow, for example, usage in desktop apps via custom URI schemes (`tauri://`, etc.).
- Setting both `AllowOriginFunc` and `AllowOriginWithContextFunc` is allowed, the context-specific function will be preferred if both are set.

#### Example: Using advanced options

```go
config := cors.Config{
  AllowOrigins:           []string{"https://*.foo.com", "https://bar.com"},
  AllowWildcard:          true,
  AllowMethods:           []string{"GET", "POST"},
  AllowHeaders:           []string{"Authorization", "Content-Type"},
  AllowCredentials:       true,
  AllowBrowserExtensions: true, // Allow browser extensions
  AllowWebSockets:        true, // Allow ws://
  AllowFiles:             false,
  CustomSchemas:          []string{"tauri://"},
  MaxAge:                 24 * time.Hour,
  ExposeHeaders:          []string{"X-Custom-Header"},
  AllowPrivateNetwork:    true, // Chrome/Edge feature
}
```

#### Using custom origin validation

```go
config := cors.Config{
  AllowOriginFunc: func(origin string) bool {
    // Allow any github.com subdomain or a custom rule
    return strings.HasSuffix(origin, "github.com")
  },
}
```

#### With Gin context

```go
config := cors.Config{
  AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool {
    // Allow only if a certain header is present
    return c.Request.Header.Get("X-Allow-CORS") == "yes"
  },
}
```

---

## Helper Methods

For dynamically adding methods or headers to the config:

- **AddAllowMethods(...string):** Add allowed methods.
- **AddAllowHeaders(...string):** Add allowed headers.
- **AddExposeHeaders(...string):** Add exposed headers.

**Example:**

```go
config.AddAllowMethods("DELETE", "OPTIONS")
config.AddAllowHeaders("X-My-Header")
config.AddExposeHeaders("X-Other-Header")
```

---

## Validation & Error Handling

Calling `Validate()` on a `Config` checks for misconfiguration (called internally):

- If `AllowAllOrigins` is set, you cannot also set `AllowOrigins` or any `AllowOriginFunc`.
- If neither `AllowAllOrigins`, `AllowOriginFunc`, nor `AllowOrigins` is set, an error is raised.
- If an `AllowOrigin` contains a wild-card but `AllowWildcard` is not enabled, or more than one `*` is present, a panic is triggered.
- Invalid origin schemas or unsupported wildcards are rejected.

---

## Important Notes

- **Enabling all origins disables cookies:** When `AllowAllOrigins` is enabled, Gin cannot set cookies for clients. If you need credential sharing (cookies, authentication headers), **do not** allow all origins.
- For detailed documentation and configuration options, see the [GoDoc](https://godoc.org/github.com/gin-contrib/cors).
