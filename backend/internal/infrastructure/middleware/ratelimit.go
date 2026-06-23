package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu       sync.Mutex
	limiters = make(map[string]*ipLimiter)
)

func init() {
	go cleanupLimiters()
}

func cleanupLimiters() {
	for range time.Tick(5 * time.Minute) {
		mu.Lock()
		for ip, l := range limiters {
			if time.Since(l.lastSeen) > 10*time.Minute {
				delete(limiters, ip)
			}
		}
		mu.Unlock()
	}
}

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	l, ok := limiters[ip]
	if !ok {
		// 10回/分 = 約 0.167 req/s、バースト 10
		l = &ipLimiter{limiter: rate.NewLimiter(rate.Every(6*time.Second), 10)}
		limiters[ip] = l
	}
	l.lastSeen = time.Now()
	return l.limiter
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// XFF は "IP1, IP2, ..." 形式。最初のアドレスが元クライアント。
		// Cloud Run 等のプロキシが末尾に自身の IP を追加するため、先頭要素のみ使用する。
		first := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
		if net.ParseIP(first) != nil {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// LoginRateLimit はログインエンドポイントに IP 単位のレート制限（10回/分）を適用するミドルウェアです。
func LoginRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !getLimiter(ip).Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":  nil,
				"error": "リクエストが多すぎます。しばらく待ってから再試行してください。",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
