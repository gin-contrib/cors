package cors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newTestRouter(config Config) *gin.Engine {
	router := gin.New()
	router.Use(New(config))
	router.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "get") })
	router.POST("/", func(c *gin.Context) { c.String(http.StatusOK, "post") })
	router.PATCH("/", func(c *gin.Context) { c.String(http.StatusOK, "patch") })
	return router
}

func multiGroupRouter(config Config) *gin.Engine {
	router := gin.New()
	router.Use(New(config))
	app1 := router.Group("/app1")
	app1.GET("", func(c *gin.Context) { c.String(http.StatusOK, "app1") })
	app2 := router.Group("/app2")
	app2.GET("", func(c *gin.Context) { c.String(http.StatusOK, "app2") })
	app3 := router.Group("/app3")
	app3.GET("", func(c *gin.Context) { c.String(http.StatusOK, "app3") })
	return router
}

func performRequest(r http.Handler, method, origin string) *httptest.ResponseRecorder {
	return performRequestWithHeaders(r, method, "/", origin, http.Header{})
}

func performRequestWithHeaders(r http.Handler, method, path, origin string, header http.Header) *httptest.ResponseRecorder {
	req, _ := http.NewRequestWithContext(context.Background(), method, path, nil)
	req.Host = header.Get("Host")
	header.Del("Host")
	if origin != "" {
		header.Set("Origin", origin)
	}
	req.Header = header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestConfigAddAllow(t *testing.T) {
	config := Config{}
	config.AddAllowMethods("POST")
	config.AddAllowMethods("GET", "PUT")
	config.AddExposeHeaders()
	config.AddAllowHeaders("Some", " cool")
	config.AddAllowHeaders("header")
	config.AddExposeHeaders()
	config.AddExposeHeaders()
	config.AddExposeHeaders("exposed", "header")
	config.AddExposeHeaders("hey")

	assert.Equal(t, []string{"POST", "GET", "PUT"}, config.AllowMethods)
	assert.Equal(t, []string{"Some", " cool", "header"}, config.AllowHeaders)
	assert.Equal(t, []string{"exposed", "header", "hey"}, config.ExposeHeaders)
}

func TestBadConfig(t *testing.T) {
	tests := []Config{
		{},
		{AllowAllOrigins: true, AllowOrigins: []string{"http://google.com"}},
		{AllowAllOrigins: true, AllowOriginFunc: func(origin string) bool { return false }},
		{AllowOrigins: []string{"google.com"}},
		{AllowOrigins: []string{"/http://google.com"}},
		{AllowOrigins: []string{"http?://google.com"}},
		{AllowOrigins: []string{"http?://google.com/g"}},
	}
	for _, cfg := range tests {
		assert.Panics(t, func() { New(cfg) })
	}
}

func TestNormalize(t *testing.T) {
	assert.Equal(t, []string{"http-access", "post", ""}, normalize([]string{
		"http-Access ", "Post", "POST", " poSt  ", "HTTP-Access", "",
	}))
	assert.Nil(t, normalize(nil))
	assert.Equal(t, []string{}, normalize([]string{}))
}

func TestConvert(t *testing.T) {
	methods := []string{"Get", "GET", "get"}
	headers := []string{"X-CSRF-TOKEN", "X-CSRF-Token", "x-csrf-token"}
	assert.Equal(t, []string{"GET", "GET", "GET"}, convert(methods, strings.ToUpper))
	assert.Equal(t, []string{"X-Csrf-Token", "X-Csrf-Token", "X-Csrf-Token"}, convert(headers, http.CanonicalHeaderKey))
}

func TestGenerateNormalHeaders(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		expect map[string]string
		len    int
	}{
		{
			"AllowAllOrigins false",
			Config{AllowAllOrigins: false},
			map[string]string{"Access-Control-Allow-Origin": "", "Vary": "Origin"},
			1,
		},
		{
			"AllowAllOrigins true",
			Config{AllowAllOrigins: true},
			map[string]string{"Access-Control-Allow-Origin": "*", "Vary": ""},
			1,
		},
		{
			"AllowCredentials true",
			Config{AllowCredentials: true},
			map[string]string{"Access-Control-Allow-Credentials": "true", "Vary": "Origin"},
			2,
		},
		{
			"ExposeHeaders set",
			Config{ExposeHeaders: []string{"X-user", "xPassword"}},
			map[string]string{"Access-Control-Expose-Headers": "X-User,Xpassword", "Vary": "Origin"},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := generateNormalHeaders(tt.config)
			for k, v := range tt.expect {
				assert.Equal(t, v, header.Get(k))
			}
			assert.Len(t, header, tt.len)
		})
	}
}

