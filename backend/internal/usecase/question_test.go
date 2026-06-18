package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// --- モック定義 ---

// mockQuestionRepository は domain.QuestionRepository のモックです。
type mockQuestionRepository struct {
	questions map[string]*domain.Question
	saveErr   error
	deleteErr error
}

func newMockQuestionRepository() *mockQuestionRepository {
	return &mockQuestionRepository{
		questions: make(map[string]*domain.Question),
	}
}

// newQuestionUseCase はテスト用の QuestionUseCase を生成します（teamRepo は team_test.go のモックを使用）。
func newQuestionUseCase(qRepo *mockQuestionRepository) *usecase.QuestionUseCase {
	return usecase.NewQuestionUseCase(qRepo, newMockTeamRepository())
}

// addQuestion はテスト用の問題をモックに追加します。
func (m *mockQuestionRepository) addQuestion(q *domain.Question) {
	m.questions[q.ID] = q
}

func (m *mockQuestionRepository) FindByID(_ context.Context, id string) (*domain.Question, error) {
	if q, ok := m.questions[id]; ok {
		return q, nil
	}
	return nil, domain.ErrQuestionNotFound
}

func (m *mockQuestionRepository) List(_ context.Context) ([]*domain.Question, error) {
	questions := make([]*domain.Question, 0, len(m.questions))
	for _, q := range m.questions {
		questions = append(questions, q)
	}
	return questions, nil
}

func (m *mockQuestionRepository) Save(_ context.Context, question *domain.Question) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.questions[question.ID] = question
	return nil
}

func (m *mockQuestionRepository) FindByTagID(_ context.Context, tagID string) ([]*domain.Question, error) {
	var result []*domain.Question
	for _, q := range m.questions {
		for _, tid := range q.Tags {
			if tid == tagID {
				result = append(result, q)
				break
			}
		}
	}
	return result, nil
}

func (m *mockQuestionRepository) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.questions[id]; !ok {
		return domain.ErrQuestionNotFound
	}
	delete(m.questions, id)
	return nil
}

// --- テストヘルパー ---

// testQuestion はテスト用の問題エンティティを生成します。
func testQuestion(id, title, createdBy string) *domain.Question {
	return &domain.Question{
		ID:               id,
		Title:            title,
		Body:             "## 問題\nテスト問題文",
		Answer:           "## 解答\nテスト解答",
		Explanation:      "## 解説\nテスト解説",
		Memo:             "## メモ\nテストメモ",
		Tags:             []string{"テスト"},
		Status:           domain.QuestionStatusDraft,
		VisibilityScope:  domain.VisibilityScopeAll,
		PublishedTeamIDs: []string{},
		CreatedBy:        createdBy,
	}
}

// --- QuestionUseCase のテスト ---

// TestQuestionUseCase_CreateQuestion_Success は正常系の問題作成テストです。
func TestQuestionUseCase_CreateQuestion_Success(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	q, err := uc.CreateQuestion(context.Background(), usecase.CreateQuestionInput{
		CallerID:    "user-1",
		Title:       "テスト問題",
		Body:        "## 問題\n問題文",
		Answer:      "## 解答\n解答",
		Explanation: "## 解説\n解説",
		Memo:        "## メモ\nメモ",
		Tags:        []string{"CISSP"},
	})

	if err != nil {
		t.Fatalf("問題作成に失敗しました: %v", err)
	}
	if q.Title != "テスト問題" {
		t.Errorf("タイトルが期待値と異なります: got %s, want テスト問題", q.Title)
	}
	// デフォルト値の確認
	if q.Status != domain.QuestionStatusDraft {
		t.Errorf("statusが期待値と異なります: got %s, want draft", q.Status)
	}
	if q.VisibilityScope != domain.VisibilityScopeAll {
		t.Errorf("visibility_scopeが期待値と異なります: got %s, want all", q.VisibilityScope)
	}
	if q.CreatedBy != "user-1" {
		t.Errorf("created_byが期待値と異なります: got %s, want user-1", q.CreatedBy)
	}
	if q.ID == "" {
		t.Error("IDが生成されていません")
	}
}

// TestQuestionUseCase_CreateQuestion_EmptyTitle はタイトルが空の場合のバリデーションテストです。
func TestQuestionUseCase_CreateQuestion_EmptyTitle(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	_, err := uc.CreateQuestion(context.Background(), usecase.CreateQuestionInput{
		CallerID: "user-1",
		Title:    "", // 空タイトル
		Body:     "問題文",
	})

	if err == nil {
		t.Fatal("タイトルが空の場合はエラーが返されるべきです")
	}
}

