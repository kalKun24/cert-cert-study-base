package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// AuthHandler は認証に関するHTTPハンドラです。
type AuthHandler struct {
	authUC *usecase.AuthUseCase
}

// NewAuthHandler は AuthHandler を生成します。
func NewAuthHandler(authUC *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC: authUC}
}

// HandleLogin は POST /api/v1/auth/login を処理します。
// username/パスワードで認証し、JWTトークンを返します。
// - is_active: false のユーザーは 401 を返します。
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, response{Error: "リクエストボディが不正です"})
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, response{Error: "username と password は必須です"})
		return
	}

	out, err := h.authUC.Login(usecase.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrUserInactive):
			// ErrInvalidCredentials と ErrUserInactive を同一メッセージで返すことで
			// username 列挙攻撃（ユーザー存在確認・停止状態の推測）を防ぎます。
			writeJSON(w, http.StatusUnauthorized, response{Error: "usernameまたはパスワードが正しくありません"})
		default:
			slog.Error("ログイン処理でエラーが発生しました", "error", err)
			writeJSON(w, http.StatusInternalServerError, response{Error: "サーバー内部エラーが発生しました"})
		}
		return
	}

	writeJSON(w, http.StatusOK, response{
		Data: LoginResponseData{
			Token: out.Token,
			User:  toUserDTO(out.User),
		},
	})
}

// HandleLogout は POST /api/v1/auth/logout を処理します。
// サーバーサイドのトークン無効化リストは保持しません。
// クライアント側でトークンを削除することでセッションを終了します。
// このエンドポイントは将来的なサーバーサイド無効化への拡張口として提供します。
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		Data: map[string]string{
			"message": "ログアウトしました。クライアント側でトークンを削除してください。",
		},
	})
}

// writeJSON はJSONレスポンスを書き込むヘルパー関数です。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("JSONエンコードに失敗しました", "error", err)
	}
}
