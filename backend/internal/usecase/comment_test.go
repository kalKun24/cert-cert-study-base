package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// --- モック定義 ---

// mockCommentRepository は domain.CommentRepository のモックです。
type mockCommentRepository struct {
	comments  map[string]*domain.Comment // key: questionID/commentID
	saveErr   error
	deleteErr error
}

func newMockCommentRepository() *mockCommentRepository {
	return &mockCommentRepository{
		comments: make(map[string]*domain.Comment),
	}
}

func commentKey(questionID, commentID string) string {
	return questionID + "/" + commentID
}

func (m *mockCommentRepository) addComment(c *domain.Comment) {
	m.comments[commentKey(c.QuestionID, c.ID)] = c
}

func (m *mockCommentRepository) FindByID(_ context.Context, questionID, commentID string) (*domain.Comment, error) {
	if c, ok := m.comments[commentKey(questionID, commentID)]; ok {
		return c, nil
	}
	return nil, domain.ErrCommentNotFound
}

func (m *mockCommentRepository) ListByQuestionID(_ context.Context, questionID string) ([]*domain.Comment, error) {
	var result []*domain.Comment
	for _, c := range m.comments {
		if c.QuestionID == questionID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCommentRepository) Save(_ context.Context, comment *domain.Comment) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.comments[commentKey(comment.QuestionID, comment.ID)] = comment
	return nil
}

func (m *mockCommentRepository) Delete(_ context.Context, questionID, commentID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	key := commentKey(questionID, commentID)
	if _, ok := m.comments[key]; !ok {
		return domain.ErrCommentNotFound
	}
	delete(m.comments, key)
	return nil
}

// --- テストヘルパー ---

// newCommentUseCase はテスト用の CommentUseCase を生成します。
func newCommentUseCase(
	cRepo *mockCommentRepository,
	qRepo *mockQuestionRepository,
	uRepo *mockUserRepository,
) *usecase.CommentUseCase {
	return usecase.NewCommentUseCase(cRepo, qRepo, uRepo, newMockTeamRepository())
}

// testPublishedAllQuestion はテスト用の published+all 公開問題を生成します。
func testPublishedAllQuestion(id, createdBy string) *domain.Question {
	q := testQuestion(id, "公開問題", createdBy)
	q.Status = domain.QuestionStatusPublished
	q.VisibilityScope = domain.VisibilityScopeAll
	return q
}

// testPublishedTeamQuestion はテスト用の published+team 公開問題を生成します。
func testPublishedTeamQuestion(id, createdBy string, teamIDs []string) *domain.Question {
	q := testQuestion(id, "チーム限定問題", createdBy)
	q.Status = domain.QuestionStatusPublished
	q.VisibilityScope = domain.VisibilityScopeTeam
	q.PublishedTeamIDs = teamIDs
	return q
}

// testComment はテスト用のコメントエンティティを生成します。
func testComment(id, questionID, createdBy, body string) *domain.Comment {
	now := time.Now().UTC()
	return &domain.Comment{
		ID:         id,
		QuestionID: questionID,
		Body:       body,
		CreatedBy:  createdBy,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// testUserWithDisplayName はコメントテスト用の表示名を持つユーザーエンティティを生成します。
func testUserWithDisplayName(id, username, displayName string) *domain.User {
	u := testUser(id, username, username+"@example.com", domain.RoleUser, true)
	u.DisplayName = displayName
	return u
}

// --- CreateComment テスト ---

// TestCommentUseCase_CreateComment_Success は正常系のコメント投稿テストです。
func TestCommentUseCase_CreateComment_Success(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	uRepo := newMockUserRepository()
	uRepo.addUser(testUserWithDisplayName("user-1", "alice", "Alice Smith"))

	cRepo := newMockCommentRepository()
	uc := newCommentUseCase(cRepo, qRepo, uRepo)

	result, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "テストコメント",
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("コメント投稿に失敗しました: %v", err)
	}
	if result.Body != "テストコメント" {
		t.Errorf("BodyがBody期待値と異なります: got %s, want テストコメント", result.Body)
	}
	if result.CreatedBy != "user-1" {
		t.Errorf("CreatedByが期待値と異なります: got %s, want user-1", result.CreatedBy)
	}
	if result.DisplayName != "Alice Smith" {
		t.Errorf("DisplayNameが期待値と異なります: got %s, want Alice Smith", result.DisplayName)
	}
	if result.ID == "" {
		t.Error("IDが生成されていません")
	}
}

// TestCommentUseCase_CreateComment_EmptyBody はコメント本文が空の場合のバリデーションテストです。
func TestCommentUseCase_CreateComment_EmptyBody(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "", // 空
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrCommentBodyEmpty) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrCommentBodyEmpty)
	}
}