// TestQuestionUseCase_CreateQuestion_InvalidStatus は無効なステータスのテストです。
func TestQuestionUseCase_CreateQuestion_InvalidStatus(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	_, err := uc.CreateQuestion(context.Background(), usecase.CreateQuestionInput{
		CallerID: "user-1",
		Title:    "テスト問題",
		Status:   domain.QuestionStatus("invalid"),
	})

	if !errors.Is(err, domain.ErrInvalidQuestionStatus) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidQuestionStatus)
	}
}

// TestQuestionUseCase_CreateQuestion_InvalidVisibilityScope は無効な公開範囲のテストです。
func TestQuestionUseCase_CreateQuestion_InvalidVisibilityScope(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	_, err := uc.CreateQuestion(context.Background(), usecase.CreateQuestionInput{
		CallerID:        "user-1",
		Title:           "テスト問題",
		VisibilityScope: domain.VisibilityScope("invalid"),
	})

	if !errors.Is(err, domain.ErrInvalidVisibilityScope) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidVisibilityScope)
	}
}

// TestQuestionUseCase_ListQuestions_Success は正常系の問題一覧取得テストです。
// draft 問題は作成者本人のみ返ることを確認します。
func TestQuestionUseCase_ListQuestions_Success(t *testing.T) {
	repo := newMockQuestionRepository()
	// user-1 が作成した draft 問題
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))
	// user-2 が作成した draft 問題
	repo.addQuestion(testQuestion("q-2", "問題2", "user-2"))

	uc := newQuestionUseCase(repo)

	// user-1 として取得 → 自分の draft のみ1件
	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 1", len(questions))
	}
}

// TestQuestionUseCase_ListQuestions_AdminGetsAll は admin が全件取得できることのテストです。
func TestQuestionUseCase_ListQuestions_AdminGetsAll(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))
	repo.addQuestion(testQuestion("q-2", "問題2", "user-2"))

	uc := newQuestionUseCase(repo)

	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "admin-1",
		CallerRole: domain.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(questions))
	}
}

// TestQuestionUseCase_ListQuestions_Empty は問題が0件の場合のテストです。
func TestQuestionUseCase_ListQuestions_Empty(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 0", len(questions))
	}
}

// TestQuestionUseCase_GetQuestion_Success は正常系の問題詳細取得テストです（作成者本人）。
func TestQuestionUseCase_GetQuestion_Success(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	q, err := uc.GetQuestion(context.Background(), "q-1", usecase.GetQuestionInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題取得に失敗しました: %v", err)
	}
	if q.ID != "q-1" {
		t.Errorf("IDが期待値と異なります: got %s, want q-1", q.ID)
	}
}

// TestQuestionUseCase_GetQuestion_NotFound は存在しないIDで取得した場合のテストです。
func TestQuestionUseCase_GetQuestion_NotFound(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	_, err := uc.GetQuestion(context.Background(), "nonexistent-id", usecase.GetQuestionInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestQuestionUseCase_GetQuestion_VisibilityDenied は閲覧不可の問題取得が404になるテストです。
func TestQuestionUseCase_GetQuestion_VisibilityDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	// user-1 が作成した draft 問題（user-2 からは見えない）
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	_, err := uc.GetQuestion(context.Background(), "q-1", usecase.GetQuestionInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("閲覧不可の問題は404になるべきです: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestQuestionUseCase_UpdateQuestion_ByOwner は作成者本人による問題更新テストです。
func TestQuestionUseCase_UpdateQuestion_ByOwner(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "元のタイトル", "user-1"))

	uc := newQuestionUseCase(repo)

	newTitle := "更新後のタイトル"
	q, err := uc.UpdateQuestion(context.Background(), "q-1", usecase.UpdateQuestionInput{
		CallerID:   "user-1", // 作成者本人
		CallerRole: domain.RoleUser,
		Title:      &newTitle,
	})

	if err != nil {
		t.Fatalf("問題更新に失敗しました: %v", err)
	}
	if q.Title != "更新後のタイトル" {
		t.Errorf("タイトルが期待値と異なります: got %s, want 更新後のタイトル", q.Title)
	}
}

// TestQuestionUseCase_UpdateQuestion_ByAdmin は admin ロールによる問題更新テストです。
func TestQuestionUseCase_UpdateQuestion_ByAdmin(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "元のタイトル", "user-1"))

	uc := newQuestionUseCase(repo)

	newTitle := "adminが更新"
	q, err := uc.UpdateQuestion(context.Background(), "q-1", usecase.UpdateQuestionInput{
		CallerID:   "admin-1", // 別ユーザーだが admin ロール
		CallerRole: domain.RoleAdmin,
		Title:      &newTitle,
	})

	if err != nil {
		t.Fatalf("問題更新に失敗しました: %v", err)
	}
	if q.Title != "adminが更新" {
		t.Errorf("タイトルが期待値と異なります: got %s, want adminが更新", q.Title)
	}
}