func TestGeneratePreflightHeaders(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		expect map[string]string
		len    int
	}{
		{
			"AllowAllOrigins false",
			Config{AllowAllOrigins: false},
			map[string]string{"Access-Control-Allow-Origin": "", "Vary": "Origin"},
			1,
		},
		{
			"AllowAllOrigins true",
			Config{AllowAllOrigins: true},
			map[string]string{"Access-Control-Allow-Origin": "*", "Vary": ""},
			1,
		},
		{
			"AllowCredentials true",
			Config{AllowCredentials: true},
			map[string]string{"Access-Control-Allow-Credentials": "true", "Vary": "Origin"},
			2,
		},
		{
			"AllowPrivateNetwork true",
			Config{AllowPrivateNetwork: true},
			map[string]string{"Access-Control-Allow-Private-Network": "true", "Vary": "Origin"},
			2,
		},
		{
			"AllowMethods set",
			Config{AllowMethods: []string{"GET ", "post", "PUT", " put  "}},
			map[string]string{"Access-Control-Allow-Methods": "GET,POST,PUT", "Vary": "Origin"},
			2,
		},
		{
			"AllowHeaders set",
			Config{AllowHeaders: []string{"X-user", "Content-Type"}},
			map[string]string{"Access-Control-Allow-Headers": "X-User,Content-Type", "Vary": "Origin"},
			2,
		},
		{
			"MaxAge set",
			Config{MaxAge: 12 * time.Hour},
			map[string]string{"Access-Control-Max-Age": "43200", "Vary": "Origin"},
			2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := generatePreflightHeaders(tt.config)
			for k, v := range tt.expect {
				assert.Equal(t, v, header.Get(k))
			}
			assert.Len(t, header, tt.len)
		})
	}
}

func TestValidateOrigin(t *testing.T) {
	type originTest struct {
		config  Config
		origins map[string]bool
	}
	tests := []originTest{
		{
			Config{AllowAllOrigins: true},
			map[string]bool{
				"http://google.com": true, "https://google.com": true, "example.com": true, "chrome-extension://random-extension-id": true,
			},
		},
		{
			Config{
				AllowOrigins:           []string{"https://google.com", "https://github.com"},
				AllowOriginFunc:        func(origin string) bool { return origin == "http://news.ycombinator.com" },
				AllowBrowserExtensions: true,
			},
			map[string]bool{
				"http://google.com": false, "https://google.com": true, "https://github.com": true,
				"http://news.ycombinator.com": true, "http://example.com": false, "google.com": false, "chrome-extension://random-extension-id": false,
			},
		},
		{
			Config{AllowOrigins: []string{"https://google.com", "https://github.com"}},
			map[string]bool{
				"chrome-extension://random-extension-id": false, "file://some-dangerous-file.js": false, "wss://socket-connection": false,
			},
		},
		{
			Config{
				AllowOrigins: []string{
					"chrome-extension://*",
					"safari-extension://my-extension-*-app",
					"*.some-domain.com",
				},
				AllowBrowserExtensions: true,
				AllowWildcard:          true,
			},
			map[string]bool{
				"chrome-extension://random-extension-id": true, "chrome-extension://another-one": true,
				"safari-extension://my-extension-one-app": true, "safari-extension://my-extension-two-app": true,
				"moz-extension://ext-id-we-not-allow": false, "http://api.some-domain.com": true, "http://api.another-domain.com": false,
			},
		},
		{
			Config{
				AllowOrigins:    []string{"file://safe-file.js", "wss://some-session-layer-connection"},
				AllowFiles:      true,
				AllowWebSockets: true,
			},
			map[string]bool{
				"file://safe-file.js": true, "file://some-dangerous-file.js": false,
				"wss://some-session-layer-connection": true, "ws://not-what-we-expected": false,
			},
		},
		{
			Config{AllowOrigins: []string{"*"}},
			map[string]bool{
				"http://google.com": true, "https://google.com": true, "example.com": true, "chrome-extension://random-extension-id": true,
			},
		},
		{
			Config{AllowOrigins: []string{"/https?://(?:.+\\.)?google\\.com/g"}},
			map[string]bool{
				"http://google.com": true, "https://google.com": true, "https://maps.google.com": true, "https://maps.test.google.com": true, "https://maps.google.it": false,
			},
		},
	}
	for i, test := range tests {
		cors := newCors(test.config)
		for origin, want := range test.origins {
			got := cors.validateOrigin(origin)
			assert.Equalf(t, want, got, "case %d: origin=%s", i, origin)
		}
	}
}

