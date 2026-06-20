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
// callerID が testTeamID のメンバーとして登録された teamRepo を使います。
func newQuestionUseCase(qRepo *mockQuestionRepository) *usecase.QuestionUseCase {
	tRepo := newMockTeamRepository()
	// testCallerID ("user-1") を testTeamID ("team-1") のメンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID: testTeamID,
		UserID: testCallerID,
		Role:   domain.MemberRoleMember,
	})
	return usecase.NewQuestionUseCase(qRepo, tRepo)
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

func (m *mockQuestionRepository) ListByTeam(_ context.Context, teamID string) ([]*domain.Question, error) {
	questions := make([]*domain.Question, 0)
	for _, q := range m.questions {
		if q.TeamID == teamID {
			questions = append(questions, q)
		}
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

func (m *mockQuestionRepository) SearchByTeam(_ context.Context, teamID string, filter domain.QuestionSearchFilter) ([]*domain.Question, error) {
	var result []*domain.Question
	for _, q := range m.questions {
		// チームIDフィルタ
		if q.TeamID != teamID {
			continue
		}
		// タグANDフィルタリング
		if len(filter.TagIDs) > 0 {
			tagSet := make(map[string]struct{}, len(q.Tags))
			for _, t := range q.Tags {
				tagSet[t] = struct{}{}
			}
			match := true
			for _, tid := range filter.TagIDs {
				if _, ok := tagSet[tid]; !ok {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}
		// キーワードフィルタリング
		if filter.Keyword != "" {
			kw := filter.Keyword
			if !contains(q.Title, kw) && !contains(q.Body, kw) && !contains(q.Explanation, kw) && !contains(q.Memo, kw) {
				continue
			}
		}
		result = append(result, q)
	}
	return result, nil
}

// contains は s が substr を含むかどうかを返すヘルパーです。
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
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

// testQuestion はテスト用の問題エンティティを生成します（testTeamID に所属）。
func testQuestion(id, title, createdBy string) *domain.Question {
	return &domain.Question{
		ID:          id,
		TeamID:      testTeamID,
		Title:       title,
		Body:        "## 問題\nテスト問題文",
		Answer:      "## 解答\nテスト解答",
		Explanation: "## 解説\nテスト解説",
		Memo:        "## メモ\nテストメモ",
		Tags:        []string{"テスト"},
		Status:      domain.QuestionStatusDraft,
		CreatedBy:   createdBy,
	}
}

// newQuestionUseCaseWithTeam はチームとメンバーを自由に設定できる QuestionUseCase を生成します。
func newQuestionUseCaseWithTeam(qRepo *mockQuestionRepository, tRepo *mockTeamRepository) *usecase.QuestionUseCase {
	return usecase.NewQuestionUseCase(qRepo, tRepo)
}

// --- QuestionUseCase のテスト ---

// TestQuestionUseCase_CreateQuestion_Success は正常系の問題作成テストです。
func TestQuestionUseCase_CreateQuestion_Success(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	q, err := uc.CreateQuestion(context.Background(), testTeamID, usecase.CreateQuestionInput{
		CallerID:    testCallerID,
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
	if q.TeamID != testTeamID {
		t.Errorf("teamIDが期待値と異なります: got %s, want %s", q.TeamID, testTeamID)
	}
	// デフォルト値の確認
	if q.Status != domain.QuestionStatusDraft {
		t.Errorf("statusが期待値と異なります: got %s, want draft", q.Status)
	}
	if q.CreatedBy != testCallerID {
		t.Errorf("created_byが期待値と異なります: got %s, want %s", q.CreatedBy, testCallerID)
	}
	if q.ID == "" {
		t.Error("IDが生成されていません")
	}
}

// TestQuestionUseCase_CreateQuestion_NonMemberDenied はチーム非メンバーが作成できないテストです。
func TestQuestionUseCase_CreateQuestion_NonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo) // testCallerID のみメンバー登録済み

	_, err := uc.CreateQuestion(context.Background(), testTeamID, usecase.CreateQuestionInput{
		CallerID: "user-not-member",
		Title:    "テスト問題",
	})

	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("チーム非メンバーは 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_CreateQuestion_AdminNonMemberDenied は admin でもチーム非メンバーなら作成できないテストです。
func TestQuestionUseCase_CreateQuestion_AdminNonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	// admin-1 はどのチームにもメンバー登録されていない
	tRepo := newMockTeamRepository()
	uc := usecase.NewQuestionUseCase(repo, tRepo)

	_, err := uc.CreateQuestion(context.Background(), testTeamID, usecase.CreateQuestionInput{
		CallerID: "admin-1",
		Title:    "テスト問題",
	})

	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("admin もチーム非メンバーなら 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_CreateQuestion_EmptyTitle はタイトルが空の場合のバリデーションテストです。
func TestQuestionUseCase_CreateQuestion_EmptyTitle(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	_, err := uc.CreateQuestion(context.Background(), testTeamID, usecase.CreateQuestionInput{
		CallerID: testCallerID,
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

	_, err := uc.CreateQuestion(context.Background(), testTeamID, usecase.CreateQuestionInput{
		CallerID: testCallerID,
		Title:    "テスト問題",
		Status:   domain.QuestionStatus("invalid"),
	})

	if !errors.Is(err, domain.ErrInvalidQuestionStatus) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidQuestionStatus)
	}
}

// TestQuestionUseCase_ListQuestions_Success は正常系の問題一覧取得テストです。
// draft 問題は作成者本人のみ返ることを確認します。
func TestQuestionUseCase_ListQuestions_Success(t *testing.T) {
	repo := newMockQuestionRepository()
	tRepo := newMockTeamRepository()

	// user-1 が作成した draft 問題
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))
	// user-2 が作成した draft 問題（testTeamID に所属）
	repo.addQuestion(testQuestion("q-2", "問題2", "user-2"))

	// testCallerID と "user-2" を team-1 のメンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	// testCallerID として取得 → 自分の draft のみ1件（user-2 の draft は非表示）
	questions, err := uc.ListQuestions(context.Background(), testTeamID, usecase.ListQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 1", len(questions))
	}
}

// TestQuestionUseCase_ListQuestions_NonMemberDenied はチーム非メンバーがアクセスできないテストです。
func TestQuestionUseCase_ListQuestions_NonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	uc := newQuestionUseCase(repo) // testCallerID のみメンバー登録済み

	_, err := uc.ListQuestions(context.Background(), testTeamID, usecase.ListQuestionsInput{
		CallerID:   "user-not-member",
		CallerRole: domain.RoleUser,
	})
	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("チーム非メンバーは 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_ListQuestions_AdminNonMemberDenied は admin もチーム非メンバーなら 403 になるテストです。
func TestQuestionUseCase_ListQuestions_AdminNonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	// admin はどのチームにもメンバー登録されていない
	tRepo := newMockTeamRepository()
	uc := usecase.NewQuestionUseCase(repo, tRepo)

	_, err := uc.ListQuestions(context.Background(), testTeamID, usecase.ListQuestionsInput{
		CallerID:   "admin-1",
		CallerRole: domain.RoleAdmin,
	})
	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("admin もチーム非メンバーなら 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_ListQuestions_Empty は問題が0件の場合のテストです。
func TestQuestionUseCase_ListQuestions_Empty(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	questions, err := uc.ListQuestions(context.Background(), testTeamID, usecase.ListQuestionsInput{
		CallerID:   testCallerID,
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
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	uc := newQuestionUseCase(repo)

	q, err := uc.GetQuestion(context.Background(), "q-1", testTeamID, usecase.GetQuestionInput{
		CallerID:   testCallerID,
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

	_, err := uc.GetQuestion(context.Background(), "nonexistent-id", testTeamID, usecase.GetQuestionInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestQuestionUseCase_GetQuestion_TeamMismatch は別チームの問題取得が404になるテストです。
func TestQuestionUseCase_GetQuestion_TeamMismatch(t *testing.T) {
	repo := newMockQuestionRepository()
	// q-1 は team-2 に所属（testTeamID = "team-1" ではない）
	q := testQuestion("q-1", "問題1", testCallerID)
	q.TeamID = "team-2"
	repo.addQuestion(q)

	uc := newQuestionUseCase(repo)

	_, err := uc.GetQuestion(context.Background(), "q-1", testTeamID, usecase.GetQuestionInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("チーム不一致の場合は404になるべきです: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestQuestionUseCase_GetQuestion_DraftVisibilityDenied は draft 問題が作成者以外から見えないテストです。
func TestQuestionUseCase_GetQuestion_DraftVisibilityDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	// user-1 が作成した draft 問題（testCallerID からは見える、user-2 からは見えない）
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID と "user-2" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	_, err := uc.GetQuestion(context.Background(), "q-1", testTeamID, usecase.GetQuestionInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("draft 問題は作成者以外には404になるべきです: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestQuestionUseCase_GetQuestion_NonMemberDenied はチーム非メンバーが取得できないテストです。
func TestQuestionUseCase_GetQuestion_NonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	uc := newQuestionUseCase(repo) // testCallerID のみメンバー登録済み

	_, err := uc.GetQuestion(context.Background(), "q-1", testTeamID, usecase.GetQuestionInput{
		CallerID:   "user-not-member",
		CallerRole: domain.RoleUser,
	})
	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("チーム非メンバーは 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_UpdateQuestion_ByOwner は作成者本人による問題更新テストです。
func TestQuestionUseCase_UpdateQuestion_ByOwner(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "元のタイトル", testCallerID))

	uc := newQuestionUseCase(repo)

	newTitle := "更新後のタイトル"
	q, err := uc.UpdateQuestion(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionInput{
		CallerID:   testCallerID, // 作成者本人
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

// TestQuestionUseCase_UpdateQuestion_ByAdmin は admin ロールによる問題更新テストです（チームメンバーである前提）。
func TestQuestionUseCase_UpdateQuestion_ByAdmin(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "元のタイトル", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID と "admin-1" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "admin-1", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	newTitle := "adminが更新"
	q, err := uc.UpdateQuestion(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionInput{
		CallerID:   "admin-1", // チームメンバーかつ admin ロール
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

// TestQuestionUseCase_UpdateQuestion_AdminNonMemberDenied は admin でもチーム非メンバーなら更新できないテストです。
func TestQuestionUseCase_UpdateQuestion_AdminNonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "元のタイトル", testCallerID))

	// admin-1 はどのチームにもメンバー登録されていない
	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	newTitle := "adminが更新"
	_, err := uc.UpdateQuestion(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionInput{
		CallerID:   "admin-1", // チームメンバーではない
		CallerRole: domain.RoleAdmin,
		Title:      &newTitle,
	})

	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("admin もチーム非メンバーなら 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_UpdateQuestion_PermissionDenied は権限のないユーザーによる更新テストです。
func TestQuestionUseCase_UpdateQuestion_PermissionDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID と "user-2" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	newTitle := "不正な更新"
	_, err := uc.UpdateQuestion(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionInput{
		CallerID:   "user-2", // チームメンバーだが作成者でも admin でもない
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
	_, err := uc.UpdateQuestion(context.Background(), "nonexistent-id", testTeamID, usecase.UpdateQuestionInput{
		CallerID:   testCallerID,
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
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	uc := newQuestionUseCase(repo)

	emptyTitle := ""
	_, err := uc.UpdateQuestion(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionInput{
		CallerID:   testCallerID,
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
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	uc := newQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "q-1", testTeamID, testCallerID, domain.RoleUser)
	if err != nil {
		t.Fatalf("問題削除に失敗しました: %v", err)
	}

	// 削除後に取得しようとすると NotFound になること
	_, err = repo.FindByID(context.Background(), "q-1")
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Error("削除後も問題が存在しています")
	}
}

// TestQuestionUseCase_DeleteQuestion_ByAdmin は admin ロールによる問題削除テストです（チームメンバーである前提）。
func TestQuestionUseCase_DeleteQuestion_ByAdmin(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID と "admin-1" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "admin-1", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	err := uc.DeleteQuestion(context.Background(), "q-1", testTeamID, "admin-1", domain.RoleAdmin)
	if err != nil {
		t.Fatalf("問題削除に失敗しました: %v", err)
	}
}

// TestQuestionUseCase_DeleteQuestion_AdminNonMemberDenied は admin でもチーム非メンバーなら削除できないテストです。
func TestQuestionUseCase_DeleteQuestion_AdminNonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	err := uc.DeleteQuestion(context.Background(), "q-1", testTeamID, "admin-1", domain.RoleAdmin)
	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("admin もチーム非メンバーなら 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_DeleteQuestion_PermissionDenied は権限のないユーザーによる削除テストです。
func TestQuestionUseCase_DeleteQuestion_PermissionDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID と "user-2" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	err := uc.DeleteQuestion(context.Background(), "q-1", testTeamID, "user-2", domain.RoleUser)
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestQuestionUseCase_DeleteQuestion_NotFound は存在しないIDで削除した場合のテストです。
func TestQuestionUseCase_DeleteQuestion_NotFound(t *testing.T) {
	repo := newMockQuestionRepository()
	uc := newQuestionUseCase(repo)

	err := uc.DeleteQuestion(context.Background(), "nonexistent-id", testTeamID, testCallerID, domain.RoleUser)
	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// --- UpdateQuestionVisibility のテスト ---

// TestQuestionUseCase_UpdateQuestionVisibility_ByOwner は作成者による公開設定変更テストです。
func TestQuestionUseCase_UpdateQuestionVisibility_ByOwner(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	uc := newQuestionUseCase(repo)

	q, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionVisibilityInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Status:     domain.QuestionStatusPublished,
	})

	if err != nil {
		t.Fatalf("公開設定変更に失敗しました: %v", err)
	}
	if q.Status != domain.QuestionStatusPublished {
		t.Errorf("statusが期待値と異なります: got %s, want published", q.Status)
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_ByAdmin は admin による公開設定変更テストです（チームメンバーである前提）。
func TestQuestionUseCase_UpdateQuestionVisibility_ByAdmin(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "admin-1", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	q, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionVisibilityInput{
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

// TestQuestionUseCase_UpdateQuestionVisibility_AdminNonMemberDenied は admin でもチーム非メンバーなら変更できないテストです。
func TestQuestionUseCase_UpdateQuestionVisibility_AdminNonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	_, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionVisibilityInput{
		CallerID:   "admin-1",
		CallerRole: domain.RoleAdmin,
		Status:     domain.QuestionStatusPublished,
	})

	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("admin もチーム非メンバーなら 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_UpdateQuestionVisibility_PermissionDenied は権限のないユーザーによる変更テストです。
func TestQuestionUseCase_UpdateQuestionVisibility_PermissionDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	_, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionVisibilityInput{
		CallerID:   "user-2", // チームメンバーだが作成者でも admin でもない
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

	_, err := uc.UpdateQuestionVisibility(context.Background(), "nonexistent-id", testTeamID, usecase.UpdateQuestionVisibilityInput{
		CallerID:   testCallerID,
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
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	uc := newQuestionUseCase(repo)

	_, err := uc.UpdateQuestionVisibility(context.Background(), "q-1", testTeamID, usecase.UpdateQuestionVisibilityInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Status:     domain.QuestionStatus("invalid"),
	})

	if !errors.Is(err, domain.ErrInvalidQuestionStatus) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidQuestionStatus)
	}
}

// --- ListQuestions 可視性フィルタリングのテスト ---

// testPublishedQuestion は公開済みの問題エンティティを生成します（testTeamID に所属）。
func testPublishedQuestion(id, createdBy string) *domain.Question {
	q := testQuestion(id, "公開問題", createdBy)
	q.Status = domain.QuestionStatusPublished
	return q
}

// TestQuestionUseCase_ListQuestions_PublishedVisibleToAllMembers は
// published 問題がチームメンバー全員に返るテストです。
func TestQuestionUseCase_ListQuestions_PublishedVisibleToAllMembers(t *testing.T) {
	repo := newMockQuestionRepository()
	// testCallerID が作成した published 問題
	repo.addQuestion(testPublishedQuestion("q-1", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID と "user-2" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	// user-2 として取得 → published なので見える
	questions, err := uc.ListQuestions(context.Background(), testTeamID, usecase.ListQuestionsInput{
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

// TestQuestionUseCase_ListQuestions_DraftOnlyOwner は draft 問題が作成者のみに返るテストです。
func TestQuestionUseCase_ListQuestions_DraftOnlyOwner(t *testing.T) {
	repo := newMockQuestionRepository()
	// testCallerID の draft 問題
	repo.addQuestion(testQuestion("q-1", "問題1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	// user-2 には返らない
	questions, err := uc.ListQuestions(context.Background(), testTeamID, usecase.ListQuestionsInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 0 {
		t.Errorf("draft 問題は作成者以外には見えないはずです: got %d, want 0", len(questions))
	}

	// testCallerID には返る
	questions, err = uc.ListQuestions(context.Background(), testTeamID, usecase.ListQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("問題一覧取得に失敗しました: %v", err)
	}
	if len(questions) != 1 {
		t.Errorf("draft 問題は作成者には見えるはずです: got %d, want 1", len(questions))
	}
}

// --- SearchQuestions のテスト ---

// testPublishedQuestionWithFields はキーワード検索テスト用の公開問題エンティティを生成します（testTeamID に所属）。
func testPublishedQuestionWithFields(id, title, body, explanation, memo string, tags []string) *domain.Question {
	return &domain.Question{
		ID:          id,
		TeamID:      testTeamID,
		Title:       title,
		Body:        body,
		Answer:      "## 解答\nテスト解答",
		Explanation: explanation,
		Memo:        memo,
		Tags:        tags,
		Status:      domain.QuestionStatusPublished,
		CreatedBy:   testCallerID,
	}
}

// TestQuestionUseCase_SearchQuestions_NoFilter はフィルタなしで全件返るテストです。
func TestQuestionUseCase_SearchQuestions_NoFilter(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testPublishedQuestionWithFields("q-1", "問題1", "本文1", "解説1", "メモ1", []string{"tag-1"}))
	repo.addQuestion(testPublishedQuestionWithFields("q-2", "問題2", "本文2", "解説2", "メモ2", []string{"tag-2"}))

	uc := newQuestionUseCase(repo)

	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("総件数が期待値と異なります: got %d, want 2", result.Total)
	}
	if len(result.Items) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(result.Items))
	}
}

// TestQuestionUseCase_SearchQuestions_NonMemberDenied はチーム非メンバーが検索できないテストです。
func TestQuestionUseCase_SearchQuestions_NonMemberDenied(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testPublishedQuestionWithFields("q-1", "問題1", "本文1", "解説1", "メモ1", []string{}))

	uc := newQuestionUseCase(repo) // testCallerID のみメンバー登録済み

	_, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   "user-not-member",
		CallerRole: domain.RoleUser,
		Page:       1,
		PerPage:    20,
	})
	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("チーム非メンバーは 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestQuestionUseCase_SearchQuestions_TagFilter はタグIDでフィルタリングするテストです。
func TestQuestionUseCase_SearchQuestions_TagFilter(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testPublishedQuestionWithFields("q-1", "問題1", "本文1", "解説1", "メモ1", []string{"tag-1", "tag-2"}))
	repo.addQuestion(testPublishedQuestionWithFields("q-2", "問題2", "本文2", "解説2", "メモ2", []string{"tag-2", "tag-3"}))
	repo.addQuestion(testPublishedQuestionWithFields("q-3", "問題3", "本文3", "解説3", "メモ3", []string{"tag-3"}))

	uc := newQuestionUseCase(repo)

	// tag-2 のみ指定 → q-1, q-2 の2件
	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		TagIDs:     []string{"tag-2"},
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("総件数が期待値と異なります: got %d, want 2", result.Total)
	}
}

// TestQuestionUseCase_SearchQuestions_TagFilterAND は複数タグAND絞り込みのテストです。
func TestQuestionUseCase_SearchQuestions_TagFilterAND(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testPublishedQuestionWithFields("q-1", "問題1", "本文1", "解説1", "メモ1", []string{"tag-1", "tag-2"}))
	repo.addQuestion(testPublishedQuestionWithFields("q-2", "問題2", "本文2", "解説2", "メモ2", []string{"tag-2", "tag-3"}))

	uc := newQuestionUseCase(repo)

	// tag-1 AND tag-2 → q-1 のみ1件
	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		TagIDs:     []string{"tag-1", "tag-2"},
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("AND絞り込みの総件数が期待値と異なります: got %d, want 1", result.Total)
	}
}

// TestQuestionUseCase_SearchQuestions_KeywordTitle はタイトルのキーワード検索テストです。
func TestQuestionUseCase_SearchQuestions_KeywordTitle(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testPublishedQuestionWithFields("q-1", "暗号化の基礎", "本文1", "解説1", "メモ1", []string{}))
	repo.addQuestion(testPublishedQuestionWithFields("q-2", "リスク管理", "本文2", "解説2", "メモ2", []string{}))

	uc := newQuestionUseCase(repo)

	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Keyword:    "暗号化",
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("総件数が期待値と異なります: got %d, want 1", result.Total)
	}
}

// TestQuestionUseCase_SearchQuestions_EmptyResult は検索結果0件が空配列と200を返すテストです。
func TestQuestionUseCase_SearchQuestions_EmptyResult(t *testing.T) {
	repo := newMockQuestionRepository()
	repo.addQuestion(testPublishedQuestionWithFields("q-1", "問題1", "本文1", "解説1", "メモ1", []string{"tag-1"}))

	uc := newQuestionUseCase(repo)

	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Keyword:    "存在しないキーワードxyz",
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("検索結果0件でもエラーは返されるべきではありません: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("総件数が期待値と異なります: got %d, want 0", result.Total)
	}
	if result.Items == nil {
		t.Error("Items は nil ではなく空スライスが返されるべきです")
	}
	if len(result.Items) != 0 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 0", len(result.Items))
	}
}

// TestQuestionUseCase_SearchQuestions_Pagination はページネーションのテストです。
func TestQuestionUseCase_SearchQuestions_Pagination(t *testing.T) {
	repo := newMockQuestionRepository()
	// 公開問題を5件作成
	for i := 1; i <= 5; i++ {
		repo.addQuestion(testPublishedQuestionWithFields(
			fmt.Sprintf("q-%d", i),
			fmt.Sprintf("問題%d", i),
			"本文",
			"解説",
			"メモ",
			[]string{},
		))
	}

	uc := newQuestionUseCase(repo)

	// 1ページあたり2件、2ページ目を取得
	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Page:       2,
		PerPage:    2,
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("総件数が期待値と異なります: got %d, want 5", result.Total)
	}
	if result.TotalPages != 3 {
		t.Errorf("総ページ数が期待値と異なります: got %d, want 3", result.TotalPages)
	}
	if result.Page != 2 {
		t.Errorf("ページ番号が期待値と異なります: got %d, want 2", result.Page)
	}
	if len(result.Items) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(result.Items))
	}
}

// TestQuestionUseCase_SearchQuestions_PaginationFirstPage はページネーションの1ページ目テストです。
func TestQuestionUseCase_SearchQuestions_PaginationFirstPage(t *testing.T) {
	repo := newMockQuestionRepository()
	for i := 1; i <= 3; i++ {
		repo.addQuestion(testPublishedQuestionWithFields(
			fmt.Sprintf("q-%d", i),
			fmt.Sprintf("問題%d", i),
			"本文", "解説", "メモ", []string{},
		))
	}

	uc := newQuestionUseCase(repo)

	// デフォルトパラメータ（page=0はpage=1として扱われる）
	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Page:       0, // 0以下は1とみなす
		PerPage:    0, // 0以下はデフォルト20とみなす
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	if result.Page != 1 {
		t.Errorf("ページ番号が期待値と異なります: got %d, want 1", result.Page)
	}
	if result.PerPage != 20 {
		t.Errorf("1ページあたりの件数が期待値と異なります: got %d, want 20", result.PerPage)
	}
	if result.Total != 3 {
		t.Errorf("総件数が期待値と異なります: got %d, want 3", result.Total)
	}
}

// TestQuestionUseCase_SearchQuestions_VisibilityFilter は可視性フィルタが検索に適用されるテストです。
func TestQuestionUseCase_SearchQuestions_VisibilityFilter(t *testing.T) {
	repo := newMockQuestionRepository()
	// testCallerID の draft 問題（"user-2" からは見えない）
	repo.addQuestion(testQuestion("q-1", "テスト問題", testCallerID))
	// testCallerID の公開問題（全メンバーに見える）
	repo.addQuestion(testPublishedQuestionWithFields("q-2", "テスト問題（公開）", "本文", "解説", "メモ", []string{}))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	// user-2 として検索 → 公開問題 q-2 のみ見える
	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
		Keyword:    "テスト",
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("可視性フィルタ後の総件数が期待値と異なります: got %d, want 1", result.Total)
	}
}

// TestQuestionUseCase_SearchQuestions_AdminMemberSeesAll は
// admin かつチームメンバーが全問題を検索できるテストです。
func TestQuestionUseCase_SearchQuestions_AdminMemberSeesAll(t *testing.T) {
	repo := newMockQuestionRepository()
	// draft 問題（作成者のみ見える）
	repo.addQuestion(testQuestion("q-1", "テスト問題（draft）", testCallerID))
	// 公開問題
	repo.addQuestion(testPublishedQuestionWithFields("q-2", "テスト問題（公開）", "本文", "解説", "メモ", []string{}))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "admin-1", Role: domain.MemberRoleMember})

	uc := newQuestionUseCaseWithTeam(repo, tRepo)

	// admin-1 として検索 → チームメンバーなので draft 以外は見える（draft は testCallerID のものなので見えない）
	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   "admin-1",
		CallerRole: domain.RoleAdmin,
		Keyword:    "テスト",
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	// admin-1 は draft の作成者ではないので q-1 は見えない。q-2 のみ見える
	if result.Total != 1 {
		t.Errorf("admin のチームメンバーが見える件数が期待値と異なります: got %d, want 1", result.Total)
	}
}

// TestQuestionUseCase_SearchQuestions_PerPageMax はper_page最大値の正規化テストです。
func TestQuestionUseCase_SearchQuestions_PerPageMax(t *testing.T) {
	repo := newMockQuestionRepository()
	for i := 1; i <= 3; i++ {
		repo.addQuestion(testPublishedQuestionWithFields(
			fmt.Sprintf("q-%d", i),
			fmt.Sprintf("問題%d", i),
			"本文", "解説", "メモ", []string{},
		))
	}

	uc := newQuestionUseCase(repo)

	// per_page=200 は最大値100に正規化される
	result, err := uc.SearchQuestions(context.Background(), testTeamID, usecase.SearchQuestionsInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Page:       1,
		PerPage:    200,
	})
	if err != nil {
		t.Fatalf("問題検索に失敗しました: %v", err)
	}
	if result.PerPage != 100 {
		t.Errorf("per_page が最大値に正規化されていません: got %d, want 100", result.PerPage)
	}
}
