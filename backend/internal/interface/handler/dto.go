// Package handler はHTTPハンドラとDTOを定義します。
// このパッケージはHTTPリクエスト/レスポンスのデータ変換を担当します。
package handler

import (
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

// response は統一レスポンスフォーマットです。
type response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// UserDTO はAPIレスポンス用のユーザーDTOです。
// パスワードハッシュは含みません。
type UserDTO struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toUserDTO はドメインエンティティをAPIレスポンス用DTOに変換します。
func toUserDTO(u *domain.User) UserDTO {
	return UserDTO{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		Role:        string(u.Role),
		IsActive:    u.IsActive,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// LoginRequestDTO はログインリクエストのDTOです。
type LoginRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponseData はログインレスポンスのdataフィールドです。
type LoginResponseData struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// CreateUserRequestDTO はユーザー作成リクエストのDTOです。
type CreateUserRequestDTO struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
}

// UpdateUserRequestDTO はユーザー更新リクエストのDTOです。
// 各フィールドはポインタ型にしてゼロ値との区別を可能にします。
type UpdateUserRequestDTO struct {
	DisplayName *string `json:"display_name"`
	Email       *string `json:"email"`
	Role        *string `json:"role"`
	Password    *string `json:"password"`
}

// UpdateUserStatusRequestDTO はユーザーステータス更新リクエストのDTOです。
type UpdateUserStatusRequestDTO struct {
	IsActive bool `json:"is_active"`
}