func TestValidateTauri(t *testing.T) {
	c := Config{
		AllowOrigins:           []string{"tauri://localhost:1234"},
		AllowBrowserExtensions: true,
	}
	assert.Error(t, c.Validate())

	c = Config{
		AllowOrigins:           []string{"tauri://localhost:1234"},
		AllowBrowserExtensions: true,
		CustomSchemas:          []string{"tauri"},
	}
	assert.Nil(t, c.Validate())
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	config.AllowAllOrigins = true
	router := newTestRouter(config)
	w := performRequest(router, "GET", "http://google.com")
	assert.Equal(t, "get", w.Body.String())
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))
}

func TestCORS_AllowOrigins_NoOrigin(t *testing.T) {
	router := newTestRouter(Config{
		AllowOrigins:               []string{"http://google.com"},
		AllowMethods:               []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"},
		AllowHeaders:               []string{"Content-type", "timeStamp "},
		ExposeHeaders:              []string{"Data", "x-User"},
		AllowCredentials:           false,
		MaxAge:                     12 * time.Hour,
		AllowOriginFunc:            func(origin string) bool { return origin == "http://github.com" },
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool { return origin == "http://sample.com" },
	})
	w := performRequest(router, "GET", "")
	assert.Equal(t, "get", w.Body.String())
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))
}

func TestCORS_AllowOrigins_OriginIsHost(t *testing.T) {
	router := newTestRouter(Config{
		AllowOrigins:               []string{"http://google.com"},
		AllowMethods:               []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"},
		AllowHeaders:               []string{"Content-type", "timeStamp "},
		ExposeHeaders:              []string{"Data", "x-User"},
		AllowCredentials:           false,
		MaxAge:                     12 * time.Hour,
		AllowOriginFunc:            func(origin string) bool { return origin == "http://github.com" },
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool { return origin == "http://sample.com" },
	})
	h := http.Header{}
	h.Set("Host", "facebook.com")
	w := performRequestWithHeaders(router, "GET", "/", "http://facebook.com", h)
	assert.Equal(t, "get", w.Body.String())
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))
}

func TestCORS_AllowOrigins_AllowedOrigin(t *testing.T) {
	router := newTestRouter(Config{
		AllowOrigins:               []string{"http://google.com"},
		AllowMethods:               []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"},
		AllowHeaders:               []string{"Content-type", "timeStamp "},
		ExposeHeaders:              []string{"Data", "x-User"},
		AllowCredentials:           false,
		MaxAge:                     12 * time.Hour,
		AllowOriginFunc:            func(origin string) bool { return origin == "http://github.com" },
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool { return origin == "http://sample.com" },
	})

	tests := []struct {
		origin, wantExpose string
	}{
		{"http://google.com", "Data,X-User"},
		{"http://github.com", "Data,X-User"},
	}
	for _, tt := range tests {
		w := performRequest(router, "GET", tt.origin)
		assert.Equal(t, "get", w.Body.String())
		assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, tt.wantExpose, w.Header().Get("Access-Control-Expose-Headers"))
	}
}

