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
	IsTeamOwner bool      `json:"is_team_owner"`
	MaxTeams    int       `json:"max_teams"`
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
		IsTeamOwner: u.IsTeamOwner,
		MaxTeams:    u.MaxTeams,
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

// UpdateMyProfileRequestDTO はプロフィール編集リクエストのDTOです。
type UpdateMyProfileRequestDTO struct {
	DisplayName string `json:"display_name"`
}

// ChangeMyPasswordRequestDTO はパスワード変更リクエストのDTOです。
type ChangeMyPasswordRequestDTO struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// QuestionDTO はAPIレスポンス用の問題DTOです。
type QuestionDTO struct {
	ID          string    `json:"id"`
	TeamID      string    `json:"team_id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Answer      string    `json:"answer"`
	Explanation string    `json:"explanation"`
	Memo        string    `json:"memo"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toQuestionDTO はドメインエンティティをAPIレスポンス用DTOに変換します。
func toQuestionDTO(q *domain.Question) QuestionDTO {
	tags := q.Tags
	if tags == nil {
		tags = []string{}
	}
	return QuestionDTO{
		ID:          q.ID,
		TeamID:      q.TeamID,
		Title:       q.Title,
		Body:        q.Body,
		Answer:      q.Answer,
		Explanation: q.Explanation,
		Memo:        q.Memo,
		Tags:        tags,
		Status:      string(q.Status),
		CreatedBy:   q.CreatedBy,
		CreatedAt:   q.CreatedAt,
		UpdatedAt:   q.UpdatedAt,
	}
}

// QuestionListResponseDTO は問題一覧APIのレスポンス用DTOです（ページネーション情報付き）。
type QuestionListResponseDTO struct {
	Items      []QuestionDTO `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

// CreateQuestionRequestDTO は問題作成リクエストのDTOです。
type CreateQuestionRequestDTO struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
	Memo        string   `json:"memo"`
	Tags        []string `json:"tags"`
	Status      string   `json:"status"`
}

// UpdateQuestionRequestDTO は問題更新リクエストのDTOです。
// 各フィールドはポインタ型にしてゼロ値との区別を可能にします。
type UpdateQuestionRequestDTO struct {
	Title       *string  `json:"title"`
	Body        *string  `json:"body"`
	Answer      *string  `json:"answer"`
	Explanation *string  `json:"explanation"`
	Memo        *string  `json:"memo"`
	Tags        []string `json:"tags"`
	Status      *string  `json:"status"`
}

// UpdateQuestionVisibilityRequestDTO は問題公開設定変更リクエストのDTOです。
type UpdateQuestionVisibilityRequestDTO struct {
	// Status は変更後の公開ステータス（必須）
	Status string `json:"status"`
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
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// TeamDetailDTO はチーム詳細のDTOです（メンバー一覧を含む）。
type TeamDetailDTO struct {
	TeamDTO
	Members []TeamMemberDTO `json:"members"`
}

// TeamMemberStatsDTO はチームメンバーの統計情報DTOです。
// GET /api/v1/teams/{id}/members のレスポンス用に使用します。
type TeamMemberStatsDTO struct {
	UserID        string     `json:"user_id"`
	DisplayName   string     `json:"display_name"`
	Role          string     `json:"role"`
	IsTeamOwner   bool       `json:"is_team_owner"`
	QuestionCount int        `json:"question_count"`
	CommentCount  int        `json:"comment_count"`
	LastLoginAt   *time.Time `json:"last_login_at"`
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

// UpdateTeamOwnerStatusRequestDTO はグローバルチームオーナー権限更新リクエストのDTOです。
type UpdateTeamOwnerStatusRequestDTO struct {
	IsTeamOwner bool `json:"is_team_owner"`
	MaxTeams    int  `json:"max_teams"`
}

// ChangeMemberRoleRequestDTO はチームメンバーロール変更リクエストのDTOです。
type ChangeMemberRoleRequestDTO struct {
	Role string `json:"role"`
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
		Role:     string(m.Role),
		JoinedAt: m.JoinedAt,
	}
}

// TagDTO はAPIレスポンス用のタグDTOです。
type TagDTO struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// toTagDTO はドメインエンティティをAPIレスポンス用DTOに変換します。
func toTagDTO(t *domain.Tag) TagDTO {
	return TagDTO{
		ID:        t.ID,
		TeamID:    t.TeamID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}
}

// CreateTagRequestDTO はタグ作成リクエストのDTOです。
type CreateTagRequestDTO struct {
	Name string `json:"name"`
}

// UpdateTagRequestDTO はタグ更新リクエストのDTOです。
// 各フィールドはポインタ型にしてゼロ値との区別を可能にします。
type UpdateTagRequestDTO struct {
	Name *string `json:"name"`
}

// CommentDTO はAPIレスポンス用のコメントDTOです。
// 投稿者の display_name を含みます。
type CommentDTO struct {
	ID          string    `json:"id"`
	QuestionID  string    `json:"question_id"`
	Body        string    `json:"body"`
	CreatedBy   string    `json:"created_by"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toCommentDTO はドメインエンティティと表示名をAPIレスポンス用DTOに変換します。
func toCommentDTO(c *domain.Comment, displayName string) CommentDTO {
	return CommentDTO{
		ID:          c.ID,
		QuestionID:  c.QuestionID,
		Body:        c.Body,
		CreatedBy:   c.CreatedBy,
		DisplayName: displayName,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// CreateCommentRequestDTO はコメント投稿リクエストのDTOです。
type CreateCommentRequestDTO struct {
	Body string `json:"body"`
}

// UpdateCommentRequestDTO はコメント編集リクエストのDTOです。
type UpdateCommentRequestDTO struct {
	Body string `json:"body"`
}

// InvitationDTO はAPIレスポンス用の招待DTOです。
// invitee_identifier はメールアドレス等の個人情報を含む可能性があるためレスポンスに含めません。
type InvitationDTO struct {
	ID            string    `json:"id"`
	TeamID        string    `json:"team_id"`
	InvitedBy     string    `json:"invited_by"`
	InviteeUserID string    `json:"invitee_user_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// SendInvitationRequestDTO は招待送信リクエストのDTOです。
type SendInvitationRequestDTO struct {
	InviteeIdentifier string `json:"invitee_identifier"`
}

// RespondInvitationRequestDTO は招待受諾/拒否リクエストのDTOです。
type RespondInvitationRequestDTO struct {
	// Status は "accepted" または "rejected" を指定する
	Status string `json:"status"`
}

// toInvitationDTO はドメインエンティティをAPIレスポンス用DTOに変換します。
func toInvitationDTO(inv *domain.Invitation) InvitationDTO {
	return InvitationDTO{
		ID:            inv.ID,
		TeamID:        inv.TeamID,
		InvitedBy:     inv.InvitedBy,
		InviteeUserID: inv.InviteeUserID,
		Status:        string(inv.Status),
		CreatedAt:     inv.CreatedAt,
	}
}
