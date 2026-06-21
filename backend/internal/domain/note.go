// Package domain はビジネスエンティティとルールを定義します。
// このパッケージは外部ライブラリへの依存を持ちません。
package domain

import (
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
