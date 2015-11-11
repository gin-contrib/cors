package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
			AllowedOrigins:  []string{"http://google.com"},
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
		"http-access ", "post", "POST", " poSt  ",
		"HTTP-Access", "",
	})
	assert.Equal(t, values, []string{"Http-Access", "Post", ""})

	values = normalize(nil)
	assert.Nil(t, values)

	values = normalize([]string{})
	assert.Equal(t, values, []string{})
}

func TestGenerateNormalHeaders(t *testing.T) {
	header := generateNormalHeaders(Config{
		AllowAllOrigins: false,
	})
	assert.Contains(t, header.Get("Access-Control-Allow-Origin"), "")
	assert.Contains(t, header.Get("Vary"), "Origin")

	header = generateNormalHeaders(Config{
		AllowAllOrigins: true,
	})
	assert.Contains(t, header.Get("Access-Control-Allow-Origin"), "*")
	assert.Contains(t, header.Get("Vary"), "")

	header = generateNormalHeaders(Config{
		AllowCredentials: true,
	})
	assert.Contains(t, header.Get("Access-Control-Allow-Credentials"), "true")

	header = generateNormalHeaders(Config{
		AllowCredentials: false,
	})
	assert.Contains(t, header.Get("Access-Control-Allow-Credentials"), "")

	header = generateNormalHeaders(Config{
		ExposedHeaders: []string{"x-user", "xpassword"},
	})
	assert.Contains(t, header.Get("Access-Control-Expose-Headers"), "x-user, xpassword")
}

//
// func TestDeny0(t *testing.T) {
// 	called := false
//
// 	router := gin.New()
// 	router.Use(New(Config{
// 		AllowedOrigins: []string{"http://example.com"},
// 	}))
// 	router.GET("/", func(c *gin.Context) {
// 		called = true
// 	})
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/", nil)
// 	req.Header.Set("Origin", "https://example.com")
// 	router.ServeHTTP(w, req)
//
// 	assert.True(t, called)
// 	assert.NotContains(t, w.Header(), "Access-Control")
// }
//
// func TestDenyAbortOnError(t *testing.T) {
// 	called := false
//
// 	router := gin.New()
// 	router.Use(New(Config{
// 		AbortOnError:   true,
// 		AllowedOrigins: []string{"http://example.com"},
// 	}))
// 	router.GET("/", func(c *gin.Context) {
// 		called = true
// 	})
//
// 	w := httptest.NewRecorder()
// 	req, _ := http.NewRequest("GET", "/", nil)
// 	req.Header.Set("Origin", "https://example.com")
// 	router.ServeHTTP(w, req)
//
// 	assert.False(t, called)
// 	assert.NotContains(t, w.Header(), "Access-Control")
// }
//
// func TestDeny2(t *testing.T) {
//
// }
// func TestDeny3(t *testing.T) {
//
// }
//
// func TestPasses0(t *testing.T) {
//
// }
//
// func TestPasses1(t *testing.T) {
//
// }
//
// func TestPasses2(t *testing.T) {
//
// }