// TestQuestionUseCase_UpdateQuestion_PermissionDenied は権限のないユーザーによる更新テストです。
func TestQuestionUseCase_UpdateQuestion_PermissionDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	newTitle := "不正な更新"
	_, err := uc.UpdateQuestion(context.Background(), "q-1", usecase.UpdateQuestionInput{
		CallerID:   "user-2", // 作成者でも admin でもない
		CallerRole: domain.RoleUser,
		Title:      &newTitle,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestQuestionUseCase_UpdateQuestion_NotFound は存在しないIDで更新した場合のテストです。
func TestQuestionUseCase_UpdateQuestion_NotFound(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	title := "更新"
	_, err := uc.UpdateQuestion(context.Background(), "nonexistent-id", usecase.UpdateQuestionInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
		Title:      &title,
	})

	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestQuestionUseCase_UpdateQuestion_EmptyTitle はタイトルを空にする更新テストです。
func TestQuestionUseCase_UpdateQuestion_EmptyTitle(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	emptyTitle := ""
	_, err := uc.UpdateQuestion(context.Background(), "q-1", usecase.UpdateQuestionInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
		Title:      &emptyTitle,
	})

	if err == nil {
		t.Fatal("タイトルを空にする場合はエラーが返されるべきです")
	}
}

// TestQuestionUseCase_DeleteQuestion_ByOwner は作成者本人による問題削除テストです。
func TestQuestionUseCase_DeleteQuestion_ByOwner(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "q-1", "user-1", domain.RoleUser)
	if err != nil {
		t.Fatalf("問題削除に失敗しました: %v", err)
	}

	// 削除後に取得しようとすると NotFound になること
	_, err = repo.FindByID(context.Background(), "q-1")
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Error("削除後も問題が存在しています")
	}
}

// TestQuestionUseCase_DeleteQuestion_ByAdmin は admin ロールによる問題削除テストです。
func TestQuestionUseCase_DeleteQuestion_ByAdmin(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "q-1", "admin-1", domain.RoleAdmin)
	if err != nil {
		t.Fatalf("問題削除に失敗しました: %v", err)
	}
}

// TestQuestionUseCase_DeleteQuestion_PermissionDenied は権限のないユーザーによる削除テストです。
func TestQuestionUseCase_DeleteQuestion_PermissionDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "q-1", "user-2", domain.RoleUser)
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestQuestionUseCase_DeleteQuestion_NotFound は存在しないIDで削除した場合のテストです。
func TestQuestionUseCase_DeleteQuestion_NotFound(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "nonexistent-id", "user-1", domain.RoleUser)
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// --- UpdateQuestionVisibility のテスト ---

// testPublishedQuestion は公開済みの問題エンティティを生成します（テストヘルパー）。
func testPublishedQuestion(id, createdBy string, scope domain.VisibilityScope, teamIDs []string) *domain.Question {
	q := testQuestion(id, "公開問題", createdBy)
	q.Status = domain.QuestionStatusPublished
	q.VisibilityScope = scope
	q.PublishedTeamIDs = teamIDs
	return q
}

