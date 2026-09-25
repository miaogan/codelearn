package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ipWindow 单个 IP 在固定时间窗内的请求计数
type ipWindow struct {
	start time.Time
	count int
}

// rateLimiter 固定窗口内存限流器（按 IP，进程内有效）
type rateLimiter struct {
	mu    sync.Mutex
	limit int
	win   time.Duration
	hits  map[string]*ipWindow
}

func newRateLimiter(limit int, win time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, win: win, hits: map[string]*ipWindow{}}
}

func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()

	// 惰性清理过期窗口，防止 map 随不同 IP 无限增长
	if len(l.hits) > 1000 {
		for k, w := range l.hits {
			if now.Sub(w.start) >= l.win {
				delete(l.hits, k)
			}
		}
	}

	w, ok := l.hits[key]
	if !ok || now.Sub(w.start) >= l.win {
		l.hits[key] = &ipWindow{start: now, count: 1}
		return true
	}
	w.count++
	return w.count <= l.limit
}

// RateLimit 按 IP 固定窗口限流：窗口内超限返回 429
func RateLimit(limit int, win time.Duration) gin.HandlerFunc {
	l := newRateLimiter(limit, win)
	return func(c *gin.Context) {
		if !l.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			return
		}
		c.Next()
	}
}
