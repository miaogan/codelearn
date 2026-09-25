package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimit_BlocksAfterLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/t", RateLimit(3, time.Minute), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	var last int
	for i := 0; i < 6; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		r.ServeHTTP(w, req)
		last = w.Code
	}
	if last != http.StatusTooManyRequests {
		t.Errorf("expected 429 after exceeding limit, got %d", last)
	}
}

func TestRateLimit_WindowResets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/t", RateLimit(2, 50*time.Millisecond), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	hit := func() int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		r.ServeHTTP(w, req)
		return w.Code
	}

	if code := hit(); code != http.StatusOK {
		t.Fatalf("first request should pass, got %d", code)
	}
	if code := hit(); code != http.StatusOK {
		t.Fatalf("second request should pass, got %d", code)
	}
	if code := hit(); code != http.StatusTooManyRequests {
		t.Fatalf("third request should be limited, got %d", code)
	}
	time.Sleep(60 * time.Millisecond)
	if code := hit(); code != http.StatusOK {
		t.Fatalf("after window reset should pass, got %d", code)
	}
}
