## Rate Limiter Package for Go

### Example Usage

```go
package main

import (
	"net/http"
	"time"

	"github.com/inblack67/ratelimiter/ratelimiter"
	"github.com/redis/go-redis/v9"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	var res = []byte("hello worlds")
	w.Write(res)
}

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	limiter := ratelimiter.NewFixedWindowRateLimiter(redisClient, 5, 60*time.Second)
	rateLimitedHandler := limiter.MiddleWare()(http.HandlerFunc(defaultHandler))
	http.Handle("/", rateLimitedHandler)
	http.ListenAndServe(":3000", nil)
}

```
