// Package main はバックエンドサーバーのエントリーポイントです。
// HTTPサーバーを起動し、ヘルスチェックエンドポイントを提供します。
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// response は統一レスポンスフォーマットです。
type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func main() {
	// 構造化ログ（JSON形式）をセットアップ
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// ポート番号として有効な範囲（1〜65535）を検証
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		slog.Error("PORT 環境変数が無効です", "port", port)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)

	srv := &http.Server{
		Addr:         net.JoinHostPort("", port),
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// グレースフルシャットダウンのためにシグナルを待機
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		slog.Info("シャットダウンシグナルを受信しました。サーバーを停止します。")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("サーバーのシャットダウンに失敗しました", "error", err)
		}
		close(idleConnsClosed)
	}()

	slog.Info("サーバーを起動します", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("サーバーの起動に失敗しました", "error", err)
		os.Exit(1)
	}

	<-idleConnsClosed
	slog.Info("サーバーを正常に停止しました")
}

// handleHealth はヘルスチェックエンドポイントのハンドラです。
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, response{
		Data: map[string]string{
			"status": "ok",
		},
	})
}

// loggingMiddleware はリクエストの構造化ログを出力するミドルウェアです。
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		slog.Info("HTTPリクエスト",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"latency_ms", time.Since(start).Milliseconds(),
		)
	})
}

// responseWriter はHTTPステータスコードを記録するResponseWriterのラッパーです。
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// writeJSON はJSONレスポンスを書き込むヘルパー関数です。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("JSONエンコードに失敗しました", "error", err)
	}
}