// TestCommentUseCase_CreateComment_QuestionNotFound は存在しない問題へのコメント投稿テストです。
func TestCommentUseCase_CreateComment_QuestionNotFound(t *testing.T) {
	uc := newCommentUseCase(newMockCommentRepository(), newMockQuestionRepository(), newMockUserRepository())

	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "no-such-question",
		Body:       "コメント",
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestCommentUseCase_CreateComment_DraftQuestionForbidden は draft 問題へのコメント投稿が作成者以外に拒否されることのテストです。
func TestCommentUseCase_CreateComment_DraftQuestionForbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	// draft 問題: owner-1 が作成
	qRepo.addQuestion(testQuestion("q-1", "draft問題", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	// user-1（非作成者）はコメント不可
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "コメント",
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCommentUseCase_CreateComment_DraftQuestionOwnerAllowed は draft 問題へのコメント投稿が作成者本人に許可されることのテストです。
func TestCommentUseCase_CreateComment_DraftQuestionOwnerAllowed(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testQuestion("q-1", "draft問題", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	// owner-1（作成者本人）はコメント可
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "コメント",
		CallerID:   "owner-1",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("作成者本人のコメント投稿に失敗しました: %v", err)
	}
}

// TestCommentUseCase_CreateComment_PublishedAllAllowed は published+all 問題への全ユーザーのコメント投稿が許可されることのテストです。
func TestCommentUseCase_CreateComment_PublishedAllAllowed(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "コメント",
		CallerID:   "any-user",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("published+all 問題へのコメント投稿に失敗しました: %v", err)
	}
}

// TestCommentUseCase_CreateComment_PublishedTeamForbidden は published+team 問題へのチーム外ユーザーのコメント投稿が拒否されることのテストです。
// チーム情報はモックの teamRepo が返すため、team_test.go のモックが teamID を持たない限り拒否されます。
func TestCommentUseCase_CreateComment_PublishedTeamForbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	// チームID "team-A" に属するメンバーのみ閲覧可
	qRepo.addQuestion(testPublishedTeamQuestion("q-1", "owner-1", []string{"team-A"}))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	// "other-user" はいずれのチームにも所属していない → 拒否
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "コメント",
		CallerID:   "other-user",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCommentUseCase_CreateComment_AdminAllowed は admin が draft 問題にコメントを投稿できることのテストです。
func TestCommentUseCase_CreateComment_AdminAllowed(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testQuestion("q-1", "draft問題", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	// admin は全問題にコメント可
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "adminコメント",
		CallerID:   "admin-user",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Fatalf("adminのコメント投稿に失敗しました: %v", err)
	}
}

// --- ListComments テスト ---

// TestCommentUseCase_ListComments_Success は正常系のコメント一覧取得テストです。
func TestCommentUseCase_ListComments_Success(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	uRepo := newMockUserRepository()
	uRepo.addUser(testUserWithDisplayName("user-1", "alice", "Alice Smith"))
	uRepo.addUser(testUserWithDisplayName("user-2", "bob", "Bob Jones"))

	cRepo := newMockCommentRepository()
	now := time.Now().UTC()
	// 投稿順を逆にして挿入し、ソートが正しく機能することを確認
	cRepo.addComment(&domain.Comment{
		ID: "c-2", QuestionID: "q-1", Body: "コメント2", CreatedBy: "user-2",
		CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute),
	})
	cRepo.addComment(&domain.Comment{
		ID: "c-1", QuestionID: "q-1", Body: "コメント1", CreatedBy: "user-1",
		CreatedAt: now, UpdatedAt: now,
	})

	uc := newCommentUseCase(cRepo, qRepo, uRepo)

	result, err := uc.ListComments(context.Background(), usecase.ListCommentsInput{
		QuestionID: "q-1",
		CallerID:   "any-user",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("コメント一覧取得に失敗しました: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("取得件数が期待値と異なります: got %d, want 2", len(result))
	}
	// 昇順ソートの確認
	if result[0].ID != "c-1" {
		t.Errorf("1件目のIDが期待値と異なります: got %s, want c-1", result[0].ID)
	}
	if result[1].ID != "c-2" {
		t.Errorf("2件目のIDが期待値と異なります: got %s, want c-2", result[1].ID)
	}
	// display_name の確認
	if result[0].DisplayName != "Alice Smith" {
		t.Errorf("display_nameが期待値と異なります: got %s, want Alice Smith", result[0].DisplayName)
	}
	if result[1].DisplayName != "Bob Jones" {
		t.Errorf("display_nameが期待値と異なります: got %s, want Bob Jones", result[1].DisplayName)
	}
}

// TestCommentUseCase_ListComments_Forbidden は閲覧権限がないユーザーの一覧取得が拒否されることのテストです。
func TestCommentUseCase_ListComments_Forbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	// draft 問題
	qRepo.addQuestion(testQuestion("q-1", "draft問題", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	_, err := uc.ListComments(context.Background(), usecase.ListCommentsInput{
		QuestionID: "q-1",
		CallerID:   "other-user",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// --- UpdateComment テスト ---

// TestCommentUseCase_UpdateComment_Success は正常系のコメント編集テストです。
func TestCommentUseCase_UpdateComment_Success(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	uRepo := newMockUserRepository()
	uRepo.addUser(testUserWithDisplayName("user-1", "alice", "Alice Smith"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", "user-1", "元のコメント"))

	uc := newCommentUseCase(cRepo, qRepo, uRepo)

	result, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		QuestionID: "q-1",
		CommentID:  "c-1",
		Body:       "編集後のコメント",
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("コメント編集に失敗しました: %v", err)
	}
	if result.Body != "編集後のコメント" {
		t.Errorf("Bodyが期待値と異なります: got %s, want 編集後のコメント", result.Body)
	}
}

// TestCommentUseCase_UpdateComment_NotOwner は投稿者以外の編集が拒否されることのテストです。
func TestCommentUseCase_UpdateComment_NotOwner(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", "user-1", "コメント"))

	uc := newCommentUseCase(cRepo, qRepo, newMockUserRepository())

	// user-2（非投稿者）は編集不可
	_, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		QuestionID: "q-1",
		CommentID:  "c-1",
		Body:       "編集コメント",
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCommentUseCase_UpdateComment_EmptyBody は本文が空の編集が拒否されることのテストです。
func TestCommentUseCase_UpdateComment_EmptyBody(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", "user-1", "コメント"))

	uc := newCommentUseCase(cRepo, qRepo, newMockUserRepository())

	_, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		QuestionID: "q-1",
		CommentID:  "c-1",
		Body:       "", // 空
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrCommentBodyEmpty) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrCommentBodyEmpty)
	}
}

// --- DeleteComment テスト ---

// TestCommentUseCase_DeleteComment_OwnerSuccess は投稿者本人による削除の正常系テストです。
func TestCommentUseCase_DeleteComment_OwnerSuccess(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", "user-1", "コメント"))

	uc := newCommentUseCase(cRepo, qRepo, newMockUserRepository())

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		CommentID:  "c-1",
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("投稿者本人による削除に失敗しました: %v", err)
	}
}

// TestCommentUseCase_DeleteComment_AdminSuccess は admin による他者コメントの削除テストです。
func TestCommentUseCase_DeleteComment_AdminSuccess(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", "user-1", "コメント"))

	uc := newCommentUseCase(cRepo, qRepo, newMockUserRepository())

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		CommentID:  "c-1",
		CallerID:   "admin-user",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Fatalf("admin による削除に失敗しました: %v", err)
	}
}

// TestCommentUseCase_DeleteComment_NotOwnerForbidden は投稿者以外（非 admin）の削除が拒否されることのテストです。
func TestCommentUseCase_DeleteComment_NotOwnerForbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", "user-1", "コメント"))

	uc := newCommentUseCase(cRepo, qRepo, newMockUserRepository())

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		CommentID:  "c-1",
		CallerID:   "user-2", // 非投稿者
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCommentUseCase_DeleteComment_CommentNotFound は存在しないコメントの削除テストです。
func TestCommentUseCase_DeleteComment_CommentNotFound(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		CommentID:  "no-such-comment",
		CallerID:   "user-1",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrCommentNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrCommentNotFound)
	}
}

// --- 閲覧権限チェックのテスト（受け入れ条件に対応） ---

// TestCommentUseCase_VisibilityCheck_PublishedAll_AllowsAllUsers は published+all 問題が全ユーザーに開放されることのテストです。
func TestCommentUseCase_VisibilityCheck_PublishedAll_AllowsAllUsers(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedAllQuestion("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	// 任意のユーザーがコメント可能
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "コメント",
		CallerID:   "random-user",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("published+all 問題へのコメントが拒否されました: %v", err)
	}
}

// TestCommentUseCase_VisibilityCheck_PublishedTeam_OwnerAllowed は published+team 問題の作成者がコメント可能であることのテストです。
func TestCommentUseCase_VisibilityCheck_PublishedTeam_OwnerAllowed(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedTeamQuestion("q-1", "owner-1", []string{"team-X"}))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	// 作成者本人は常にコメント可能
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		Body:       "オーナーコメント",
		CallerID:   "owner-1",
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("published+team 問題の作成者によるコメントが拒否されました: %v", err)
	}
}

