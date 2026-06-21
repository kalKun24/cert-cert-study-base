// Package domain はビジネスエンティティとルールを定義します。
// このパッケージは外部ライブラリへの依存を持ちません。
package domain

import (
	"context"
	"errors"
	"time"
)

// NoteStatus はノートの公開ステータスを表す型です。
type NoteStatus string

const (
	// NoteStatusDraft は下書き状態です。作成時のデフォルト値です。
	NoteStatusDraft NoteStatus = "draft"
	// NoteStatusPrivate は非公開状態です。
	NoteStatusPrivate NoteStatus = "private"
	// NoteStatusPublished は公開状態です。
	NoteStatusPublished NoteStatus = "published"
)

// IsValid は NoteStatus が有効な値かどうかを検証します。
func (s NoteStatus) IsValid() bool {
	switch s {
	case NoteStatusDraft, NoteStatusPrivate, NoteStatusPublished:
		return true
	default:
		return false
	}
}

// Note は知識共有ノートエンティティです。
// 本文・議論点・メモをMarkdown形式で保持します。
// 各ノートは1つのチームに属し、チームメンバーのみアクセス可能です。
// GCSパス: teams/{team_id}/notes.json
type Note struct {
	// ID はノートID（UUID形式）
	ID string
	// TeamID は所属チームID
	TeamID string
	// Title はノートタイトル
	Title string
	// Body は本文（Markdown形式）
	Body string
	// DiscussionPoints は議論点（Markdown形式）
	DiscussionPoints string
	// Memo は自由記述メモ（Markdown形式）
	Memo string
	// Tags はタグ（フラット・複数付与可）
	Tags []string
	// Status は公開ステータス（draft / private / published）
	Status NoteStatus
	// CreatedBy は作成者のユーザーID
	CreatedBy string
	// CreatedAt は作成日時
	CreatedAt time.Time
	// UpdatedAt は更新日時
	UpdatedAt time.Time
}

// ドメインエラーの定義
var (
	// ErrNoteNotFound はノートが見つからない場合のエラーです。
	ErrNoteNotFound = errors.New("ノートが見つかりません")
	// ErrInvalidNoteStatus はノートのステータスが無効な値の場合のエラーです。
	ErrInvalidNoteStatus = errors.New("無効なノートステータスです")
)

// NoteSearchFilter はノート検索・フィルタリング条件を表します。
type NoteSearchFilter struct {
	// TagIDs はAND絞り込みするタグIDの一覧。空の場合はタグフィルタなし。
	TagIDs []string
	// Keyword はタイトル・本文・議論点・メモを対象としたキーワード検索（部分一致）。空の場合は検索なし。
	Keyword string
}

// NoteRepository はノートの永続化操作を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
// GCSパス: teams/{team_id}/notes.json
// 各メソッドはリクエストスコープの context.Context を受け取り、
// キャンセル・タイムアウトをGCS操作に伝播させます。
type NoteRepository interface {
	// FindByID はチームIDとIDでノートを検索します。
	// ノートが存在しない場合は ErrNoteNotFound を返します。
	FindByID(ctx context.Context, teamID, id string) (*Note, error)

	// ListByTeam は指定チームのノート一覧を返します。
	ListByTeam(ctx context.Context, teamID string) ([]*Note, error)

	// SearchByTeam は指定チームのノートを検索・フィルタリングして返します。
	// フィルタ条件が空の場合はチーム全件を返します。
	SearchByTeam(ctx context.Context, teamID string, filter NoteSearchFilter) ([]*Note, error)

	// Save はノートを新規作成または更新します。
	// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
	Save(ctx context.Context, note *Note) error

	// Delete はチームIDとIDで指定したノートを削除します。
	// ノートが存在しない場合は ErrNoteNotFound を返します。
	Delete(ctx context.Context, teamID, id string) error

	// FindByTagID は指定チームの指定タグIDを持つノートの一覧を返します。
	// 該当するノートが存在しない場合は空のスライスを返します。
	FindByTagID(ctx context.Context, teamID, tagID string) ([]*Note, error)
}
