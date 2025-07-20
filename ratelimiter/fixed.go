package ratelimiter

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type FixedWindowLimiter struct {
	Redis      *redis.Client
	Limit      int
	WindowSize time.Duration
	KeyPrefix  string
}

func NewFixedWindowRateLimiter(r *redis.Client, limit int, windowSize time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		Redis:      r,
		Limit:      limit,
		WindowSize: windowSize,
		KeyPrefix:  "ratelimiter:v1",
	}
}

func (f *FixedWindowLimiter) MiddleWare() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)
			currentWindow := time.Now().Unix() / int64(f.WindowSize.Seconds())
			key := fmt.Sprintf("%s:%s:%d", f.KeyPrefix, ip, currentWindow)
			count, err := f.Redis.Incr(context.Background(), key).Result()
			if err != nil {
				http.Error(w, "Redis error", http.StatusInternalServerError)
				return
			}
			if count == 1 {
				// first time
				f.Redis.Expire(context.Background(), key, f.WindowSize)
			}
			if count > int64(f.Limit) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}

func getIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return strings.Split(xff, ",")[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
