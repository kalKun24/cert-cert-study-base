package domain

import (
	"context"
	"errors"
	"time"
)

// Team はチームエンティティです。
type Team struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TeamMember はチームメンバーエンティティです。
type TeamMember struct {
	TeamID   string
	UserID   string
	JoinedAt time.Time
}

// チームドメインエラーの定義
var (
	ErrTeamNotFound          = errors.New("チームが見つかりません")
	ErrTeamNameAlreadyExists = errors.New("このチーム名は既に使用されています")
	ErrMemberAlreadyExists   = errors.New("このユーザーはすでにチームのメンバーです")
	ErrMemberNotFound        = errors.New("このユーザーはチームのメンバーではありません")
)

// TeamRepository はチームの永続化操作を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
type TeamRepository interface {
	// FindByID はIDでチームを検索します。
	// チームが存在しない場合は ErrTeamNotFound を返します。
	FindByID(ctx context.Context, id string) (*Team, error)

	// FindByName はチーム名でチームを検索します。
	// チームが存在しない場合は ErrTeamNotFound を返します。
	FindByName(ctx context.Context, name string) (*Team, error)

	// List は全チームの一覧を返します。
	List(ctx context.Context) ([]*Team, error)

	// ListByOwnerOrMember はユーザーがオーナーまたはメンバーであるチームを返します。
	ListByOwnerOrMember(ctx context.Context, userID string) ([]*Team, error)

	// Save はチームを新規作成または更新します。
	Save(ctx context.Context, team *Team) error

	// Delete はIDで指定したチームを削除します。
	// チームが存在しない場合は ErrTeamNotFound を返します。
	Delete(ctx context.Context, id string) error

	// AddMember はチームにメンバーを追加します。
	// 既にメンバーの場合は ErrMemberAlreadyExists を返します。
	AddMember(ctx context.Context, member *TeamMember) error

	// RemoveMember はチームからメンバーを除外します。
	// メンバーが存在しない場合は ErrMemberNotFound を返します。
	RemoveMember(ctx context.Context, teamID, userID string) error

	// ListMembers はチームのメンバー一覧を返します。
	ListMembers(ctx context.Context, teamID string) ([]*TeamMember, error)

	// IsMember はユーザーがチームのメンバーかどうかを返します。
	IsMember(ctx context.Context, teamID, userID string) (bool, error)
}
