// Package domain はビジネスエンティティとルールを定義します。
// このパッケージは外部ライブラリへの依存を持ちません。
package domain

import (
	"context"
	"errors"
	"time"
)

// Tag はタグエンティティです。
// 問題に付与するフラット構造のタグを表します。
type Tag struct {
	// ID はタグID（UUID形式）
	ID string
	// Name はタグ名（一意）
	Name string
	// CreatedAt は作成日時
	CreatedAt time.Time
}

// ドメインエラーの定義
var (
	// ErrTagNotFound はタグが見つからない場合のエラーです。
	ErrTagNotFound = errors.New("タグが見つかりません")
	// ErrTagNameAlreadyExists はタグ名が既に使用されている場合のエラーです。
	ErrTagNameAlreadyExists = errors.New("このタグ名は既に使用されています")
	// ErrTagInUse はタグが問題に使用中で削除できない場合のエラーです。
	ErrTagInUse = errors.New("このタグは使用中のため削除できません")
)

// TagRepository はタグの永続化操作を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
// 各メソッドはリクエストスコープの context.Context を受け取り、
// キャンセル・タイムアウトをGCS操作に伝播させます。
type TagRepository interface {
	// FindByID はIDでタグを検索します。
	// タグが存在しない場合は ErrTagNotFound を返します。
	FindByID(ctx context.Context, id string) (*Tag, error)

	// FindByName はタグ名でタグを検索します。
	// タグが存在しない場合は ErrTagNotFound を返します。
	FindByName(ctx context.Context, name string) (*Tag, error)

	// List は全タグの一覧を返します。
	List(ctx context.Context) ([]*Tag, error)

	// Save はタグを新規作成または更新します。
	// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
	Save(ctx context.Context, tag *Tag) error

	// Delete はIDで指定したタグを削除します。
	// タグが存在しない場合は ErrTagNotFound を返します。
	Delete(ctx context.Context, id string) error
}
