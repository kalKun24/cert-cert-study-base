// Package main はバックエンドサーバーのエントリーポイントです。
// HTTPサーバーを起動し、ヘルスチェックおよびAPI v1エンドポイントを提供します。
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

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/auth"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/repository"
	gcsStorage "github.com/kalKun24/cert-study-base/backend/internal/infrastructure/storage"
	"github.com/kalKun24/cert-study-base/backend/internal/interface/handler"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"

	gcs "cloud.google.com/go/storage"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET 環境変数が設定されていません")
		os.Exit(1)
	}

	gcsBucket := os.Getenv("GCS_BUCKET")
	if gcsBucket == "" {
		slog.Error("GCS_BUCKET 環境変数が設定されていません")
		os.Exit(1)
	}

	ctx := context.Background()

	// GCSクライアントを初期化
	gcsClient, err := gcs.NewClient(ctx)
	if err != nil {
		slog.Error("GCSクライアントの初期化に失敗しました", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := gcsClient.Close(); err != nil {
			slog.Warn("GCSクライアントのクローズに失敗しました", "error", err)
		}
	}()

	// StorageClient アダプター（GCS → 独自インターフェース）
	sc := gcsStorage.NewGCSStorageClient(gcsClient)

	// 依存関係を構築（コンポジションルート）
	userRepo := repository.NewGCSUserRepository(sc, gcsBucket)
	bcryptHasher := auth.NewBcryptHasher()
	jwtManager := auth.NewJWTManager(jwtSecret)

	authUC := usecase.NewAuthUseCase(userRepo, bcryptHasher, jwtManager)
	userUC := usecase.NewUserUseCase(userRepo, bcryptHasher)

	authHandler := handler.NewAuthHandler(authUC)
	userHandler := handler.NewUserHandler(userUC)

	// ルーティング設定
	mux := http.NewServeMux()

	// ヘルスチェック（認証不要）
	mux.HandleFunc("GET /health", handleHealth)

	// 認証エンドポイント（認証不要）
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.HandleLogin)

	// authMiddleware はDBからユーザーの最新 is_active を確認する認証ミドルウェアです。
	// トークン発行後に管理者がユーザーを停止した場合も即座に反映されます。
	authMiddleware := jwtManager.AuthMiddlewareWithRepo(userRepo)

	// ログアウト（JWTの検証のみ、ロール不問）
	mux.Handle("POST /api/v1/auth/logout",
		authMiddleware(http.HandlerFunc(authHandler.HandleLogout)),
	)

	// admin ロールのみ許可するミドルウェアチェーンを組み立てるヘルパー
	withAdmin := func(h http.HandlerFunc) http.Handler {
		return authMiddleware(
			auth.RequireRole(domain.RoleAdmin)(h),
		)
	}

	// ユーザー管理（admin のみ）
	mux.Handle("GET /api/v1/users", withAdmin(userHandler.HandleListUsers))
	mux.Handle("POST /api/v1/users", withAdmin(userHandler.HandleCreateUser))
	mux.Handle("GET /api/v1/users/{id}", withAdmin(userHandler.HandleGetUser))
	mux.Handle("PUT /api/v1/users/{id}", withAdmin(userHandler.HandleUpdateUser))
	mux.Handle("DELETE /api/v1/users/{id}", withAdmin(userHandler.HandleDeleteUser))
	mux.Handle("PATCH /api/v1/users/{id}/status", withAdmin(userHandler.HandleUpdateUserStatus))

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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
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
