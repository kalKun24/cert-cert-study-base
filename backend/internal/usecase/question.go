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

// QuestionUseCase は問題管理に関するユースケースを実装します。
type QuestionUseCase struct {
	questionRepo domain.QuestionRepository
	teamRepo     domain.TeamRepository
}

// NewQuestionUseCase は QuestionUseCase を生成します（コンストラクタインジェクション）。
func NewQuestionUseCase(questionRepo domain.QuestionRepository, teamRepo domain.TeamRepository) *QuestionUseCase {
	return &QuestionUseCase{
		questionRepo: questionRepo,
		teamRepo:     teamRepo,
	}
}

// callerTeamIDs はリクエストユーザーが所属するチームIDの集合を返します。
// チームリポジトリが nil の場合（テスト用）は空の集合を返します。
func (uc *QuestionUseCase) callerTeamIDs(ctx context.Context, callerID string) (map[string]struct{}, error) {
	if uc.teamRepo == nil {
		return map[string]struct{}{}, nil
	}
	teams, err := uc.teamRepo.ListByOwnerOrMember(ctx, callerID)
	if err != nil {
		return nil, fmt.Errorf("所属チームの取得に失敗しました: %w", err)
	}
	ids := make(map[string]struct{}, len(teams))
	for _, t := range teams {
		ids[t.ID] = struct{}{}
	}
	return ids, nil
}

// CreateQuestionInput は問題作成ユースケースの入力です。
type CreateQuestionInput struct {
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// Title は問題タイトル（必須）
	Title string
	// Body は問題文（Markdown形式、必須）
	Body string
	// Answer は解答（Markdown形式）
	Answer string
	// Explanation は解説（Markdown形式）
	Explanation string
	// Memo は議論点メモ（Markdown形式）
	Memo string
	// Tags はタグ（フラット・複数付与可）
	Tags []string
	// Status は公開ステータス（省略時は draft）
	Status domain.QuestionStatus
	// VisibilityScope は公開範囲（省略時は all）
	VisibilityScope domain.VisibilityScope
	// PublishedTeamIDs は公開対象チームIDの一覧
	PublishedTeamIDs []string
}

// CreateQuestion は新しい問題を作成します。
// 認証済みユーザー（user以上）であれば誰でも作成可能です。
func (uc *QuestionUseCase) CreateQuestion(ctx context.Context, input CreateQuestionInput) (*domain.Question, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("タイトルは必須です")
	}

	// ステータスのデフォルト値を設定
	status := input.Status
	if status == "" {
		status = domain.QuestionStatusDraft
	} else if !status.IsValid() {
		return nil, domain.ErrInvalidQuestionStatus
	}

	// 公開範囲のデフォルト値を設定
	visibilityScope := input.VisibilityScope
	if visibilityScope == "" {
		visibilityScope = domain.VisibilityScopeAll
	} else if !visibilityScope.IsValid() {
		return nil, domain.ErrInvalidVisibilityScope
	}

	// Tags と PublishedTeamIDs のゼロ値を空スライスにする
	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}
	publishedTeamIDs := input.PublishedTeamIDs
	if publishedTeamIDs == nil {
		publishedTeamIDs = []string{}
	}

	now := time.Now().UTC()
	question := &domain.Question{
		ID:               uuid.NewString(),
		Title:            input.Title,
		Body:             input.Body,
		Answer:           input.Answer,
		Explanation:      input.Explanation,
		Memo:             input.Memo,
		Tags:             tags,
		Status:           status,
		VisibilityScope:  visibilityScope,
		PublishedTeamIDs: publishedTeamIDs,
		CreatedBy:        input.CallerID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := uc.questionRepo.Save(ctx, question); err != nil {
		return nil, fmt.Errorf("問題の保存に失敗しました: %w", err)
	}

	return question, nil
}

// ListQuestionsInput は問題一覧取得ユースケースの入力です。
type ListQuestionsInput struct {
	// CallerID はリクエストユーザーのID
	CallerID string
	// CallerRole はリクエストユーザーのロール
	CallerRole domain.Role
}

