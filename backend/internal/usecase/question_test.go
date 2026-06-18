package usecase_test

import (
	"context"
	"errors"
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
	uc := usecase.NewQuestionUseCase(repo)

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
	uc := usecase.NewQuestionUseCase(repo)

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
	uc := usecase.NewQuestionUseCase(repo)

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
	uc := usecase.NewQuestionUseCase(repo)

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
func TestQuestionUseCase_ListQuestions_Success(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))
	repo.addQuestion(testQuestion("q-2", "問題2", "user-2"))

	uc := usecase.NewQuestionUseCase(repo)

	questions, err := uc.ListQuestions(context.Background())
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
	uc := usecase.NewQuestionUseCase(repo)

	questions, err := uc.ListQuestions(context.Background())
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 0", len(questions))
	}
}

// TestQuestionUseCase_GetQuestion_Success は正常系の問題詳細取得テストです。
func TestQuestionUseCase_GetQuestion_Success(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := usecase.NewQuestionUseCase(repo)

	q, err := uc.GetQuestion(context.Background(), "q-1")
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
	uc := usecase.NewQuestionUseCase(repo)

	_, err := uc.GetQuestion(context.Background(), "nonexistent-id")
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestQuestionUseCase_UpdateQuestion_ByOwner は作成者本人による問題更新テストです。
func TestQuestionUseCase_UpdateQuestion_ByOwner(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "元のタイトル", "user-1"))

	uc := usecase.NewQuestionUseCase(repo)

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

	uc := usecase.NewQuestionUseCase(repo)

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

	uc := usecase.NewQuestionUseCase(repo)

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
	uc := usecase.NewQuestionUseCase(repo)

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

	uc := usecase.NewQuestionUseCase(repo)

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

	uc := usecase.NewQuestionUseCase(repo)

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

	uc := usecase.NewQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "q-1", "admin-1", domain.RoleAdmin)
	if err != nil {
		t.Fatalf("問題削除に失敗しました: %v", err)
	}
}

// TestQuestionUseCase_DeleteQuestion_PermissionDenied は権限のないユーザーによる削除テストです。
func TestQuestionUseCase_DeleteQuestion_PermissionDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", "user-1"))

	uc := usecase.NewQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "q-1", "user-2", domain.RoleUser)
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestQuestionUseCase_DeleteQuestion_NotFound は存在しないIDで削除した場合のテストです。
func TestQuestionUseCase_DeleteQuestion_NotFound(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := usecase.NewQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "nonexistent-id", "user-1", domain.RoleUser)
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}
