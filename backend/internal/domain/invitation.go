package domain

import (
	"context"
	"errors"
	"time"
)

// InvitationStatus は招待のステータスを表す型です。
type InvitationStatus string

const (
	// StatusPending は未処理（招待中）を示します。
	StatusPending InvitationStatus = "pending"
	// StatusAccepted は受諾済みを示します。
	StatusAccepted InvitationStatus = "accepted"
	// StatusRejected は拒否済みを示します。
	StatusRejected InvitationStatus = "rejected"
)

// Invitation はチームへの招待エンティティです。
type Invitation struct {
	// ID は招待ID（UUID形式）
	ID string
	// TeamID は招待先チームID
	TeamID string
	// InvitedBy は招待したユーザーのID
	InvitedBy string
	// InviteeIdentifier は招待時に指定した識別子（UUID / ユーザー名 / メールアドレス）
	InviteeIdentifier string
	// InviteeUserID は解決済みの招待先ユーザーID（保存時に設定）
	InviteeUserID string
	// Status は招待のステータス（pending / accepted / rejected）
	Status InvitationStatus
	// CreatedAt は招待日時
	CreatedAt time.Time
}

// 招待ドメインエラーの定義
var (
	// ErrInvitationNotFound は招待が見つからない場合のエラーです。
	ErrInvitationNotFound = errors.New("招待が見つかりません")
	// ErrInvitationAlreadyExists は同一チームへの招待がすでに存在する場合のエラーです。
	ErrInvitationAlreadyExists = errors.New("同一チームへの招待がすでに存在します")
	// ErrInvitationNotPending は招待がすでに処理済みの場合のエラーです。
	ErrInvitationNotPending = errors.New("この招待はすでに処理済みです")
)

// InvitationRepository は招待の永続化操作を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
type InvitationRepository interface {
	// Save は招待を新規作成または更新します。
	Save(ctx context.Context, inv *Invitation) error

	// FindByID はIDで招待を検索します。
	// 招待が存在しない場合は ErrInvitationNotFound を返します。
	FindByID(ctx context.Context, id string) (*Invitation, error)

	// ListByInvitee は招待先ユーザーIDで招待一覧を返します。
	ListByInvitee(ctx context.Context, inviteeUserID string) ([]*Invitation, error)

	// ListByTeam はチームIDで招待一覧を返します。
	ListByTeam(ctx context.Context, teamID string) ([]*Invitation, error)

	// FindPendingByTeamAndInvitee は特定チーム・招待先ユーザーの pending 招待を返します。
	// 存在しない場合は ErrInvitationNotFound を返します。
	FindPendingByTeamAndInvitee(ctx context.Context, teamID, inviteeUserID string) (*Invitation, error)
}
