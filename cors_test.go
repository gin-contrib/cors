package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestBadConfig(t *testing.T) {
	assert.Panics(t, func() { New(Config{}) })
	assert.Panics(t, func() {
		New(Config{
			AllowAllOrigins: true,
			AllowedOrigins:  []string{"http://google.com"},
		})
	})
	assert.Panics(t, func() {
		New(Config{
			AllowAllOrigins: true,
			AllowOriginFunc: func(origin string) bool { return false },
		})
	})
	assert.Panics(t, func() {
		New(Config{
			AllowedOrigins: []string{"google.com"},
		})
	})
}

func TestNormalize(t *testing.T) {
	values := normalize([]string{
		"http-Access ", "Post", "POST", " poSt  ",
		"HTTP-Access", "",
	})
	assert.Equal(t, values, []string{"http-access", "post", ""})

	values = normalize(nil)
	assert.Nil(t, values)

	values = normalize([]string{})
	assert.Equal(t, values, []string{})
}

func TestGenerateNormalHeaders_AllowAllOrigins(t *testing.T) {
	header := generateNormalHeaders(Config{
		AllowAllOrigins: false,
	})
	assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "")
	assert.Equal(t, header.Get("Vary"), "Origin")
	assert.Len(t, header, 1)

	header = generateNormalHeaders(Config{
		AllowAllOrigins: true,
	})
	assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "*")
	assert.Equal(t, header.Get("Vary"), "")
	assert.Len(t, header, 1)
}

func TestGenerateNormalHeaders_AllowCredentials(t *testing.T) {
	header := generateNormalHeaders(Config{
		AllowCredentials: true,
	})
	assert.Equal(t, header.Get("Access-Control-Allow-Credentials"), "true")
	assert.Equal(t, header.Get("Vary"), "Origin")
	assert.Len(t, header, 2)
}

func TestGenerateNormalHeaders_ExposedHeaders(t *testing.T) {
	header := generateNormalHeaders(Config{
		ExposedHeaders: []string{"X-user", "xPassword"},
	})
	assert.Equal(t, header.Get("Access-Control-Expose-Headers"), "x-user, xpassword")
	assert.Equal(t, header.Get("Vary"), "Origin")
	assert.Len(t, header, 2)
}

func TestGeneratePreflightHeaders(t *testing.T) {
	header := generatePreflightHeaders(Config{
		AllowAllOrigins: false,
	})
	assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "")
	assert.Equal(t, header.Get("Vary"), "Origin")
	assert.Len(t, header, 1)

	header = generateNormalHeaders(Config{
		AllowAllOrigins: true,
	})
	assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "*")
	assert.Equal(t, header.Get("Vary"), "")
	assert.Len(t, header, 1)
}

func TestGeneratePreflightHeaders_AllowCredentials(t *testing.T) {
	header := generatePreflightHeaders(Config{
		AllowCredentials: true,
	})
	assert.Equal(t, header.Get("Access-Control-Allow-Credentials"), "true")
	assert.Equal(t, header.Get("Vary"), "Origin")
	assert.Len(t, header, 2)
}

func TestGeneratePreflightHeaders_AllowedMethods(t *testing.T) {
	header := generatePreflightHeaders(Config{
		AllowedMethods: []string{"GET ", "post", "PUT", " put  "},
	})
	assert.Equal(t, header.Get("Access-Control-Allow-Methods"), "get, post, put")
	assert.Equal(t, header.Get("Vary"), "Origin")
	assert.Len(t, header, 2)
}

func TestGeneratePreflightHeaders_AllowedHeaders(t *testing.T) {
	header := generatePreflightHeaders(Config{
		AllowedHeaders: []string{"X-user", "Content-Type"},
	})
	assert.Equal(t, header.Get("Access-Control-Allow-Headers"), "x-user, content-type")
	assert.Equal(t, header.Get("Vary"), "Origin")
	assert.Len(t, header, 2)
}

