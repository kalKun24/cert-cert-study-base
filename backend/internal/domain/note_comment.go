// Package domain はビジネスエンティティとルールを定義します。
// このパッケージは外部ライブラリへの依存を持ちません。
package domain

import (
	"context"
	"errors"
	"time"
)

// NoteComment はノートコメントエンティティです。
// ノートに対して閲覧権限を持つユーザーが投稿するMarkdown形式のコメントです。
// domain.Comment（QuestionID フィールドを持つ）とは完全に独立した型です。
type NoteComment struct {
	// ID はコメントID（UUID形式）
	ID string
	// NoteID は対象のノートID
	NoteID string
	// Body はコメント本文（Markdown形式）
	Body string
	// CreatedBy は投稿者のユーザーID
	CreatedBy string
	// CreatedAt は投稿日時
	CreatedAt time.Time
	// UpdatedAt は更新日時
	UpdatedAt time.Time
}

// ドメインエラーの定義
var (
	// ErrNoteCommentNotFound はノートコメントが見つからない場合のエラーです。
	ErrNoteCommentNotFound = errors.New("ノートコメントが見つかりません")
)

// NoteCommentRepository はノートコメントの永続化操作を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
// 各メソッドはリクエストスコープの context.Context を受け取り、
// キャンセル・タイムアウトをGCS操作に伝播させます。
type NoteCommentRepository interface {
	// FindByID はIDでノートコメントを検索します。
	// コメントが存在しない場合は ErrNoteCommentNotFound を返します。
	FindByID(ctx context.Context, teamID, noteID, commentID string) (*NoteComment, error)

	// ListByNoteID は指定したチームIDとノートIDのコメント一覧を返します。
	// 返却順序は保証しません。投稿日時によるソートはユースケース層で実施します。
	ListByNoteID(ctx context.Context, teamID, noteID string) ([]*NoteComment, error)

	// Save はノートコメントを新規作成または更新します。
	// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
	Save(ctx context.Context, teamID string, comment *NoteComment) error

	// Delete はIDで指定したノートコメントを削除します。
	// コメントが存在しない場合は ErrNoteCommentNotFound を返します。
	Delete(ctx context.Context, teamID, noteID, commentID string) error
}
