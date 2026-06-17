package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kalKun24/cert-study-base/backend/internal/contextkey"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

// jwtTokenExpiry はJWTトークンの有効期限です。
const jwtTokenExpiry = 24 * time.Hour

// Claims はJWTのペイロード（クレーム）です。
type Claims struct {
	UserID string      `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager はJWTトークンの生成と検証を行います。
// usecase.TokenGenerator インターフェースを実装します。
type JWTManager struct {
	secretKey []byte
}

// NewJWTManager は JWTManager を生成します。
// secretKey は環境変数から取得した秘密鍵を渡してください。
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{secretKey: []byte(secretKey)}
}

// Generate はユーザー情報からJWTトークンを生成します。
func (m *JWTManager) Generate(user *domain.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtTokenExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("JWTトークンの署名に失敗しました: %w", err)
	}
	return signed, nil
}

// Parse はJWTトークンを検証し、クレームを返します。
func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("予期しない署名アルゴリズム: %v", t.Header["alg"])
		}
		return m.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("JWTトークンの検証に失敗しました: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("無効なJWTトークンです")
	}

	return claims, nil
}

// AuthMiddlewareWithRepo はJWTトークンを検証した後、DBからユーザーの最新状態を確認する
// 認証ミドルウェアです。
// Authorization: Bearer <token> ヘッダーからトークンを取得して検証します。
// - トークンが不正または期限切れの場合は 401 を返します。
// - DBからユーザーを取得できない場合は 401 を返します。
// - ユーザーが停止中（is_active: false）の場合は 403 を返します。
// トークン発行後に管理者がユーザーを停止した場合も、次のリクエストから即座に反映されます。
func (m *JWTManager) AuthMiddlewareWithRepo(repo domain.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "認証トークンがありません")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeJSONError(w, http.StatusUnauthorized, "Authorization ヘッダーの形式が不正です")
				return
			}

			claims, err := m.Parse(parts[1])
			if err != nil {
				slog.Warn("JWTトークン検証失敗", "error", err)
				writeJSONError(w, http.StatusUnauthorized, "トークンが無効または期限切れです")
				return
			}

			// DBからユーザーの最新状態を取得して is_active を確認する。
			// これにより、トークン発行後に管理者がユーザーを停止した場合も
			// 既存トークンの有効期限内に即座に反映される。
			// r.Context() を渡すことでリクエストのキャンセル・タイムアウトを伝播させる。
			user, err := repo.FindByID(r.Context(), claims.UserID)
			if err != nil {
				slog.Warn("認証ミドルウェアでユーザー取得に失敗しました", "user_id", claims.UserID, "error", err)
				writeJSONError(w, http.StatusUnauthorized, "トークンが無効または期限切れです")
				return
			}

			if !user.IsActive {
				writeJSONError(w, http.StatusForbidden, "このアカウントは停止中です")
				return
			}

			// 検証済みのユーザー情報をコンテキストに格納
			ctx := r.Context()
			ctx = contextWithValue(ctx, contextkey.UserID, user.ID)
			ctx = contextWithValue(ctx, contextkey.UserRole, string(user.Role))
			ctx = contextWithValue(ctx, contextkey.IsActive, user.IsActive)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole は指定ロールを持つユーザーのみ通過を許可する認可ミドルウェアです。
// 指定ロール以外のユーザーは 403 を返します。
// このミドルウェアは AuthMiddlewareWithRepo の後に適用してください。
func RequireRole(roles ...domain.Role) func(http.Handler) http.Handler {
	allowedRoles := make(map[domain.Role]struct{}, len(roles))
	for _, r := range roles {
		allowedRoles[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleStr, _ := r.Context().Value(contextkey.UserRole).(string)
			role := domain.Role(roleStr)

			if _, ok := allowedRoles[role]; !ok {
				writeJSONError(w, http.StatusForbidden, "この操作を行う権限がありません")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
