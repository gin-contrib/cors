package cors

import (
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type cors struct {
	allowAllOrigins   bool
	allowedOriginFunc func(string) bool
	allowedOrigins    []string
	allowedMethods    []string
	allowedHeaders    []string
	exposedHeaders    []string
	normalHeaders     http.Header
	preflightHeaders  http.Header
}

func newCors(c Config) *cors {
	if err := c.Validate(); err != nil {
		panic(err.Error())
	}
	return &cors{
		allowedOriginFunc: c.AllowOriginFunc,
		allowAllOrigins:   c.AllowAllOrigins,
		allowedOrigins:    normalize(c.AllowedOrigins),
		allowedMethods:    normalize(c.AllowedMethods),
		allowedHeaders:    normalize(c.AllowedHeaders),
		normalHeaders:     generateNormalHeaders(c),
		preflightHeaders:  generatePreflightHeaders(c),
	}
}

func (cors *cors) applyCors(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if len(origin) == 0 {
		// request is not a CORS request
		return
	}
	if !cors.validateOrigin(origin) {
		goto failed
	}

	if c.Request.Method == "OPTIONS" {
		if !cors.handlePreflight(c) {
			goto failed
		}
	} else if !cors.handleNormal(c) {
		goto failed
	}
	if cors.allowAllOrigins {
		c.Header("Access-Control-Allow-Origin", "*")
	} else {
		c.Header("Access-Control-Allow-Origin", origin)
	}
	return

failed:
	c.AbortWithStatus(http.StatusForbidden)
}

func (cors *cors) validateOrigin(origin string) bool {
	if cors.allowAllOrigins {
		return true
	}
	if cors.allowedOriginFunc != nil {
		return cors.allowedOriginFunc(origin)
	}
	for _, value := range cors.allowedOrigins {
		if value == origin {
			return true
		}
	}
	return false
}

func (cors *cors) validateMethod(method string) bool {
	for _, value := range cors.allowedMethods {
		if strings.EqualFold(value, method) {
			return true
		}
	}
	return false
}

func (cors *cors) validateHeader(header string) bool {
	for _, value := range cors.allowedHeaders {
		if strings.EqualFold(value, header) {
			return true
		}
	}
	return false
}

func (cors *cors) handlePreflight(c *gin.Context) bool {
	c.AbortWithStatus(200)
	if !cors.validateMethod(c.Request.Header.Get("Access-Control-Request-Method")) {
		return false
	}
	if !cors.validateHeader(c.Request.Header.Get("Access-Control-Request-Header")) {
		return false
	}
	for key, value := range cors.preflightHeaders {
		c.Writer.Header()[key] = value
	}
	return true
}

func (cors *cors) handleNormal(c *gin.Context) bool {
	for key, value := range cors.normalHeaders {
		c.Writer.Header()[key] = value
	}
	return true
}

func generateNormalHeaders(c Config) http.Header {
	headers := make(http.Header)
	if c.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if len(c.ExposedHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(c.ExposedHeaders, ", "))
	}
	if c.AllowAllOrigins {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Vary", "Origin")
	}
	return headers
}

func generatePreflightHeaders(c Config) http.Header {
	headers := make(http.Header)
	if c.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if len(c.AllowedMethods) > 0 {
		value := strings.Join(c.AllowedMethods, ", ")
		headers.Set("Access-Control-Allow-Methods", value)
	}
	if len(c.AllowedHeaders) > 0 {
		value := strings.Join(c.AllowedHeaders, ", ")
		headers.Set("Access-Control-Allow-Headers", value)
	}
	if c.MaxAge > time.Duration(0) {
		value := strconv.FormatInt(int64(c.MaxAge/time.Second), 10)
		headers.Set("Access-Control-Max-Age", value)
	}
	if c.AllowAllOrigins {
		headers.Set("Access-Control-Allow-Origin", "*")
	} else {
		headers.Set("Vary", "Origin")
	}
	return headers
}

func normalize(values []string) []string {
	if values == nil {
		return nil
	}
	distinctMap := make(map[string]bool, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		value = textproto.CanonicalMIMEHeaderKey(value)
		if _, seen := distinctMap[value]; !seen {
			normalized = append(normalized, value)
			distinctMap[value] = true
		}
	}
	return normalized
}
