// Package usecase はビジネスロジック（ユースケース）を実装します。
// このパッケージは domain パッケージのみに依存します。
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

// NoteUseCase はノート管理に関するユースケースを実装します。
type NoteUseCase struct {
	noteRepo domain.NoteRepository
	teamRepo domain.TeamRepository
}

// NewNoteUseCase は NoteUseCase を生成します（コンストラクタインジェクション）。
func NewNoteUseCase(noteRepo domain.NoteRepository, teamRepo domain.TeamRepository) *NoteUseCase {
	return &NoteUseCase{
		noteRepo: noteRepo,
		teamRepo: teamRepo,
	}
}

// checkNoteTeamMembership は呼び出し元がチームメンバーかどうかを確認します。
// admin の場合はメンバーシップチェックをスキップします。
func (uc *NoteUseCase) checkNoteTeamMembership(ctx context.Context, callerID string, callerRole domain.Role, teamID string) error {
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

// isNoteEditor は callerID がノートを編集・削除できるかどうかを返します。
// チームオーナー、admin、または作成者本人が編集可能です。
func (uc *NoteUseCase) isNoteEditor(ctx context.Context, note *domain.Note, callerID string, callerRole domain.Role) (bool, error) {
	if callerRole == domain.RoleAdmin {
		return true, nil
	}
	if note.CreatedBy == callerID {
		return true, nil
	}
	// チームオーナーかどうかを確認
	owners, err := uc.teamRepo.FindOwners(ctx, note.TeamID)
	if err != nil {
		return false, fmt.Errorf("チームオーナー確認に失敗しました: %w", err)
	}
	for _, o := range owners {
		if o.UserID == callerID {
			return true, nil
		}
	}
	return false, nil
}

// isNoteVisibleToTeamMember はチームメンバーに対してノートが可視かどうかを返します。
// draft は作成者本人のみ、published/private はチームメンバー全員。
func isNoteVisibleToTeamMember(n *domain.Note, callerID string) bool {
	switch n.Status {
	case domain.NoteStatusDraft:
		return n.CreatedBy == callerID
	case domain.NoteStatusPublished, domain.NoteStatusPrivate:
		return true
	default:
		// status 未設定（後方互換）: draft 扱いで作成者のみ
		return n.CreatedBy == callerID
	}
}

// CreateNoteInput はノート作成ユースケースの入力です。
type CreateNoteInput struct {
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
	// Title はノートタイトル（必須）
	Title string
	// Body は本文（Markdown形式）
	Body string
	// DiscussionPoints は議論点（Markdown形式）
	DiscussionPoints string
	// Memo は自由記述メモ（Markdown形式）
	Memo string
	// Tags はタグ（フラット・複数付与可）
	Tags []string
	// Status は公開ステータス（省略時は draft）
	Status domain.NoteStatus
}

// CreateNote は新しいノートを作成します。
// チームメンバーまたは admin が作成可能です。
func (uc *NoteUseCase) CreateNote(ctx context.Context, teamID string, input CreateNoteInput) (*domain.Note, error) {
	if err := uc.checkNoteTeamMembership(ctx, input.CallerID, input.CallerRole, teamID); err != nil {
		return nil, err
	}

	if input.Title == "" {
		return nil, fmt.Errorf("タイトルは必須です")
	}

	// ステータスのデフォルト値を設定
	status := input.Status
	if status == "" {
		status = domain.NoteStatusDraft
	} else if !status.IsValid() {
		return nil, domain.ErrInvalidNoteStatus
	}

	// Tags のゼロ値を空スライスにする
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}

	now := time.Now().UTC()
	note := &domain.Note{
		ID:               uuid.NewString(),
		TeamID:           teamID,
		Title:            input.Title,
		Body:             input.Body,
		DiscussionPoints: input.DiscussionPoints,
		Memo:             input.Memo,
		Tags:             tags,
		Status:           status,
		CreatedBy:        input.CallerID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := uc.noteRepo.Save(ctx, note); err != nil {
		return nil, fmt.Errorf("ノートの保存に失敗しました: %w", err)
	}

	return note, nil
}

// SearchNotesInput はノート検索・フィルタリングユースケースの入力です。
type SearchNotesInput struct {
	// CallerID はリクエストユーザーのID
	CallerID string
	// CallerRole はリクエストユーザーのロール
	CallerRole domain.Role
	// TagIDs はAND絞り込みするタグIDの一覧。空の場合はタグフィルタなし。
	TagIDs []string
	// Keyword はキーワード検索文字列。空の場合は検索なし。
	Keyword string
	// Page はページ番号（1始まり）。0以下の場合は1とみなします。
	Page int
	// PerPage は1ページあたりの件数。0以下の場合は20とみなします。最大100。
	PerPage int
}

// SearchNotesResult はノート検索・フィルタリングユースケースの結果です。
type SearchNotesResult struct {
	// Items は現在ページのノート一覧
	Items []*domain.Note
	// Total はフィルタリング後の総件数
	Total int
	// Page は現在のページ番号（1始まり）
	Page int
	// PerPage は1ページあたりの件数
	PerPage int
	// TotalPages は総ページ数
	TotalPages int
}

// SearchNotes は検索・フィルタリング条件に基づき、可視性フィルタを適用したノート一覧をページネーション付きで返します。
// - チームメンバーまたは admin のみアクセス可能です。
// - タグIDは複数指定した場合AND絞り込みを行います。
// - キーワードは title / body / discussion_points / memo を対象に部分一致検索します。
// - 可視性ルール: draft は作成者本人または admin のみ、published/private はメンバー全員。
// - 検索結果0件は空のItemsと200を返します（エラーにしません）。
func (uc *NoteUseCase) SearchNotes(ctx context.Context, teamID string, input SearchNotesInput) (*SearchNotesResult, error) {
	if err := uc.checkNoteTeamMembership(ctx, input.CallerID, input.CallerRole, teamID); err != nil {
		return nil, err
	}

	// ページネーションパラメータの正規化
	page := input.Page
	if page < 1 {
		page = 1
	}
	perPage := input.PerPage
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	// リポジトリに検索フィルタを渡して候補を取得
	filter := domain.NoteSearchFilter{
		TagIDs:  input.TagIDs,
		Keyword: input.Keyword,
	}
	candidates, err := uc.noteRepo.SearchByTeam(ctx, teamID, filter)
	if err != nil {
		return nil, fmt.Errorf("ノート検索に失敗しました: %w", err)
	}

	// 可視性フィルタリング: draft は作成者本人または admin のみ、published/private はメンバー全員
	visible := make([]*domain.Note, 0, len(candidates))
	for _, n := range candidates {
		if input.CallerRole == domain.RoleAdmin || isNoteVisibleToTeamMember(n, input.CallerID) {
			visible = append(visible, n)
		}
	}

	// ページネーション計算
	total := len(visible)
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	// ページ範囲のクリッピング
	if page > totalPages {
		page = totalPages
	}

	// オフセット計算とスライス
	start := (page - 1) * perPage
	end := start + perPage
	if end > total {
		end = total
	}
	items := visible[start:end]

	return &SearchNotesResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// ListNotesInput はノート一覧取得ユースケースの入力です。
type ListNotesInput struct {
	// CallerID はリクエストユーザーのID
	CallerID string
	// CallerRole はリクエストユーザーのロール
	CallerRole domain.Role
}

// ListNotes は可視性ルールに基づいてフィルタリングしたノート一覧を返します。
// - チームメンバーまたは admin のみアクセス可能です。
// - status=draft → 作成者本人または admin のみ返す。
// - status=published / private → チームメンバー全員に返す。
func (uc *NoteUseCase) ListNotes(ctx context.Context, teamID string, input ListNotesInput) ([]*domain.Note, error) {
	if err := uc.checkNoteTeamMembership(ctx, input.CallerID, input.CallerRole, teamID); err != nil {
		return nil, err
	}

	notes, err := uc.noteRepo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("ノート一覧の取得に失敗しました: %w", err)
	}

	visible := make([]*domain.Note, 0, len(notes))
	for _, n := range notes {
		if input.CallerRole == domain.RoleAdmin || isNoteVisibleToTeamMember(n, input.CallerID) {
			visible = append(visible, n)
		}
	}
	return visible, nil
}

// GetNoteInput はノート詳細取得ユースケースの入力です。
type GetNoteInput struct {
	// CallerID はリクエストユーザーのID
	CallerID string
	// CallerRole はリクエストユーザーのロール
	CallerRole domain.Role
}

// GetNote はIDでノートを取得します。
// チームメンバーチェックを行い、ノートがそのチームに属するかを確認します。
// チーム不一致の場合や可視性ルールによりアクセス不可の場合は ErrNoteNotFound を返します。
func (uc *NoteUseCase) GetNote(ctx context.Context, noteID string, teamID string, input GetNoteInput) (*domain.Note, error) {
	if err := uc.checkNoteTeamMembership(ctx, input.CallerID, input.CallerRole, teamID); err != nil {
		return nil, err
	}

	note, err := uc.noteRepo.FindByID(ctx, teamID, noteID)
	if err != nil {
		return nil, fmt.Errorf("ノートの取得に失敗しました: %w", err)
	}

	// admin はすべてのノート（draft 含む）を閲覧可能
	if input.CallerRole != domain.RoleAdmin && !isNoteVisibleToTeamMember(note, input.CallerID) {
		return nil, fmt.Errorf("ノートの取得に失敗しました: %w", domain.ErrNoteNotFound)
	}

	return note, nil
}

// UpdateNoteVisibilityInput は公開設定変更ユースケースの入力です。
type UpdateNoteVisibilityInput struct {
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
	// Status は変更後の公開ステータス
	Status domain.NoteStatus
}

// UpdateNoteVisibility は指定IDのノートの公開設定を変更します。
// チームメンバーチェックを行い、チームオーナー・admin・作成者本人のみ変更可能です。
func (uc *NoteUseCase) UpdateNoteVisibility(ctx context.Context, noteID string, teamID string, input UpdateNoteVisibilityInput) (*domain.Note, error) {
	if err := uc.checkNoteTeamMembership(ctx, input.CallerID, input.CallerRole, teamID); err != nil {
		return nil, err
	}

	note, err := uc.noteRepo.FindByID(ctx, teamID, noteID)
	if err != nil {
		return nil, fmt.Errorf("ノートの取得に失敗しました: %w", err)
	}

	// 認可チェック: チームオーナー・admin・作成者本人のみ変更可能
	canEdit, err := uc.isNoteEditor(ctx, note, input.CallerID, input.CallerRole)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, domain.ErrPermissionDenied
	}

	// status の検証と設定
	if !input.Status.IsValid() {
		return nil, domain.ErrInvalidNoteStatus
	}
	note.Status = input.Status

	note.UpdatedAt = time.Now().UTC()

	if err := uc.noteRepo.Save(ctx, note); err != nil {
		return nil, fmt.Errorf("ノートの保存に失敗しました: %w", err)
	}

	return note, nil
}