// TestCommentUseCase_VisibilityCheck_DraftAndPrivate_OnlyOwner は draft/private 問題が作成者以外に拒否されることのテストです。
func TestCommentUseCase_VisibilityCheck_DraftAndPrivate_OnlyOwner(t *testing.T) {
	tests := []struct {
		name   string
		status domain.QuestionStatus
	}{
		{"draft問題", domain.QuestionStatusDraft},
		{"private問題", domain.QuestionStatusPrivate},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			qRepo := newMockQuestionRepository()
			q := testQuestion("q-1", tt.name, "owner-1")
			q.Status = tt.status
			qRepo.addQuestion(q)

			uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

			// 非作成者は拒否
			_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
				QuestionID: "q-1",
				Body:       "コメント",
				CallerID:   "other-user",
				CallerRole: domain.RoleUser,
			})

			if !errors.Is(err, domain.ErrPermissionDenied) {
				t.Errorf("[%s] エラーが期待値と異なります: got %v, want %v", tt.name, err, domain.ErrPermissionDenied)
			}

			// 作成者本人は許可
			_, err = uc.CreateComment(context.Background(), usecase.CreateCommentInput{
				QuestionID: "q-1",
				Body:       "コメント",
				CallerID:   "owner-1",
				CallerRole: domain.RoleUser,
			})

			if err != nil {
				t.Errorf("[%s] 作成者本人のコメントが拒否されました: %v", tt.name, err)
			}
		})
	}
}
