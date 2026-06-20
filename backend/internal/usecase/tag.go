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
	tagRepo  domain.TagRepository
	teamRepo domain.TeamRepository
}

// NewTagUseCase は TagUseCase を生成します（コンストラクタインジェクション）。
func NewTagUseCase(tagRepo domain.TagRepository, teamRepo domain.TeamRepository) *TagUseCase {
	return &TagUseCase{
		tagRepo:  tagRepo,
		teamRepo: teamRepo,
	}
}

// checkTeamAccess は呼び出し元がチームメンバーかadminかを確認します。
// admin の場合はメンバーシップチェックをスキップします。
func (uc *TagUseCase) checkTeamAccess(ctx context.Context, callerID string, callerRole domain.Role, teamID string) error {
	if callerRole == domain.RoleAdmin {
		return nil
	}
	isMember, err := uc.teamRepo.IsMember(ctx, teamID, callerID)
	if err != nil {
		return fmt.Errorf("チームメンバー確認に失敗しました: %w", err)
	}
	if !isMember {
		return domain.ErrMemberNotFound
	}
	return nil
}

// ListTags は指定チームのタグ一覧を取得します。
// チームメンバーまたは admin のみ取得可能です。
func (uc *TagUseCase) ListTags(ctx context.Context, callerID string, callerRole domain.Role, teamID string) ([]*domain.Tag, error) {
	if err := uc.checkTeamAccess(ctx, callerID, callerRole, teamID); err != nil {
		return nil, err
	}
	tags, err := uc.tagRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("タグ一覧の取得に失敗しました: %w", err)
	}
	return tags, nil
}

// CreateTagInput はタグ作成ユースケースの入力です。
type CreateTagInput struct {
	// Name はタグ名（必須・チーム内で一意）
	Name string
}

// CreateTag は指定チームに新しいタグを作成します。
// チームメンバーまたは admin が実行可能です。
// タグ名のチーム内重複は ErrTagNameAlreadyExists を返します。
func (uc *TagUseCase) CreateTag(ctx context.Context, callerID string, callerRole domain.Role, teamID string, input CreateTagInput) (*domain.Tag, error) {
	if err := uc.checkTeamAccess(ctx, callerID, callerRole, teamID); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("タグ名は必須です")
	}
	if len([]rune(name)) > tagNameMaxLength {
		return nil, fmt.Errorf("タグ名は%d文字以内で入力してください", tagNameMaxLength)
	}

	// チーム内でのタグ名重複チェック
	if _, err := uc.tagRepo.FindByName(ctx, teamID, name); err == nil {
		return nil, domain.ErrTagNameAlreadyExists
	}

	now := time.Now().UTC()
	tag := &domain.Tag{
		ID:        uuid.NewString(),
		TeamID:    teamID,
		Name:      name,
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

// UpdateTag は指定チームの指定IDのタグを更新します。
// admin のみ実行可能です。
// タグ名のチーム内重複は ErrTagNameAlreadyExists を返します。
func (uc *TagUseCase) UpdateTag(ctx context.Context, callerRole domain.Role, teamID string, id string, input UpdateTagInput) (*domain.Tag, error) {
	if callerRole != domain.RoleAdmin {
		return nil, domain.ErrMemberNotFound
	}

	tag, err := uc.tagRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("タグの取得に失敗しました: %w", err)
	}

	// タグが指定チームに属することを確認
	if tag.TeamID != teamID {
		return nil, domain.ErrTagTeamMismatch
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, fmt.Errorf("タグ名は必須です")
		}
		if len([]rune(name)) > tagNameMaxLength {
			return nil, fmt.Errorf("タグ名は%d文字以内で入力してください", tagNameMaxLength)
		}
		// チーム内でのタグ名重複チェック（自分自身は除外）
		existing, err := uc.tagRepo.FindByName(ctx, teamID, name)
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

// DeleteTag は指定チームの指定IDのタグを削除します。
// チームメンバーまたは admin が実行可能です。
// 使用中チェックは TagRepository の Delete 実装（インフラ層）に委譲します。
func (uc *TagUseCase) DeleteTag(ctx context.Context, callerID string, callerRole domain.Role, teamID string, id string) error {
	if err := uc.checkTeamAccess(ctx, callerID, callerRole, teamID); err != nil {
		return err
	}

	// タグが指定チームに属することを確認
	tag, err := uc.tagRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("タグの取得に失敗しました: %w", err)
	}
	if tag.TeamID != teamID {
		return domain.ErrTagTeamMismatch
	}

	if err := uc.tagRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("タグの削除に失敗しました: %w", err)
	}

	return nil
}