// SearchQuestionsInput は問題検索・フィルタリングユースケースの入力です。
type SearchQuestionsInput struct {
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

// SearchQuestionsResult は問題検索・フィルタリングユースケースの結果です。
type SearchQuestionsResult struct {
	// Items は現在ページの問題一覧
	Items []*domain.Question
	// Total はフィルタリング後の総件数
	Total int
	// Page は現在のページ番号（1始まり）
	Page int
	// PerPage は1ページあたりの件数
	PerPage int
	// TotalPages は総ページ数
	TotalPages int
}

// defaultPerPage はページネーションのデフォルト件数です。
const defaultPerPage = 20

// maxPerPage はページネーションの最大件数です。
const maxPerPage = 100

// SearchQuestions は検索・フィルタリング条件に基づき、可視性フィルタを適用した問題一覧をページネーション付きで返します。
// - タグIDは複数指定した場合AND絞り込みを行います。
// - キーワードはtitle / body / explanation / memo を対象に部分一致検索します。
// - 可視性ルール（status / visibility_scope）はListQuestionsと同一ルールを適用します。
// - 検索結果0件は空のItemsと200を返します（エラーにしません）。
func (uc *QuestionUseCase) SearchQuestions(ctx context.Context, input SearchQuestionsInput) (*SearchQuestionsResult, error) {
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
	filter := domain.QuestionSearchFilter{
		TagIDs:  input.TagIDs,
		Keyword: input.Keyword,
	}
	candidates, err := uc.questionRepo.Search(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("問題検索に失敗しました: %w", err)
	}

	// 可視性フィルタリング
	isAdmin := input.CallerRole == domain.RoleAdmin

	var callerTeamIDs map[string]struct{}
	if !isAdmin {
		callerTeamIDs, err = uc.callerTeamIDs(ctx, input.CallerID)
		if err != nil {
			return nil, fmt.Errorf("チーム情報の取得に失敗しました: %w", err)
		}
	}

	visible := make([]*domain.Question, 0, len(candidates))
	for _, q := range candidates {
		if q.IsVisibleTo(input.CallerID, isAdmin, callerTeamIDs) {
			visible = append(visible, q)
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

	return &SearchQuestionsResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// ListQuestions は可視性ルールに基づいてフィルタリングした問題一覧を返します。
// - status=published かつ visibility_scope=all → 全ログインユーザーに返す
// - status=published かつ visibility_scope=team → リクエストユーザーが published_team_ids のいずれかに所属する場合のみ返す
// - status=draft / private → 作成者本人のみ返す（admin は全件取得可）
func (uc *QuestionUseCase) ListQuestions(ctx context.Context, input ListQuestionsInput) ([]*domain.Question, error) {
	questions, err := uc.questionRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("問題一覧の取得に失敗しました: %w", err)
	}

	isAdmin := input.CallerRole == domain.RoleAdmin

	var callerTeamIDs map[string]struct{}
	if !isAdmin {
		var err error
		callerTeamIDs, err = uc.callerTeamIDs(ctx, input.CallerID)
		if err != nil {
			return nil, fmt.Errorf("チーム情報の取得に失敗しました: %w", err)
		}
	}

	visible := make([]*domain.Question, 0, len(questions))
	for _, q := range questions {
		if q.IsVisibleTo(input.CallerID, isAdmin, callerTeamIDs) {
			visible = append(visible, q)
		}
	}
	return visible, nil
}

// GetQuestionInput は問題詳細取得ユースケースの入力です。
type GetQuestionInput struct {
	// CallerID はリクエストユーザーのID
	CallerID string
	// CallerRole はリクエストユーザーのロール
	CallerRole domain.Role
}

// GetQuestion はIDで問題を取得します。
// 可視性ルールに基づき、閲覧不可の場合は ErrQuestionNotFound を返します。
func (uc *QuestionUseCase) GetQuestion(ctx context.Context, id string, input GetQuestionInput) (*domain.Question, error) {
	question, err := uc.questionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("問題の取得に失敗しました: %w", err)
	}

	isAdmin := input.CallerRole == domain.RoleAdmin

	var callerTeamIDs map[string]struct{}
	if !isAdmin {
		var err error
		callerTeamIDs, err = uc.callerTeamIDs(ctx, input.CallerID)
		if err != nil {
			return nil, fmt.Errorf("チーム情報の取得に失敗しました: %w", err)
		}
	}

	if !question.IsVisibleTo(input.CallerID, isAdmin, callerTeamIDs) {
		// 存在するが閲覧権限がない場合も404を返す（情報漏洩防止）
		return nil, fmt.Errorf("問題の取得に失敗しました: %w", domain.ErrQuestionNotFound)
	}

	return question, nil
}

// UpdateQuestionVisibilityInput は公開設定変更ユースケースの入力です。
type UpdateQuestionVisibilityInput struct {
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
	// Status は変更後の公開ステータス
	Status domain.QuestionStatus
	// VisibilityScope は変更後の公開範囲（省略時は変更しない）
	VisibilityScope *domain.VisibilityScope
	// PublishedTeamIDs は変更後の公開対象チームIDの一覧（省略時は変更しない）
	PublishedTeamIDs    []string
	PublishedTeamIDsSet bool
}

// maxPublishedTeamIDs は公開対象チームIDの最大件数です。
const maxPublishedTeamIDs = 50

// UpdateQuestionVisibility は指定IDの問題の公開設定を変更します。
// 作成者本人（created_by == callerID）または admin のみ変更可能です。
func (uc *QuestionUseCase) UpdateQuestionVisibility(ctx context.Context, id string, input UpdateQuestionVisibilityInput) (*domain.Question, error) {
	question, err := uc.questionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("問題の取得に失敗しました: %w", err)
	}

	// 認可チェック: 作成者本人または admin のみ変更可能
	if question.CreatedBy != input.CallerID && input.CallerRole != domain.RoleAdmin {
		return nil, domain.ErrPermissionDenied
	}

	// status の検証と設定
	if !input.Status.IsValid() {
		return nil, domain.ErrInvalidQuestionStatus
	}
	question.Status = input.Status

	// visibility_scope の検証と設定
	if input.VisibilityScope != nil {
		if !input.VisibilityScope.IsValid() {
			return nil, domain.ErrInvalidVisibilityScope
		}
		question.VisibilityScope = *input.VisibilityScope
	}

	// published_team_ids の設定
	if input.PublishedTeamIDsSet {
		if len(input.PublishedTeamIDs) > maxPublishedTeamIDs {
			return nil, fmt.Errorf("公開対象チームIDは最大%d件です", maxPublishedTeamIDs)
		}
		if input.PublishedTeamIDs == nil {
			question.PublishedTeamIDs = []string{}
		} else {
			question.PublishedTeamIDs = input.PublishedTeamIDs
		}
	}

	question.UpdatedAt = time.Now().UTC()

	if err := uc.questionRepo.Save(ctx, question); err != nil {
		return nil, fmt.Errorf("問題の保存に失敗しました: %w", err)
	}

	return question, nil
}

// UpdateQuestionInput は問題更新ユースケースの入力です。
// 各フィールドはポインタ型にしてゼロ値との区別を可能にします。
type UpdateQuestionInput struct {
	// CallerID は操作を実行するユーザーのID
	CallerID string
	// CallerRole は操作を実行するユーザーのロール
	CallerRole domain.Role
	// Title は問題タイトル（nil の場合は変更しない）
	Title *string
	// Body は問題文（Markdown形式、nil の場合は変更しない）
	Body *string
	// Answer は解答（Markdown形式、nil の場合は変更しない）
	Answer *string
	// Explanation は解説（Markdown形式、nil の場合は変更しない）
	Explanation *string
	// Memo は議論点メモ（Markdown形式、nil の場合は変更しない）
	Memo *string
	// Tags はタグ（nil の場合は変更しない）
	Tags []string
	// TagsSet はタグを明示的に nil（変更なし）か非 nil（更新）か区別するためのフラグ
	TagsSet bool
	// Status は公開ステータス（nil の場合は変更しない）
	Status *domain.QuestionStatus
	// VisibilityScope は公開範囲（nil の場合は変更しない）
	VisibilityScope *domain.VisibilityScope
	// PublishedTeamIDs は公開対象チームIDの一覧（nil の場合は変更しない）
	PublishedTeamIDs    []string
	PublishedTeamIDsSet bool
}

// UpdateQuestion は指定IDの問題を更新します。
// 作成者本人（created_by == callerID）または admin のみ更新可能です。
func (uc *QuestionUseCase) UpdateQuestion(ctx context.Context, id string, input UpdateQuestionInput) (*domain.Question, error) {
	question, err := uc.questionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("問題の取得に失敗しました: %w", err)
	}

	// 認可チェック: 作成者本人または admin のみ更新可能
	if question.CreatedBy != input.CallerID && input.CallerRole != domain.RoleAdmin {
		return nil, domain.ErrPermissionDenied
	}

	if input.Title != nil {
		if *input.Title == "" {
			return nil, fmt.Errorf("タイトルは必須です")
		}
		question.Title = *input.Title
	}
	if input.Body != nil {
		question.Body = *input.Body
	}
	if input.Answer != nil {
		question.Answer = *input.Answer
	}
	if input.Explanation != nil {
		question.Explanation = *input.Explanation
	}
	if input.Memo != nil {
		question.Memo = *input.Memo
	}
	if input.TagsSet {
		if input.Tags == nil {
			question.Tags = []string{}
		} else {
			question.Tags = input.Tags
		}
	}
	if input.Status != nil {
		if !input.Status.IsValid() {
			return nil, domain.ErrInvalidQuestionStatus
		}
		question.Status = *input.Status
	}
	if input.VisibilityScope != nil {
		if !input.VisibilityScope.IsValid() {
			return nil, domain.ErrInvalidVisibilityScope
		}
		question.VisibilityScope = *input.VisibilityScope
	}
	if input.PublishedTeamIDsSet {
		if input.PublishedTeamIDs == nil {
			question.PublishedTeamIDs = []string{}
		} else {
			question.PublishedTeamIDs = input.PublishedTeamIDs
		}
	}

	question.UpdatedAt = time.Now().UTC()

	if err := uc.questionRepo.Save(ctx, question); err != nil {
		return nil, fmt.Errorf("問題の保存に失敗しました: %w", err)
	}

	return question, nil
}

// DeleteQuestion は指定IDの問題を削除します。
// 作成者本人（created_by == callerID）または admin のみ削除可能です。
func (uc *QuestionUseCase) DeleteQuestion(ctx context.Context, id string, callerID string, callerRole domain.Role) error {
	question, err := uc.questionRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("問題の取得に失敗しました: %w", err)
	}

	// 認可チェック: 作成者本人または admin のみ削除可能
	if question.CreatedBy != callerID && callerRole != domain.RoleAdmin {
		return domain.ErrPermissionDenied
	}

	if err := uc.questionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("問題の削除に失敗しました: %w", err)
	}

	return nil
}