// TestQuestionUseCase_UpdateQuestionVisibility_ByOwner は作成者による公開設定変更テストです。
func TestQuestionUseCase_UpdateQuestionVisibility_ByOwner(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	vs := domain.VisibilityScopeAll
	q, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", usecase.UpdateQuestionVisibilityInput{
		CallerID:        "user-1",
		CallerRole:      domain.RoleUser,
		Status:          domain.QuestionStatusPublished,
		VisibilityScope: &vs,
	})

	if err != nil {
		t.Fatalf("公開設定変更に失敗しました: %v", err)
	}
	if q.Status != domain.QuestionStatusPublished {
		t.Errorf("statusが期待値と異なります: got %s, want published", q.Status)
	}
	if q.VisibilityScope != domain.VisibilityScopeAll {
		t.Errorf("visibility_scopeが期待値と異なります: got %s, want all", q.VisibilityScope)
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_ByAdmin は admin による公開設定変更テストです。
func TestQuestionUseCase_UpdateQuestionVisibility_ByAdmin(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	q, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", usecase.UpdateQuestionVisibilityInput{
		CallerID:   "admin-1",
		CallerRole: domain.RoleAdmin,
		Status:     domain.QuestionStatusPrivate,
	})

	if err != nil {
		t.Fatalf("公開設定変更に失敗しました: %v", err)
	}
	if q.Status != domain.QuestionStatusPrivate {
		t.Errorf("statusが期待値と異なります: got %s, want private", q.Status)
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_PermissionDenied は権限のないユーザーによる変更テストです。
func TestQuestionUseCase_UpdateQuestionVisibility_PermissionDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	_, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", usecase.UpdateQuestionVisibilityInput{
		CallerID:   "user-2", // 作成者でも admin でもない
		CallerRole: domain.RoleUser,
		Status:     domain.QuestionStatusPublished,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_NotFound は存在しないIDで変更した場合のテストです。
func TestQuestionUseCase_UpdateQuestionVisibility_NotFound(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	_, err := uc.UpdateQuestionVisibility(context.Background(), "nonexistent-id", usecase.UpdateQuestionVisibilityInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
		Status:     domain.QuestionStatusPublished,
	})

	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_InvalidStatus は無効なステータスのテストです。
func TestQuestionUseCase_UpdateQuestionVisibility_InvalidStatus(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	_, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", usecase.UpdateQuestionVisibilityInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
		Status:     domain.QuestionStatus("invalid"),
	})

	if !errors.Is(err, domain.ErrInvalidQuestionStatus) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidQuestionStatus)
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_InvalidScope は無効な公開範囲のテストです。
func TestQuestionUseCase_UpdateQuestionVisibility_InvalidScope(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	vs := domain.VisibilityScope("invalid")
	_, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", usecase.UpdateQuestionVisibilityInput{
		CallerID:        "user-1",
		CallerRole:      domain.RoleUser,
		Status:          domain.QuestionStatusPublished,
		VisibilityScope: &vs,
	})

	if !errors.Is(err, domain.ErrInvalidVisibilityScope) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidVisibilityScope)
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_TeamScope はチーム公開設定のテストです。
func TestQuestionUseCase_UpdateQuestionVisibility_TeamScope(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	vs := domain.VisibilityScopeTeam
	teamIDs := []string{"team-1", "team-2"}
	q, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", usecase.UpdateQuestionVisibilityInput{
		CallerID:            "user-1",
		CallerRole:          domain.RoleUser,
		Status:              domain.QuestionStatusPublished,
		VisibilityScope:     &vs,
		PublishedTeamIDs:    teamIDs,
		PublishedTeamIDsSet: true,
	})

	if err != nil {
		t.Fatalf("公開設定変更に失敗しました: %v", err)
	}
	if q.VisibilityScope != domain.VisibilityScopeTeam {
		t.Errorf("visibility_scopeが期待値と異なります: got %s, want team", q.VisibilityScope)
	}
	if len(q.PublishedTeamIDs) != 2 {
		t.Errorf("published_team_ids の件数が期待値と異なります: got %d, want 2", len(q.PublishedTeamIDs))
	}
}

// --- ListQuestions 可視性フィルタリングのテスト ---

// TestQuestionUseCase_ListQuestions_PublishedAll は全公開問題が全ユーザーに返るテストです。
func TestQuestionUseCase_ListQuestions_PublishedAll(t *testing.T) {
	repo := newMockQuestionRepository()
	// user-1 が公開した問題（全体公開）
	q := testPublishedQuestion("q-1", "user-1", domain.VisibilityScopeAll, []string{})
	repo.addQuestion(q)

	uc := newQuestionUseCase(repo)

	// user-2 として取得 → 全体公開なので見える
	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 1", len(questions))
	}
}

// TestQuestionUseCase_ListQuestions_PublishedTeam_Visible はチームメンバーにチーム公開問題が返るテストです。
func TestQuestionUseCase_ListQuestions_PublishedTeam_Visible(t *testing.T) {
	qRepo := newMockQuestionRepository()
	tRepo := newMockTeamRepository()

	// user-1 が team-1 に公開した問題
	q := testPublishedQuestion("q-1", "user-1", domain.VisibilityScopeTeam, []string{"team-1"})
	qRepo.addQuestion(q)

	// user-2 が team-1 のオーナー
	tRepo.addTeam(&domain.Team{ID: "team-1", Name: "チーム1", OwnerID: "user-2"})

	uc := usecase.NewQuestionUseCase(qRepo, tRepo)

	// user-2 として取得 → team-1 のオーナーなので見える
	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 1", len(questions))
	}
}

