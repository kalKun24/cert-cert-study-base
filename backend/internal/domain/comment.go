// Package domain はビジネスエンティティとルールを定義します。
// このパッケージは外部ライブラリへの依存を持ちません。
package domain

import (
	"context"
	"errors"
	"time"
)

// Comment はコメントエンティティです。
// 公開済み問題に対して閲覧権限を持つユーザーが投稿するMarkdown形式のコメントです。
type Comment struct {
	// ID はコメントID（UUID形式）
	ID string
	// QuestionID は対象の問題ID
	QuestionID string
	// Body はコメント本文（Markdown形式）
	Body string
	// CreatedBy は投稿者のユーザーID
	CreatedBy string
	// CreatedAt は投稿日時
	CreatedAt time.Time
	// UpdatedAt は更新日時
	UpdatedAt time.Time
}

// MaxCommentBodyLength はコメント本文の最大文字数（rune単位）です。
const MaxCommentBodyLength = 10000

// ドメインエラーの定義
var (
	// ErrCommentNotFound はコメントが見つからない場合のエラーです。
	ErrCommentNotFound = errors.New("コメントが見つかりません")
	// ErrCommentBodyEmpty はコメント本文が空の場合のエラーです。
	ErrCommentBodyEmpty = errors.New("コメント本文は必須です")
	// ErrCommentBodyTooLong はコメント本文が最大文字数を超えた場合のエラーです。
	ErrCommentBodyTooLong = errors.New("コメント本文が長すぎます")
)

// CommentRepository はコメントの永続化操作を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
// 各メソッドはリクエストスコープの context.Context を受け取り、
// キャンセル・タイムアウトをGCS操作に伝播させます。
type CommentRepository interface {
	// FindByID はIDでコメントを検索します。
	// コメントが存在しない場合は ErrCommentNotFound を返します。
	FindByID(ctx context.Context, teamID, questionID, commentID string) (*Comment, error)

	// ListByQuestionID は指定したチームIDと問題IDのコメント一覧を投稿日時の昇順で返します。
	ListByQuestionID(ctx context.Context, teamID, questionID string) ([]*Comment, error)

	// Save はコメントを新規作成または更新します。
	// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
	Save(ctx context.Context, teamID string, comment *Comment) error

	// Delete はIDで指定したコメントを削除します。
	// コメントが存在しない場合は ErrCommentNotFound を返します。
	Delete(ctx context.Context, teamID, questionID, commentID string) error
}
