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

// VisibilityScope は問題の公開範囲を表す型です。
type VisibilityScope string

const (
	// QuestionStatusDraft は下書き状態です。作成時のデフォルト値です。
	QuestionStatusDraft QuestionStatus = "draft"
	// QuestionStatusPrivate は非公開状態です。
	QuestionStatusPrivate QuestionStatus = "private"
	// QuestionStatusPublished は公開状態です。
	QuestionStatusPublished QuestionStatus = "published"

	// VisibilityScopeAll は全ユーザーに公開することを示します。作成時のデフォルト値です。
	VisibilityScopeAll VisibilityScope = "all"
	// VisibilityScopeTeam は特定チームのみに公開することを示します。
	VisibilityScopeTeam VisibilityScope = "team"
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

// IsValid は VisibilityScope が有効な値かどうかを検証します。
func (v VisibilityScope) IsValid() bool {
	switch v {
	case VisibilityScopeAll, VisibilityScopeTeam:
		return true
	default:
		return false
	}
}

// Question は問題エンティティです。
// 問題・解答・解説・議論点メモをMarkdown形式で保持します。
type Question struct {
	// ID は問題ID（UUID形式）
	ID string
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
	// VisibilityScope は公開範囲（all / team）
	VisibilityScope VisibilityScope
	// PublishedTeamIDs は公開対象チームIDの一覧（VisibilityScope=team の場合に使用）
	PublishedTeamIDs []string
	// CreatedBy は作成者のユーザーID
	CreatedBy string
	// CreatedAt は作成日時
	CreatedAt time.Time
	// UpdatedAt は更新日時
	UpdatedAt time.Time
}

// IsVisibleTo は指定したユーザーがこの問題を閲覧できるかどうかを返します。
// callerID はリクエストユーザーのID、isAdmin は admin ロールかどうか、
// callerTeamIDs はリクエストユーザーが所属するチームIDの集合です。
func (q *Question) IsVisibleTo(callerID string, isAdmin bool, callerTeamIDs map[string]struct{}) bool {
	// admin は全件閲覧可能
	if isAdmin {
		return true
	}

	switch q.Status {
	case QuestionStatusDraft, QuestionStatusPrivate:
		// draft / private は作成者本人のみ
		return q.CreatedBy == callerID
	case QuestionStatusPublished:
		switch q.VisibilityScope {
		case VisibilityScopeAll:
			// 全ログインユーザーに公開
			return true
		case VisibilityScopeTeam:
			// 作成者本人は常に閲覧可能
			if q.CreatedBy == callerID {
				return true
			}
			// published_team_ids のいずれかに所属するユーザーに公開
			for _, teamID := range q.PublishedTeamIDs {
				if _, ok := callerTeamIDs[teamID]; ok {
					return true
				}
			}
			return false
		default:
			// 未知の visibility_scope は安全サイドに倒して完全拒否
			return false
		}
	}
	// status が未設定（後方互換）の場合は draft 扱いで作成者のみ
	return q.CreatedBy == callerID
}

// ドメインエラーの定義
var (
	// ErrQuestionNotFound は問題が見つからない場合のエラーです。
	ErrQuestionNotFound = errors.New("問題が見つかりません")
	// ErrInvalidQuestionStatus は問題のステータスが無効な値の場合のエラーです。
	ErrInvalidQuestionStatus = errors.New("無効な問題ステータスです")
	// ErrInvalidVisibilityScope は公開範囲が無効な値の場合のエラーです。
	ErrInvalidVisibilityScope = errors.New("無効な公開範囲です")
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

	// List は全問題の一覧を返します。
	List(ctx context.Context) ([]*Question, error)

	// Search は指定したフィルタ条件に合致する問題の一覧を返します。
	// フィルタ条件が空の場合は全件を返します。
	Search(ctx context.Context, filter QuestionSearchFilter) ([]*Question, error)

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