func TestCORS_AllowOrigins_DeniedOrigin(t *testing.T) {
	router := newTestRouter(Config{
		AllowOrigins:               []string{"http://google.com"},
		AllowMethods:               []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"},
		AllowHeaders:               []string{"Content-type", "timeStamp "},
		ExposeHeaders:              []string{"Data", "x-User"},
		AllowCredentials:           false,
		MaxAge:                     12 * time.Hour,
		AllowOriginFunc:            func(origin string) bool { return origin == "http://github.com" },
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool { return origin == "http://sample.com" },
	})
	w := performRequest(router, "GET", "https://google.com")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))
}

func TestCORS_AllowOrigins_Preflight(t *testing.T) {
	router := newTestRouter(Config{
		AllowOrigins:               []string{"http://google.com"},
		AllowMethods:               []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"},
		AllowHeaders:               []string{"Content-type", "timeStamp "},
		ExposeHeaders:              []string{"Data", "x-User"},
		AllowCredentials:           false,
		MaxAge:                     12 * time.Hour,
		AllowOriginFunc:            func(origin string) bool { return origin == "http://github.com" },
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool { return origin == "http://sample.com" },
	})

	tests := []string{"http://github.com", "http://sample.com"}
	for _, origin := range tests {
		w := performRequest(router, "OPTIONS", origin)
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "GET,POST,PUT,HEAD", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type,Timestamp", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "43200", w.Header().Get("Access-Control-Max-Age"))
	}
}

func TestCORS_AllowOrigins_DeniedPreflight(t *testing.T) {
	router := newTestRouter(Config{
		AllowOrigins:               []string{"http://google.com"},
		AllowMethods:               []string{" GeT ", "get", "post", "PUT  ", "Head", "POST"},
		AllowHeaders:               []string{"Content-type", "timeStamp "},
		ExposeHeaders:              []string{"Data", "x-User"},
		AllowCredentials:           false,
		MaxAge:                     12 * time.Hour,
		AllowOriginFunc:            func(origin string) bool { return origin == "http://github.com" },
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool { return origin == "http://sample.com" },
	})
	w := performRequest(router, "OPTIONS", "http://example.com")
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
	assert.Empty(t, w.Header().Get("Access-Control-Max-Age"))
}

func TestPassesAllowAllOrigins(t *testing.T) {
	router := newTestRouter(Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{" Patch ", "get", "post", "POST"},
		AllowHeaders:     []string{"Content-type", "  testheader "},
		ExposeHeaders:    []string{"Data2", "x-User2"},
		AllowCredentials: false,
		MaxAge:           10 * time.Hour,
	})

	w := performRequest(router, "GET", "")
	assert.Equal(t, "get", w.Body.String())
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Empty(t, w.Header().Get("Access-Control-Expose-Headers"))

	w = performRequest(router, "POST", "example.com")
	assert.Equal(t, "post", w.Body.String())
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "Data2,X-User2", w.Header().Get("Access-Control-Expose-Headers"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))

	w = performRequest(router, "OPTIONS", "https://facebook.com")
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "PATCH,GET,POST", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type,Testheader", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "36000", w.Header().Get("Access-Control-Max-Age"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestWildcard(t *testing.T) {
	router := newTestRouter(Config{
		AllowOrigins:  []string{"https://*.github.com", "https://api.*", "http://*", "https://facebook.com", "*.golang.org"},
		AllowMethods:  []string{"GET"},
		AllowWildcard: true,
	})

	tests := []struct {
		origin string
		code   int
	}{
		{"https://gist.github.com", 200},
		{"https://api.github.com/v1/users", 200},
		{"https://giphy.com/", 403},
		{"http://hard-to-find-http-example.com", 200},
		{"https://facebook.com", 200},
		{"https://something.golang.org", 200},
		{"https://something.go.org", 403},
	}
	for _, tt := range tests {
		w := performRequest(router, "GET", tt.origin)
		assert.Equal(t, tt.code, w.Code)
	}

	router = newTestRouter(Config{
		AllowOrigins: []string{"https://github.com", "https://facebook.com"},
		AllowMethods: []string{"GET"},
	})

	tests2 := []struct {
		origin string
		code   int
	}{
		{"https://gist.github.com", 403},
		{"https://github.com", 200},
	}
	for _, tt := range tests2 {
		w := performRequest(router, "GET", tt.origin)
		assert.Equal(t, tt.code, w.Code)
	}
}