// UpdateNoteInput はノート更新ユースケースの入力です。
// 各フィールドはポインタ型にしてゼロ値との区別を可能にします。
type UpdateNoteInput struct {
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
	// Title はノートタイトル（nil の場合は変更しない）
	Title *string
	// Body は本文（Markdown形式、nil の場合は変更しない）
	Body *string
	// DiscussionPoints は議論点（Markdown形式、nil の場合は変更しない）
	DiscussionPoints *string
	// Memo は自由記述メモ（Markdown形式、nil の場合は変更しない）
	Memo *string
	// Tags はタグ（nil の場合は変更しない）
	Tags []string
	// TagsSet はタグを明示的に nil（変更なし）か非 nil（更新）か区別するためのフラグ
	TagsSet bool
	// Status は公開ステータス（nil の場合は変更しない）
	Status *domain.NoteStatus
}

// UpdateNote は指定IDのノートを更新します。
// チームメンバーチェックを行い、チームオーナー・admin・作成者本人のみ更新可能です。
func (uc *NoteUseCase) UpdateNote(ctx context.Context, noteID string, teamID string, input UpdateNoteInput) (*domain.Note, error) {
	if err := uc.checkNoteTeamMembership(ctx, input.CallerID, input.CallerRole, teamID); err != nil {
		return nil, err
	}

	note, err := uc.noteRepo.FindByID(ctx, teamID, noteID)
	if err != nil {
		return nil, fmt.Errorf("ノートの取得に失敗しました: %w", err)
	}

	// 認可チェック: チームオーナー・admin・作成者本人のみ更新可能
	canEdit, err := uc.isNoteEditor(ctx, note, input.CallerID, input.CallerRole)
	if err != nil {
		return nil, err
	}
	if !canEdit {
		return nil, domain.ErrPermissionDenied
	}

	if input.Title != nil {
		if *input.Title == "" {
			return nil, fmt.Errorf("タイトルは必須です")
		}
		note.Title = *input.Title
	}
	if input.Body != nil {
		note.Body = *input.Body
	}
	if input.DiscussionPoints != nil {
		note.DiscussionPoints = *input.DiscussionPoints
	}
	if input.Memo != nil {
		note.Memo = *input.Memo
	}
	if input.TagsSet {
		if input.Tags == nil {
			note.Tags = []string{}
		} else {
			note.Tags = input.Tags
		}
	}
	if input.Status != nil {
		if !input.Status.IsValid() {
			return nil, domain.ErrInvalidNoteStatus
		}
		note.Status = *input.Status
	}

	note.UpdatedAt = time.Now().UTC()

	if err := uc.noteRepo.Save(ctx, note); err != nil {
		return nil, fmt.Errorf("ノートの保存に失敗しました: %w", err)
	}

	return note, nil
}

// DeleteNote は指定IDのノートを削除します。
// チームメンバーチェックを行い、チームオーナー・admin・作成者本人のみ削除可能です。
func (uc *NoteUseCase) DeleteNote(ctx context.Context, noteID string, teamID string, callerID string, callerRole domain.Role) error {
	if err := uc.checkNoteTeamMembership(ctx, callerID, callerRole, teamID); err != nil {
		return err
	}

	note, err := uc.noteRepo.FindByID(ctx, teamID, noteID)
	if err != nil {
		return fmt.Errorf("ノートの取得に失敗しました: %w", err)
	}

	// 認可チェック: チームオーナー・admin・作成者本人のみ削除可能
	canEdit, err := uc.isNoteEditor(ctx, note, callerID, callerRole)
	if err != nil {
		return err
	}
	if !canEdit {
		return domain.ErrPermissionDenied
	}

	if err := uc.noteRepo.Delete(ctx, teamID, noteID); err != nil {
		return fmt.Errorf("ノートの削除に失敗しました: %w", err)
	}

	return nil
}
