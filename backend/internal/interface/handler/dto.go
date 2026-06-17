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

// TeamDTO はAPIレスポンス用のチームDTOです。
type TeamDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TeamMemberDTO はAPIレスポンス用のチームメンバーDTOです。
type TeamMemberDTO struct {
	TeamID   string    `json:"team_id"`
	UserID   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

// TeamDetailDTO はチーム詳細のDTOです（メンバー一覧を含む）。
type TeamDetailDTO struct {
	TeamDTO
	Members []TeamMemberDTO `json:"members"`
}

// CreateTeamRequestDTO はチーム作成リクエストのDTOです。
type CreateTeamRequestDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateTeamRequestDTO はチーム更新リクエストのDTOです。
type UpdateTeamRequestDTO struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// AddTeamMemberRequestDTO はメンバー追加リクエストのDTOです。
type AddTeamMemberRequestDTO struct {
	UserID string `json:"user_id"`
}

// toTeamDTO はドメインエンティティをAPIレスポンス用DTOに変換します。
func toTeamDTO(t *domain.Team) TeamDTO {
	return TeamDTO{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		OwnerID:     t.OwnerID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// toTeamMemberDTO はドメインエンティティをAPIレスポンス用DTOに変換します。
func toTeamMemberDTO(m *domain.TeamMember) TeamMemberDTO {
	return TeamMemberDTO{
		TeamID:   m.TeamID,
		UserID:   m.UserID,
		JoinedAt: m.JoinedAt,
	}
}
