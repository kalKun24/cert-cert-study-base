// Package usecase はビジネスロジック（ユースケース）を実装します。
// このパッケージは domain パッケージのみに依存します。
package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

// CommentUseCase はコメント管理に関するユースケースを実装します。
type CommentUseCase struct {
	commentRepo  domain.CommentRepository
	questionRepo domain.QuestionRepository
	userRepo     domain.UserRepository
	teamRepo     domain.TeamRepository
}

// NewCommentUseCase は CommentUseCase を生成します（コンストラクタインジェクション）。
func NewCommentUseCase(
	commentRepo domain.CommentRepository,
	questionRepo domain.QuestionRepository,
	userRepo domain.UserRepository,
	teamRepo domain.TeamRepository,
) *CommentUseCase {
	return &CommentUseCase{
		commentRepo:  commentRepo,
		questionRepo: questionRepo,
		userRepo:     userRepo,
		teamRepo:     teamRepo,
	}
}

// checkTeamMembershipForComment は呼び出し元がチームメンバーかどうかを確認します。
// admin の場合はメンバーシップチェックをスキップします。
func (uc *CommentUseCase) checkTeamMembershipForComment(ctx context.Context, callerID string, callerRole domain.Role, teamID string) error {
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

// checkQuestionAccess は指定した問題の閲覧権限をチームスコープで確認します。
// チームメンバーまたは admin のみアクセス可能です。
// メンバー非所属の場合は ErrPermissionDenied を返します。
// 問題が存在しない場合は ErrQuestionNotFound をラップして返します。
func (uc *CommentUseCase) checkQuestionAccess(ctx context.Context, questionID, teamID, callerID string, callerRole domain.Role) (*domain.Question, error) {
	if err := uc.checkTeamMembershipForComment(ctx, callerID, callerRole, teamID); err != nil {
		return nil, domain.ErrPermissionDenied
	}

	question, err := uc.questionRepo.FindByID(ctx, teamID, questionID)
	if err != nil {
		return nil, fmt.Errorf("問題の取得に失敗しました: %w", err)
	}

	// admin はすべての問題（draft 含む）を閲覧可能
	if callerRole != domain.RoleAdmin && !isVisibleToTeamMember(question, callerID) {
		return nil, domain.ErrPermissionDenied
	}

	return question, nil
}

// CommentWithDisplayName はコメントと投稿者の表示名をまとめた出力型です。
type CommentWithDisplayName struct {
	*domain.Comment
	DisplayName string
}

// CreateCommentInput はコメント投稿ユースケースの入力です。
type CreateCommentInput struct {
	// QuestionID は対象の問題ID
	QuestionID string
	// TeamID は対象のチームID
	TeamID string
	// Body はコメント本文（Markdown形式、必須）
	Body string
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
}

// CreateComment は指定した問題にコメントを投稿します。
// チームメンバーであり、問題の閲覧権限を持つユーザーのみ投稿可能です。
func (uc *CommentUseCase) CreateComment(ctx context.Context, input CreateCommentInput) (*CommentWithDisplayName, error) {
	if strings.TrimSpace(input.Body) == "" {
		return nil, domain.ErrCommentBodyEmpty
	}
	if len([]rune(input.Body)) > domain.MaxCommentBodyLength {
		return nil, domain.ErrCommentBodyTooLong
	}

	// チームメンバーかつ問題の閲覧権限チェック
	if _, err := uc.checkQuestionAccess(ctx, input.QuestionID, input.TeamID, input.CallerID, input.CallerRole); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	comment := &domain.Comment{
		ID:         uuid.NewString(),
		QuestionID: input.QuestionID,
		Body:       input.Body,
		CreatedBy:  input.CallerID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.commentRepo.Save(ctx, input.TeamID, comment); err != nil {
		return nil, fmt.Errorf("コメントの保存に失敗しました: %w", err)
	}

	displayName, err := uc.resolveDisplayName(ctx, input.CallerID)
	if err != nil {
		return nil, err
	}

	return &CommentWithDisplayName{Comment: comment, DisplayName: displayName}, nil
}

// ListCommentsInput はコメント一覧取得ユースケースの入力です。
type ListCommentsInput struct {
	// QuestionID は対象の問題ID
	QuestionID string
	// TeamID は対象のチームID
	TeamID string
	// CallerID はリクエストユーザーのID
	CallerID string
	// CallerRole はリクエストユーザーのロール
	CallerRole domain.Role
}

// ListComments は指定した問題のコメント一覧を投稿日時の昇順で返します。
// チームメンバーであり、問題の閲覧権限を持つユーザーのみ取得可能です。
func (uc *CommentUseCase) ListComments(ctx context.Context, input ListCommentsInput) ([]*CommentWithDisplayName, error) {
	// チームメンバーかつ問題の閲覧権限チェック
	if _, err := uc.checkQuestionAccess(ctx, input.QuestionID, input.TeamID, input.CallerID, input.CallerRole); err != nil {
		return nil, err
	}

	comments, err := uc.commentRepo.ListByQuestionID(ctx, input.TeamID, input.QuestionID)
	if err != nil {
		return nil, fmt.Errorf("コメント一覧の取得に失敗しました: %w", err)
	}

	// 投稿日時の昇順でソート
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	// ユーザーIDから表示名のマップを構築（一括取得してN+1クエリを回避）
	displayNames, err := uc.resolveDisplayNames(ctx, comments)
	if err != nil {
		return nil, err
	}

	result := make([]*CommentWithDisplayName, 0, len(comments))
	for _, c := range comments {
		result = append(result, &CommentWithDisplayName{
			Comment:     c,
			DisplayName: displayNames[c.CreatedBy],
		})
	}

	return result, nil
}

// UpdateCommentInput はコメント編集ユースケースの入力です。
type UpdateCommentInput struct {
	// QuestionID は対象の問題ID
	QuestionID string
	// TeamID は対象のチームID
	TeamID string
	// CommentID は編集するコメントのID
	CommentID string
	// Body は新しいコメント本文（Markdown形式、必須）
	Body string
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
}

// UpdateComment は指定したコメントを編集します。
// 投稿者本人のみ編集可能です。チームメンバーであり問題の閲覧権限も必要です。
func (uc *CommentUseCase) UpdateComment(ctx context.Context, input UpdateCommentInput) (*CommentWithDisplayName, error) {
	if strings.TrimSpace(input.Body) == "" {
		return nil, domain.ErrCommentBodyEmpty
	}
	if len([]rune(input.Body)) > domain.MaxCommentBodyLength {
		return nil, domain.ErrCommentBodyTooLong
	}

	// チームメンバーかつ問題の閲覧権限チェック
	if _, err := uc.checkQuestionAccess(ctx, input.QuestionID, input.TeamID, input.CallerID, input.CallerRole); err != nil {
		return nil, err
	}

	comment, err := uc.commentRepo.FindByID(ctx, input.TeamID, input.QuestionID, input.CommentID)
	if err != nil {
		return nil, fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}

	// 認可チェック: 投稿者本人または admin のみ編集可能
	if comment.CreatedBy != input.CallerID && input.CallerRole != domain.RoleAdmin {
		return nil, domain.ErrPermissionDenied
	}

	comment.Body = input.Body
	comment.UpdatedAt = time.Now().UTC()

	if err := uc.commentRepo.Save(ctx, input.TeamID, comment); err != nil {
		return nil, fmt.Errorf("コメントの保存に失敗しました: %w", err)
	}

	displayName, err := uc.resolveDisplayName(ctx, comment.CreatedBy)
	if err != nil {
		return nil, err
	}

	return &CommentWithDisplayName{Comment: comment, DisplayName: displayName}, nil
}

// DeleteCommentInput はコメント削除ユースケースの入力です。
type DeleteCommentInput struct {
	// QuestionID は対象の問題ID
	QuestionID string
	// TeamID は対象のチームID
	TeamID string
	// CommentID は削除するコメントのID
	CommentID string
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
}

// DeleteComment は指定したコメントを削除します。
// 投稿者本人または admin のみ削除可能です。チームメンバーであり問題の閲覧権限も必要です。
func (uc *CommentUseCase) DeleteComment(ctx context.Context, input DeleteCommentInput) error {
	// チームメンバーかつ問題の閲覧権限チェック
	if _, err := uc.checkQuestionAccess(ctx, input.QuestionID, input.TeamID, input.CallerID, input.CallerRole); err != nil {
		return err
	}

	comment, err := uc.commentRepo.FindByID(ctx, input.TeamID, input.QuestionID, input.CommentID)
	if err != nil {
		return fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}

	// 認可チェック: 投稿者本人または admin のみ削除可能
	if comment.CreatedBy != input.CallerID && input.CallerRole != domain.RoleAdmin {
		return domain.ErrPermissionDenied
	}

	if err := uc.commentRepo.Delete(ctx, input.TeamID, input.QuestionID, input.CommentID); err != nil {
		return fmt.Errorf("コメントの削除に失敗しました: %w", err)
	}

	return nil
}

// resolveDisplayName はユーザーIDから表示名を取得します。
// userRepo が nil の場合（テスト用）はユーザーIDをそのまま返します。
func (uc *CommentUseCase) resolveDisplayName(ctx context.Context, userID string) (string, error) {
	if uc.userRepo == nil {
		return userID, nil
	}
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		// ユーザーが見つからない場合もユーザーIDで代替（コメント一覧の取得を中断しない）
		if errors.Is(err, domain.ErrUserNotFound) {
			return userID, nil
		}
		return "", fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}
	return user.DisplayName, nil
}

// resolveDisplayNames はコメント一覧の投稿者の表示名をまとめて取得します。
// userID → displayName のマップを返します。
// userRepo が nil の場合（テスト用）はユーザーIDをそのまま使います。
func (uc *CommentUseCase) resolveDisplayNames(ctx context.Context, comments []*domain.Comment) (map[string]string, error) {
	displayNames := make(map[string]string, len(comments))

	// 重複ユーザーIDを排除してから取得
	seen := make(map[string]struct{})
	for _, c := range comments {
		seen[c.CreatedBy] = struct{}{}
	}

	for userID := range seen {
		name, err := uc.resolveDisplayName(ctx, userID)
		if err != nil {
			return nil, err
		}
		displayNames[userID] = name
	}

	return displayNames, nil
}