// TestQuestionUseCase_ListQuestions_PublishedTeam_NotVisible はチーム非所属ユーザーにチーム公開問題が返らないテストです。
func TestQuestionUseCase_ListQuestions_PublishedTeam_NotVisible(t *testing.T) {
	qRepo := newMockQuestionRepository()
	tRepo := newMockTeamRepository()

	// user-1 が team-1 に公開した問題
	q := testPublishedQuestion("q-1", "user-1", domain.VisibilityScopeTeam, []string{"team-1"})
	qRepo.addQuestion(q)

	// user-3 はどのチームにも所属していない
	uc := usecase.NewQuestionUseCase(qRepo, tRepo)

	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-3",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("チーム非所属ユーザーには見えないはずです: got %d, want 0", len(questions))
	}
}

// TestQuestionUseCase_ListQuestions_DraftOnlyOwner は draft 問題が作成者のみに返るテストです。
func TestQuestionUseCase_ListQuestions_DraftOnlyOwner(t *testing.T) {
	repo := newMockQuestionRepository()
	// user-1 の draft 問題
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	// user-2 には返らない
	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("draft 問題は作成者以外には見えないはずです: got %d, want 0", len(questions))
	}

	// user-1 には返る
	questions, err = uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("draft 問題は作成者には見えるはずです: got %d, want 1", len(questions))
	}
}

// TestQuestionUseCase_ListQuestions_PublishedTeam_MemberVisible は
// TeamMember レコード経由で追加されたメンバーにチーム公開問題が返るテストです（指摘 #2 対応）。
func TestQuestionUseCase_ListQuestions_PublishedTeam_MemberVisible(t *testing.T) {
	qRepo := newMockQuestionRepository()
	tRepo := newMockTeamRepository()

	// user-1 が team-1 に公開した問題
	q := testPublishedQuestion("q-1", "user-1", domain.VisibilityScopeTeam, []string{"team-1"})
	qRepo.addQuestion(q)

	// team-1 を登録（オーナーは user-1 自身）
	tRepo.addTeam(&domain.Team{ID: "team-1", Name: "チーム1", OwnerID: "user-1"})

	// user-3 を team-1 の TeamMember として追加（オーナーではなくメンバー）
	tRepo.members = append(tRepo.members, domain.TeamMember{TeamID: "team-1", UserID: "user-3"})

	uc := usecase.NewQuestionUseCase(qRepo, tRepo)

	// user-3 として取得 → team-1 のメンバーなので見える
	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-3",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("チームメンバーには公開問題が見えるはずです: got %d, want 1", len(questions))
	}
}

// TestQuestionUseCase_ListQuestions_PublishedTeam_NonMemberNotVisible は
// team-1 に所属していないユーザーにはチーム公開問題が返らないテストです（指摘 #2 対応）。
func TestQuestionUseCase_ListQuestions_PublishedTeam_NonMemberNotVisible(t *testing.T) {
	qRepo := newMockQuestionRepository()
	tRepo := newMockTeamRepository()

	// user-1 が team-1 に公開した問題
	q := testPublishedQuestion("q-1", "user-1", domain.VisibilityScopeTeam, []string{"team-1"})
	qRepo.addQuestion(q)

	// team-1 を登録（user-3 はメンバーとして追加しない）
	tRepo.addTeam(&domain.Team{ID: "team-1", Name: "チーム1", OwnerID: "user-1"})
	tRepo.members = append(tRepo.members, domain.TeamMember{TeamID: "team-1", UserID: "user-3"})

	// user-4 は team-1 に所属していない
	uc := usecase.NewQuestionUseCase(qRepo, tRepo)

	questions, err := uc.ListQuestions(context.Background(), usecase.ListQuestionsInput{
		CallerID:   "user-4",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("チーム非所属ユーザーには見えないはずです: got %d, want 0", len(questions))
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_TooManyTeamIDs は
// published_team_ids の上限超過のバリデーションテストです（指摘 #5 対応）。
func TestQuestionUseCase_UpdateQuestionVisibility_TooManyTeamIDs(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := newQuestionUseCase(repo)

	// 51件のチームIDを生成（上限50件を超える）
	teamIDs := make([]string, 51)
	for i := range teamIDs {
		teamIDs[i] = fmt.Sprintf("team-%d", i+1)
	}

	vs := domain.VisibilityScopeTeam
	_, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", usecase.UpdateQuestionVisibilityInput{
		CallerID:            "user-1",
		CallerRole:          domain.RoleUser,
		Status:              domain.QuestionStatusPublished,
		VisibilityScope:     &vs,
		PublishedTeamIDs:    teamIDs,
		PublishedTeamIDsSet: true,
	})

	if err == nil {
		t.Fatal("published_team_ids が上限を超える場合はエラーが返されるべきです")
	}
}
