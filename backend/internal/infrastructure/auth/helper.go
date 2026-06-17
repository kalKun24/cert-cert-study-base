package auth

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// authResponse は認証ミドルウェアで使用するレスポンスフォーマットです。
type authResponse struct {
	Error string `json:"error,omitempty"`
}

// writeJSONError はJSONエラーレスポンスを書き込むヘルパーです。
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(authResponse{Error: msg}); err != nil {
		slog.Error("JSONエンコードに失敗しました", "error", err)
	}
}

// contextWithValue はコンテキストにキーと値を格納するヘルパーです。
func contextWithValue(ctx context.Context, key, val any) context.Context {
	return context.WithValue(ctx, key, val)
}
