// Package usecase はビジネスロジック（ユースケース）を実装します。
// このパッケージは domain パッケージのみに依存します。
package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

const (
	// tagNameMaxLength はタグ名の最大長です。
	tagNameMaxLength = 50
)

// TagUseCase はタグ管理に関するユースケースを実装します。
type TagUseCase struct {
	tagRepo domain.TagRepository
}

// NewTagUseCase は TagUseCase を生成します（コンストラクタインジェクション）。
// 使用中チェックは TagRepository の Delete 実装（インフラ層）に委譲します。
func NewTagUseCase(tagRepo domain.TagRepository) *TagUseCase {
	return &TagUseCase{
		tagRepo: tagRepo,
	}
}

// ListTags は全タグの一覧を取得します。
// 認証済みユーザーであれば誰でも取得可能です。
func (uc *TagUseCase) ListTags(ctx context.Context) ([]*domain.Tag, error) {
	tags, err := uc.tagRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("タグ一覧の取得に失敗しました: %w", err)
	}
	return tags, nil
}

// CreateTagInput はタグ作成ユースケースの入力です。
type CreateTagInput struct {
	// Name はタグ名（必須・一意）
	Name string
}

// CreateTag は新しいタグを作成します。
// admin ロールのみ実行可能（認可はハンドラ層のミドルウェアで保証）。
// タグ名の重複は ErrTagNameAlreadyExists を返します。
func (uc *TagUseCase) CreateTag(ctx context.Context, input CreateTagInput) (*domain.Tag, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("タグ名は必須です")
	}
	if len([]rune(name)) > tagNameMaxLength {
		return nil, fmt.Errorf("タグ名は%d文字以内で入力してください", tagNameMaxLength)
	}
	input.Name = name

	// タグ名の重複チェック
	if _, err := uc.tagRepo.FindByName(ctx, input.Name); err == nil {
		return nil, domain.ErrTagNameAlreadyExists
	}

	now := time.Now().UTC()
	tag := &domain.Tag{
		ID:        uuid.NewString(),
		Name:      input.Name,
		CreatedAt: now,
	}

	if err := uc.tagRepo.Save(ctx, tag); err != nil {
		return nil, fmt.Errorf("タグの保存に失敗しました: %w", err)
	}

	return tag, nil
}

// UpdateTagInput はタグ更新ユースケースの入力です。
type UpdateTagInput struct {
	// Name はタグ名（nil の場合は変更しない）
	Name *string
}

// UpdateTag は指定IDのタグを更新します。
// admin ロールのみ実行可能（認可はハンドラ層のミドルウェアで保証）。
// タグ名の重複は ErrTagNameAlreadyExists を返します。
func (uc *TagUseCase) UpdateTag(ctx context.Context, id string, input UpdateTagInput) (*domain.Tag, error) {
	tag, err := uc.tagRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("タグの取得に失敗しました: %w", err)
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, fmt.Errorf("タグ名は必須です")
		}
		if len([]rune(name)) > tagNameMaxLength {
			return nil, fmt.Errorf("タグ名は%d文字以内で入力してください", tagNameMaxLength)
		}
		// タグ名の重複チェック（自分自身は除外）
		existing, err := uc.tagRepo.FindByName(ctx, name)
		if err == nil && existing.ID != id {
			return nil, domain.ErrTagNameAlreadyExists
		}
		tag.Name = name
	}

	if err := uc.tagRepo.Save(ctx, tag); err != nil {
		return nil, fmt.Errorf("タグの保存に失敗しました: %w", err)
	}

	return tag, nil
}

// DeleteTag は指定IDのタグを削除します。
// admin ロールのみ実行可能（認可はハンドラ層のミドルウェアで保証）。
// 使用中チェックは TagRepository の Delete 実装（インフラ層）に委譲します。
// 削除対象のタグIDが問題に使用されている場合は ErrTagInUse を返します。
// タグが存在しない場合は ErrTagNotFound を返します。
func (uc *TagUseCase) DeleteTag(ctx context.Context, id string) error {
	if err := uc.tagRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("タグの削除に失敗しました: %w", err)
	}

	return nil
}
