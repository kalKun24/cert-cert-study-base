// Package domain はビジネスエンティティとルールを定義します。
// このパッケージは外部ライブラリへの依存を持ちません。
package domain

import (
	"context"
	"errors"
	"time"
)

// Role はユーザーのロールを表す型です。
type Role string

const (
	// RoleAdmin は全機能・全チーム管理・ユーザー停止ができる管理者ロールです。
	RoleAdmin Role = "admin"
	// RoleTeamOwner はチーム作成権限を持つグローバルロールです。
	// 廃止予定: チーム作成権限は User.IsTeamOwner フラグで管理するため、
	// このロールは後方互換のために残しているが、新規ユーザーには付与しない。
	// 詳細は TICKET-035 を参照。今後のフォローアップで削除予定。
	RoleTeamOwner Role = "teamowner"
	// RoleUser はチーム参加・自身の問題CRUDができる一般ユーザーロールです。
	RoleUser Role = "user"
)

// IsValid はロールが有効な値かどうかを検証します。
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleTeamOwner, RoleUser:
		return true
	default:
		return false
	}
}

// User はユーザーエンティティです。
// パスワードはbcryptハッシュ済みの値のみを保持し、平文は一切保存しません。
type User struct {
	// ID はユーザーID（UUID形式）
	ID string
	// Username はログインID（一意）
	Username string
	// DisplayName はUI表示名
	DisplayName string
	// Email はメールアドレス（一意）
	Email string
	// PasswordHash はbcryptハッシュ済みパスワード（平文保存禁止）
	PasswordHash string
	// Role はユーザーのロール（admin / teamowner / user）
	Role Role
	// IsActive は有効ユーザーかどうか（false = 停止中）
	IsActive bool
	// IsTeamOwner はチームを作成できるグローバル権限（admin が付与）
	IsTeamOwner bool
	// MaxTeams は作成可能なチーム数の上限（0 = 制限なし）
	MaxTeams int
	// CreatedAt は作成日時
	CreatedAt time.Time
	// UpdatedAt は更新日時
	UpdatedAt time.Time
}

// ドメインエラーの定義
var (
	// ErrUserNotFound はユーザーが見つからない場合のエラーです。
	ErrUserNotFound = errors.New("ユーザーが見つかりません")
	// ErrUserInactive はユーザーが停止中の場合のエラーです。
	ErrUserInactive = errors.New("ユーザーは停止中です")
	// ErrUsernameAlreadyExists はusernameが既に使用されている場合のエラーです。
	ErrUsernameAlreadyExists = errors.New("このusernameは既に使用されています")
	// ErrEmailAlreadyExists はメールアドレスが既に使用されている場合のエラーです。
	ErrEmailAlreadyExists = errors.New("このメールアドレスは既に使用されています")
	// ErrInvalidCredentials は認証情報が不正な場合のエラーです。
	ErrInvalidCredentials = errors.New("usernameまたはパスワードが正しくありません")
	// ErrInvalidRole はロールが無効な値の場合のエラーです。
	ErrInvalidRole = errors.New("無効なロールです")
	// ErrPermissionDenied は操作に必要な権限がない場合のエラーです。
	ErrPermissionDenied = errors.New("この操作を行う権限がありません")
	// ErrCurrentPasswordIncorrect は現在のパスワードが正しくない場合のエラーです。
	ErrCurrentPasswordIncorrect = errors.New("current password is incorrect")
)

// UserRepository はユーザーの永続化操作を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
// 各メソッドはリクエストスコープの context.Context を受け取り、
// キャンセル・タイムアウトをGCS操作に伝播させます。
type UserRepository interface {
	// FindByID はIDでユーザーを検索します。
	// ユーザーが存在しない場合は ErrUserNotFound を返します。
	FindByID(ctx context.Context, id string) (*User, error)

	// FindByUsername はusernameでユーザーを検索します。
	// ユーザーが存在しない場合は ErrUserNotFound を返します。
	FindByUsername(ctx context.Context, username string) (*User, error)

	// FindByEmail はメールアドレスでユーザーを検索します。
	// ユーザーが存在しない場合は ErrUserNotFound を返します。
	FindByEmail(ctx context.Context, email string) (*User, error)

	// List は全ユーザーの一覧を返します。
	List(ctx context.Context) ([]*User, error)

	// Save はユーザーを新規作成または更新します。
	Save(ctx context.Context, user *User) error

	// Delete はIDで指定したユーザーを削除します。
	// ユーザーが存在しない場合は ErrUserNotFound を返します。
	Delete(ctx context.Context, id string) error
}