func TestGeneratePreflightHeaders_MaxAge(t *testing.T) {
	header := generatePreflightHeaders(Config{
		MaxAge: 12 * time.Hour,
	})
	assert.Equal(t, header.Get("Access-Control-Max-Age"), "43200") // 12*60*60
	assert.Equal(t, header.Get("Vary"), "Origin")
	assert.Len(t, header, 2)
}

func TestValidateOrigin(t *testing.T) {
	cors := newCors(Config{
		AllowAllOrigins: true,
	})
	assert.True(t, cors.validateOrigin("http://google.com"))
	assert.True(t, cors.validateOrigin("https://google.com"))
	assert.True(t, cors.validateOrigin("example.com"))

	cors = newCors(Config{
		AllowedOrigins: []string{"https://google.com", "https://github.com"},
		AllowOriginFunc: func(origin string) bool {
			return (origin == "http://news.ycombinator.com")
		},
	})
	assert.False(t, cors.validateOrigin("http://google.com"))
	assert.True(t, cors.validateOrigin("https://google.com"))
	assert.True(t, cors.validateOrigin("https://github.com"))
	assert.True(t, cors.validateOrigin("http://news.ycombinator.com"))
	assert.False(t, cors.validateOrigin("http://example.com"))
	assert.False(t, cors.validateOrigin("google.com"))
}

func TestPasses0(t *testing.T) {
	called := false
	router := gin.New()
	router.Use(New(Config{
		AllowedOrigins:   []string{"http://google.com"},
		AllowedMethods:   []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"},
		AllowedHeaders:   []string{"Content-type", "timeStamp "},
		ExposedHeaders:   []string{"Data", "x-User"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			return origin == "http://github.com"
		},
	}))
	router.GET("/", func(c *gin.Context) {
		called = true
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)
	assert.True(t, called)
	assert.NotContains(t, w.Header(), "Access-Control-Allow-Origin")
	assert.NotContains(t, w.Header(), "Access-Control-Allow-Credentials")
	assert.NotContains(t, w.Header(), "Access-Control-Expose-Headers")

	called = false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://google.com")
	router.ServeHTTP(w, req)
	assert.True(t, called)
	assert.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), "http://google.com")
	assert.Equal(t, w.Header().Get("Access-Control-Allow-Credentials"), "true")
	assert.Equal(t, w.Header().Get("Access-Control-Expose-Headers"), "data, x-user")

	called = false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://google.com")
	router.ServeHTTP(w, req)
	assert.False(t, called)
	assert.NotContains(t, w.Header(), "Access-Control-Allow-Origin")
	assert.NotContains(t, w.Header(), "Access-Control-Allow-Credentials")
	assert.NotContains(t, w.Header(), "Access-Control-Expose-Headers")

	called = false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://github.com")
	router.ServeHTTP(w, req)
	assert.False(t, called)
	assert.Equal(t, w.Header().Get("Access-Control-Allow-Origin"), "http://github.com")
	assert.Equal(t, w.Header().Get("Access-Control-Allow-Credentials"), "true")
	assert.Equal(t, w.Header().Get("Access-Control-Allow-Methods"), "get, post, put, head")
	assert.Equal(t, w.Header().Get("Access-Control-Allow-Headers"), "content-type, timestamp")
	assert.Equal(t, w.Header().Get("Access-Control-Max-Age"), "43200")

	called = false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	router.ServeHTTP(w, req)
	assert.False(t, called)
	assert.NotContains(t, w.Header(), "Access-Control-Allow-Origin")
	assert.NotContains(t, w.Header(), "Access-Control-Allow-Credentials")
	assert.NotContains(t, w.Header(), "Access-Control-Allow-Methods")
	assert.NotContains(t, w.Header(), "Access-Control-Allow-Headers")
	assert.NotContains(t, w.Header(), "Access-Control-Max-Age")
}

func TestPasses1(t *testing.T) {

}

func TestPasses2(t *testing.T) {

}