func TestMultiGroupRouter(t *testing.T) {
	router := multiGroupRouter(Config{
		AllowMethods: []string{"GET"},
		AllowOriginWithContextFunc: func(c *gin.Context, origin string) bool {
			path := c.Request.URL.Path
			switch {
			case strings.HasPrefix(path, "/app1"):
				return origin == "http://app1.example.com"
			case strings.HasPrefix(path, "/app2"):
				return origin == "http://app2.example.com"
			default:
				return true
			}
		},
	})

	emptyHeaders := http.Header{}
	app1Origin := "http://app1.example.com"
	app2Origin := "http://app2.example.com"
	randomOrigin := "http://random.com"

	tests := []struct {
		method, path, origin string
		code                 int
	}{
		{"OPTIONS", "/app1", app1Origin, http.StatusNoContent},
		{"OPTIONS", "/app2", app2Origin, http.StatusNoContent},
		{"OPTIONS", "/app3", randomOrigin, http.StatusNoContent},
		{"OPTIONS", "/app1", randomOrigin, http.StatusForbidden},
		{"OPTIONS", "/app2", randomOrigin, http.StatusForbidden},
	}
	for _, tt := range tests {
		w := performRequestWithHeaders(router, tt.method, tt.path, tt.origin, emptyHeaders)
		assert.Equal(t, tt.code, w.Code)
	}
}

func TestParseWildcardRules_NoWildcard(t *testing.T) {
	config := Config{
		AllowOrigins: []string{
			"http://example.com",
			"https://google.com",
			"github.com",
		},
		AllowWildcard: false,
	}
	assert.Equal(t, [][]string(nil), config.parseWildcardRules())
}

func TestParseWildcardRules_InvalidWildcard(t *testing.T) {
	config := Config{
		AllowOrigins: []string{
			"http://example.com",
			"https://*.google.com*",
			"*.github.com*",
		},
		AllowWildcard: true,
	}
	assert.Panics(t, func() { config.parseWildcardRules() })
}

func TestParseWildcardRules(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedResult [][]string
		expectPanic    bool
	}{
		{
			"Wildcard not allowed",
			Config{AllowWildcard: false, AllowOrigins: []string{"http://example.com", "https://*.domain.com"}},
			nil, false,
		},
		{
			"No wildcards",
			Config{AllowWildcard: true, AllowOrigins: []string{"http://example.com", "https://example.com"}},
			nil, false,
		},
		{
			"Single wildcard at the end",
			Config{AllowWildcard: true, AllowOrigins: []string{"http://*.example.com"}},
			[][]string{{"http://", ".example.com"}},
			false,
		},
		{
			"Single wildcard at the beginning",
			Config{AllowWildcard: true, AllowOrigins: []string{"*.example.com"}},
			[][]string{{"*", ".example.com"}},
			false,
		},
		{
			"Single wildcard in the middle",
			Config{AllowWildcard: true, AllowOrigins: []string{"http://example.*.com"}},
			[][]string{{"http://example.", ".com"}},
			false,
		},
		{
			"Multiple wildcards should panic",
			Config{AllowWildcard: true, AllowOrigins: []string{"http://*.*.com"}},
			nil, true,
		},
		{
			"Single wildcard in the end",
			Config{AllowWildcard: true, AllowOrigins: []string{"http://example.com/*"}},
			[][]string{{"http://example.com/", "*"}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("The code did not panic")
					}
				}()
			}
			result := tt.config.parseWildcardRules()
			if !tt.expectPanic && !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}
