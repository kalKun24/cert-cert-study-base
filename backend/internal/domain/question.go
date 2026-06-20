// Package domain はビジネスエンティティとルールを定義します。
// このパッケージは外部ライブラリへの依存を持ちません。
package domain

import (
	"context"
	"errors"
	"time"
)

// QuestionStatus は問題の公開ステータスを表す型です。
type QuestionStatus string

const (
	// QuestionStatusDraft は下書き状態です。作成時のデフォルト値です。
	QuestionStatusDraft QuestionStatus = "draft"
	// QuestionStatusPrivate は非公開状態です。
	QuestionStatusPrivate QuestionStatus = "private"
	// QuestionStatusPublished は公開状態です。
	QuestionStatusPublished QuestionStatus = "published"
)

// IsValid は QuestionStatus が有効な値かどうかを検証します。
func (s QuestionStatus) IsValid() bool {
	switch s {
	case QuestionStatusDraft, QuestionStatusPrivate, QuestionStatusPublished:
		return true
	default:
		return false
	}
}

// Question は問題エンティティです。
// 問題・解答・解説・議論点メモをMarkdown形式で保持します。
// 各問題は1つのチームに属し、チームメンバーのみアクセス可能です。
type Question struct {
	// ID は問題ID（UUID形式）
	ID string
	// TeamID は所属チームID
	TeamID string
	// Title は問題タイトル
	Title string
	// Body は問題文（Markdown形式）
	Body string
	// Answer は解答（Markdown形式）
	Answer string
	// Explanation は解説（Markdown形式）
	Explanation string
	// Memo は議論点メモ（Markdown形式）
	Memo string
	// Tags はタグ（フラット・複数付与可）
	Tags []string
	// Status は公開ステータス（draft / private / published）
	Status QuestionStatus
	// CreatedBy は作成者のユーザーID
	CreatedBy string
	// CreatedAt は作成日時
	CreatedAt time.Time
	// UpdatedAt は更新日時
	UpdatedAt time.Time
}

// ドメインエラーの定義
var (
	// ErrQuestionNotFound は問題が見つからない場合のエラーです。
	ErrQuestionNotFound = errors.New("問題が見つかりません")
	// ErrInvalidQuestionStatus は問題のステータスが無効な値の場合のエラーです。
	ErrInvalidQuestionStatus = errors.New("無効な問題ステータスです")
)

// QuestionSearchFilter は問題検索・フィルタリング条件を表します。
type QuestionSearchFilter struct {
	// TagIDs はAND絞り込みするタグIDの一覧。空の場合はタグフィルタなし。
	TagIDs []string
	// Keyword はタイトル・問題文・解説・メモを対象としたキーワード検索（部分一致）。空の場合は検索なし。
	Keyword string
}

// QuestionRepository は問題の永続化操作を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
// 各メソッドはリクエストスコープの context.Context を受け取り、
// キャンセル・タイムアウトをGCS操作に伝播させます。
type QuestionRepository interface {
	// FindByID はIDで問題を検索します。
	// 問題が存在しない場合は ErrQuestionNotFound を返します。
	FindByID(ctx context.Context, id string) (*Question, error)

	// ListByTeam は指定チームの問題一覧を返します。
	ListByTeam(ctx context.Context, teamID string) ([]*Question, error)

	// SearchByTeam は指定チームの問題を検索・フィルタリングして返します。
	// フィルタ条件が空の場合はチーム全件を返します。
	SearchByTeam(ctx context.Context, teamID string, filter QuestionSearchFilter) ([]*Question, error)

	// Save は問題を新規作成または更新します。
	// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
	Save(ctx context.Context, question *Question) error

	// FindByTagID は指定されたタグIDを持つ問題の一覧を返します。
	// 該当する問題が存在しない場合は空のスライスを返します。
	FindByTagID(ctx context.Context, tagID string) ([]*Question, error)

	// Delete はIDで指定した問題を削除します。
	// 問題が存在しない場合は ErrQuestionNotFound を返します。
	Delete(ctx context.Context, id string) error
}
