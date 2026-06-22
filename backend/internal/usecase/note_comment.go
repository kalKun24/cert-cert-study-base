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

// NoteCommentUseCase はノートコメント管理に関するユースケースを実装します。
type NoteCommentUseCase struct {
	noteCommentRepo domain.NoteCommentRepository
	noteRepo        domain.NoteRepository
	teamRepo        domain.TeamRepository
	userRepo        domain.UserRepository
}

// NewNoteCommentUseCase は NoteCommentUseCase を生成します（コンストラクタインジェクション）。
func NewNoteCommentUseCase(
	noteCommentRepo domain.NoteCommentRepository,
	noteRepo domain.NoteRepository,
	teamRepo domain.TeamRepository,
	userRepo domain.UserRepository,
) *NoteCommentUseCase {
	return &NoteCommentUseCase{
		noteCommentRepo: noteCommentRepo,
		noteRepo:        noteRepo,
		teamRepo:        teamRepo,
		userRepo:        userRepo,
	}
}

// NoteCommentWithDisplayName はノートコメントと投稿者の表示名をまとめた出力型です。
type NoteCommentWithDisplayName struct {
	*domain.NoteComment
	DisplayName string
}

// resolveNoteCommentDisplayName はユーザーIDから表示名を取得します。
func (uc *NoteCommentUseCase) resolveNoteCommentDisplayName(ctx context.Context, userID string) (string, error) {
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

// resolveNoteCommentDisplayNames はコメント一覧の投稿者の表示名をまとめて取得します。
// userID → displayName のマップを返します。
func (uc *NoteCommentUseCase) resolveNoteCommentDisplayNames(ctx context.Context, comments []*domain.NoteComment) (map[string]string, error) {
	displayNames := make(map[string]string, len(comments))
	for _, c := range comments {
		if _, ok := displayNames[c.CreatedBy]; ok {
			continue
		}
		name, err := uc.resolveNoteCommentDisplayName(ctx, c.CreatedBy)
		if err != nil {
			return nil, err
		}
		displayNames[c.CreatedBy] = name
	}
	return displayNames, nil
}

// checkNoteAccess はチームメンバーシップとノートの閲覧権限をまとめて確認します。
// チームメンバーかつノートの閲覧権限を持つ場合にノートを返します。
// admin の場合はメンバーシップチェックをスキップし、すべてのノート（draft 含む）にアクセス可能です。
func (uc *NoteCommentUseCase) checkNoteAccess(ctx context.Context, teamID, noteID, callerID string, callerRole domain.Role) (*domain.Note, error) {
	// チームメンバーチェック（admin はスキップ）
	if callerRole != domain.RoleAdmin {
		isMember, err := uc.teamRepo.IsMember(ctx, teamID, callerID)
		if err != nil {
			return nil, fmt.Errorf("チームメンバー確認に失敗しました: %w", err)
		}
		if !isMember {
			return nil, domain.ErrPermissionDenied
		}
	}

	// ノート取得
	note, err := uc.noteRepo.FindByID(ctx, teamID, noteID)
	if err != nil {
		return nil, fmt.Errorf("ノートの取得に失敗しました: %w", err)
	}

	// 可視性チェック: draft は作成者本人または admin のみ
	if callerRole != domain.RoleAdmin && !isNoteVisibleToCommentor(note, callerID) {
		return nil, domain.ErrPermissionDenied
	}

	return note, nil
}

// isNoteVisibleToCommentor はノートがコメント投稿者に対して可視かどうかを返します。
// draft は作成者本人のみ、それ以外はチームメンバー全員に可視です。
func isNoteVisibleToCommentor(n *domain.Note, callerID string) bool {
	if n.Status == domain.NoteStatusDraft {
		return n.CreatedBy == callerID
	}
	return true
}

// CreateNoteCommentInput はノートコメント投稿ユースケースの入力です。
type CreateNoteCommentInput struct {
	// TeamID は対象のチームID
	TeamID string
	// NoteID は対象のノートID
	NoteID string
	// Body はコメント本文（Markdown形式、必須）
	Body string
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
}

// CreateNoteComment は指定したノートにコメントを投稿します。
// チームメンバーであり、ノートの閲覧権限を持つユーザーのみ投稿可能です。
func (uc *NoteCommentUseCase) CreateNoteComment(ctx context.Context, input CreateNoteCommentInput) (*NoteCommentWithDisplayName, error) {
	if strings.TrimSpace(input.Body) == "" {
		return nil, domain.ErrCommentBodyEmpty
	}
	if len([]rune(input.Body)) > domain.MaxCommentBodyLength {
		return nil, domain.ErrCommentBodyTooLong
	}

	// チームメンバーかつノートの閲覧権限チェック
	if _, err := uc.checkNoteAccess(ctx, input.TeamID, input.NoteID, input.CallerID, input.CallerRole); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	comment := &domain.NoteComment{
		ID:        uuid.NewString(),
		NoteID:    input.NoteID,
		Body:      input.Body,
		CreatedBy: input.CallerID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.noteCommentRepo.Save(ctx, input.TeamID, comment); err != nil {
		return nil, fmt.Errorf("コメントの保存に失敗しました: %w", err)
	}

	displayName, err := uc.resolveNoteCommentDisplayName(ctx, input.CallerID)
	if err != nil {
		return nil, err
	}

	return &NoteCommentWithDisplayName{NoteComment: comment, DisplayName: displayName}, nil
}

// ListNoteComments は指定したノートのコメント一覧を投稿日時の昇順で返します。
// チームメンバーであり、ノートの閲覧権限を持つユーザーのみ取得可能です。
func (uc *NoteCommentUseCase) ListNoteComments(ctx context.Context, teamID, noteID, callerID string, callerRole domain.Role) ([]*NoteCommentWithDisplayName, error) {
	// チームメンバーかつノートの閲覧権限チェック
	if _, err := uc.checkNoteAccess(ctx, teamID, noteID, callerID, callerRole); err != nil {
		return nil, err
	}

	comments, err := uc.noteCommentRepo.ListByNoteID(ctx, teamID, noteID)
	if err != nil {
		return nil, fmt.Errorf("コメント一覧の取得に失敗しました: %w", err)
	}

	// 投稿日時の昇順でソート
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	// ユーザーIDから表示名のマップを構築（一括取得してN+1クエリを回避）
	displayNames, err := uc.resolveNoteCommentDisplayNames(ctx, comments)
	if err != nil {
		return nil, err
	}

	result := make([]*NoteCommentWithDisplayName, 0, len(comments))
	for _, c := range comments {
		result = append(result, &NoteCommentWithDisplayName{
			NoteComment: c,
			DisplayName: displayNames[c.CreatedBy],
		})
	}

	return result, nil
}

// UpdateNoteCommentInput はノートコメント編集ユースケースの入力です。
type UpdateNoteCommentInput struct {
	// TeamID は対象のチームID
	TeamID string
	// NoteID は対象のノートID
	NoteID string
	// CommentID は編集するコメントのID
	CommentID string
	// Body は新しいコメント本文（Markdown形式、必須）
	Body string
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
}

// UpdateNoteComment は指定したノートコメントを編集します。
// 投稿者本人のみ編集可能です。チームメンバーでありノートの閲覧権限も必要です。
func (uc *NoteCommentUseCase) UpdateNoteComment(ctx context.Context, input UpdateNoteCommentInput) (*NoteCommentWithDisplayName, error) {
	if strings.TrimSpace(input.Body) == "" {
		return nil, domain.ErrCommentBodyEmpty
	}
	if len([]rune(input.Body)) > domain.MaxCommentBodyLength {
		return nil, domain.ErrCommentBodyTooLong
	}

	// チームメンバーかつノートの閲覧権限チェック
	if _, err := uc.checkNoteAccess(ctx, input.TeamID, input.NoteID, input.CallerID, input.CallerRole); err != nil {
		return nil, err
	}

	comment, err := uc.noteCommentRepo.FindByID(ctx, input.TeamID, input.NoteID, input.CommentID)
	if err != nil {
		return nil, fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}

	// 認可チェック: 投稿者本人のみ編集可能
	if comment.CreatedBy != input.CallerID {
		return nil, domain.ErrPermissionDenied
	}

	comment.Body = input.Body
	comment.UpdatedAt = time.Now().UTC()

	if err := uc.noteCommentRepo.Save(ctx, input.TeamID, comment); err != nil {
		return nil, fmt.Errorf("コメントの保存に失敗しました: %w", err)
	}

	displayName, err := uc.resolveNoteCommentDisplayName(ctx, comment.CreatedBy)
	if err != nil {
		return nil, err
	}

	return &NoteCommentWithDisplayName{NoteComment: comment, DisplayName: displayName}, nil
}

// DeleteNoteComment は指定したノートコメントを削除します。
// 投稿者本人または admin のみ削除可能です。チームメンバーでありノートの閲覧権限も必要です。
func (uc *NoteCommentUseCase) DeleteNoteComment(ctx context.Context, teamID, noteID, commentID, callerID string, callerRole domain.Role) error {
	// チームメンバーかつノートの閲覧権限チェック
	if _, err := uc.checkNoteAccess(ctx, teamID, noteID, callerID, callerRole); err != nil {
		return err
	}

	comment, err := uc.noteCommentRepo.FindByID(ctx, teamID, noteID, commentID)
	if err != nil {
		return fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}

	// 認可チェック: 投稿者本人または admin のみ削除可能
	if comment.CreatedBy != callerID && callerRole != domain.RoleAdmin {
		return domain.ErrPermissionDenied
	}

	if err := uc.noteCommentRepo.Delete(ctx, teamID, noteID, commentID); err != nil {
		return fmt.Errorf("コメントの削除に失敗しました: %w", err)
	}

	return nil
}
