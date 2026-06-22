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

func (m *mockCommentRepository) FindByID(_ context.Context, teamID, questionID, commentID string) (*domain.Comment, error) {
	_ = teamID // テストではチームIDを区別しない（キーは questionID/commentID のみ）
	if c, ok := m.comments[commentKey(questionID, commentID)]; ok {
		return c, nil
	}
	return nil, domain.ErrCommentNotFound
}

func (m *mockCommentRepository) ListByQuestionID(_ context.Context, teamID, questionID string) ([]*domain.Comment, error) {
	_ = teamID // テストではチームIDを区別しない
	var result []*domain.Comment
	for _, c := range m.comments {
		if c.QuestionID == questionID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCommentRepository) Save(_ context.Context, teamID string, comment *domain.Comment) error {
	_ = teamID // テストではチームIDを区別しない
	if m.saveErr != nil {
		return m.saveErr
	}
	m.comments[commentKey(comment.QuestionID, comment.ID)] = comment
	return nil
}

func (m *mockCommentRepository) Delete(_ context.Context, teamID, questionID, commentID string) error {
	_ = teamID // テストではチームIDを区別しない
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
// testCallerID ("user-1") が testTeamID ("team-1") のメンバーとして登録されたチームリポジトリを使います。
func newCommentUseCase(
	cRepo *mockCommentRepository,
	qRepo *mockQuestionRepository,
	uRepo *mockUserRepository,
) *usecase.CommentUseCase {
	tRepo := newMockTeamRepository()
	// testCallerID を testTeamID のメンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID: testTeamID,
		UserID: testCallerID,
		Role:   domain.MemberRoleMember,
	})
	return usecase.NewCommentUseCase(cRepo, qRepo, uRepo, tRepo)
}

// newCommentUseCaseWithTeam は自由にチームリポジトリを指定できる CommentUseCase を生成します。
func newCommentUseCaseWithTeam(
	cRepo *mockCommentRepository,
	qRepo *mockQuestionRepository,
	uRepo *mockUserRepository,
	tRepo *mockTeamRepository,
) *usecase.CommentUseCase {
	return usecase.NewCommentUseCase(cRepo, qRepo, uRepo, tRepo)
}

// testPublishedQuestion_comment はコメントテスト用の published 問題を生成します（testTeamID に所属）。
func testPublishedQuestion_comment(id, createdBy string) *domain.Question {
	q := testQuestion(id, "公開問題", createdBy)
	q.Status = domain.QuestionStatusPublished
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
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uRepo := newMockUserRepository()
	uRepo.addUser(testUserWithDisplayName(testCallerID, "alice", "Alice Smith"))

	cRepo := newMockCommentRepository()
	uc := newCommentUseCase(cRepo, qRepo, uRepo)

	result, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		Body:       "テストコメント",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("コメント投稿に失敗しました: %v", err)
	}
	if result.Body != "テストコメント" {
		t.Errorf("BodyがBody期待値と異なります: got %s, want テストコメント", result.Body)
	}
	if result.CreatedBy != testCallerID {
		t.Errorf("CreatedByが期待値と異なります: got %s, want %s", result.CreatedBy, testCallerID)
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
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		Body:       "", // 空
		CallerID:   testCallerID,
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
		TeamID:     testTeamID,
		Body:       "コメント",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrQuestionNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrQuestionNotFound)
	}
}

// TestCommentUseCase_CreateComment_DraftQuestionForbidden は draft 問題へのコメント投稿が作成者以外に拒否されることのテストです。
func TestCommentUseCase_CreateComment_DraftQuestionForbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	// draft 問題: "owner-1" が作成
	q := testQuestion("q-1", "draft問題", "owner-1")
	qRepo.addQuestion(q)

	tRepo := newMockTeamRepository()
	// testCallerID と "owner-1" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "owner-1", Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(newMockCommentRepository(), qRepo, newMockUserRepository(), tRepo)

	// testCallerID（非作成者、チームメンバー）はコメント不可
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		Body:       "コメント",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCommentUseCase_CreateComment_DraftQuestionOwnerAllowed は draft 問題へのコメント投稿が作成者本人に許可されることのテストです。
func TestCommentUseCase_CreateComment_DraftQuestionOwnerAllowed(t *testing.T) {
	qRepo := newMockQuestionRepository()
	q := testQuestion("q-1", "draft問題", testCallerID) // testCallerID が作成者
	qRepo.addQuestion(q)

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	// testCallerID（作成者本人、チームメンバー）はコメント可
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		Body:       "コメント",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("作成者本人のコメント投稿に失敗しました: %v", err)
	}
}

// TestCommentUseCase_CreateComment_PublishedAllowedForMember は published 問題へのチームメンバーのコメント投稿が許可されるテストです。
func TestCommentUseCase_CreateComment_PublishedAllowedForMember(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		Body:       "コメント",
		CallerID:   testCallerID, // チームメンバー
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("published 問題へのメンバーコメント投稿に失敗しました: %v", err)
	}
}

// TestCommentUseCase_CreateComment_NonMemberForbidden はチーム非メンバーのコメント投稿が拒否されるテストです。
func TestCommentUseCase_CreateComment_NonMemberForbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	// "non-member-user" はチームメンバーではない → 拒否
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		Body:       "コメント",
		CallerID:   "non-member-user",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCommentUseCase_CreateComment_AdminNonMemberAllowed は admin がチーム非メンバーでもコメント投稿できるテストです。
func TestCommentUseCase_CreateComment_AdminNonMemberAllowed(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	// admin はどのチームにもメンバー登録されていない
	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(newMockCommentRepository(), qRepo, newMockUserRepository(), tRepo)

	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		Body:       "adminコメント",
		CallerID:   "admin-user",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Errorf("admin はチーム非メンバーでもコメント投稿できるべきです: %v", err)
	}
}

// TestCommentUseCase_CreateComment_AdminCanCommentOnDraft は admin が他者の draft 問題にもコメント投稿できるテストです。
func TestCommentUseCase_CreateComment_AdminCanCommentOnDraft(t *testing.T) {
	qRepo := newMockQuestionRepository()
	// draft 問題: "owner-1" が作成
	qRepo.addQuestion(testQuestion("q-1", "draft問題", "owner-1"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "owner-1", Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(newMockCommentRepository(), qRepo, newMockUserRepository(), tRepo)

	// admin は draft の作成者でなくてもコメント可能
	_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		Body:       "adminコメント",
		CallerID:   "admin-user",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Errorf("admin は他者の draft 問題にもコメントできるべきです: %v", err)
	}
}

// --- ListComments テスト ---

// TestCommentUseCase_ListComments_Success は正常系のコメント一覧取得テストです。
func TestCommentUseCase_ListComments_Success(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uRepo := newMockUserRepository()
	uRepo.addUser(testUserWithDisplayName(testCallerID, "alice", "Alice Smith"))
	uRepo.addUser(testUserWithDisplayName("user-2", "bob", "Bob Jones"))

	cRepo := newMockCommentRepository()
	now := time.Now().UTC()
	// 投稿順を逆にして挿入し、ソートが正しく機能することを確認
	cRepo.addComment(&domain.Comment{
		ID: "c-2", QuestionID: "q-1", Body: "コメント2", CreatedBy: "user-2",
		CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute),
	})
	cRepo.addComment(&domain.Comment{
		ID: "c-1", QuestionID: "q-1", Body: "コメント1", CreatedBy: testCallerID,
		CreatedAt: now, UpdatedAt: now,
	})

	uc := newCommentUseCase(cRepo, qRepo, uRepo)

	result, err := uc.ListComments(context.Background(), usecase.ListCommentsInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CallerID:   testCallerID,
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
}

// TestCommentUseCase_ListComments_Forbidden は閲覧権限がないユーザーの一覧取得が拒否されることのテストです。
func TestCommentUseCase_ListComments_Forbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	// draft 問題: "owner-1" が作成
	q := testQuestion("q-1", "draft問題", "owner-1")
	qRepo.addQuestion(q)

	tRepo := newMockTeamRepository()
	// testCallerID と "other-user" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "other-user", Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(newMockCommentRepository(), qRepo, newMockUserRepository(), tRepo)

	// "other-user"（チームメンバーだが draft 作成者ではない）は閲覧不可
	_, err := uc.ListComments(context.Background(), usecase.ListCommentsInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CallerID:   "other-user",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCommentUseCase_ListComments_NonMemberForbidden はチーム非メンバーが一覧取得できないテストです。
func TestCommentUseCase_ListComments_NonMemberForbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	_, err := uc.ListComments(context.Background(), usecase.ListCommentsInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CallerID:   "non-member-user",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("チーム非メンバーは拒否されるべきです: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// --- UpdateComment テスト ---

// TestCommentUseCase_UpdateComment_Success は正常系のコメント編集テストです。
func TestCommentUseCase_UpdateComment_Success(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uRepo := newMockUserRepository()
	uRepo.addUser(testUserWithDisplayName(testCallerID, "alice", "Alice Smith"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "元のコメント"))

	uc := newCommentUseCase(cRepo, qRepo, uRepo)

	result, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		Body:       "編集後のコメント",
		CallerID:   testCallerID,
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
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(cRepo, qRepo, newMockUserRepository(), tRepo)

	// "user-2"（チームメンバーだが非投稿者）は編集不可
	_, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
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
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	uc := newCommentUseCase(cRepo, qRepo, newMockUserRepository())

	_, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		Body:       "", // 空
		CallerID:   testCallerID,
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
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	uc := newCommentUseCase(cRepo, qRepo, newMockUserRepository())

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("投稿者本人による削除に失敗しました: %v", err)
	}
}

// TestCommentUseCase_DeleteComment_AdminMemberSuccess は admin かつチームメンバーによる他者コメントの削除テストです。
func TestCommentUseCase_DeleteComment_AdminMemberSuccess(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "admin-user", Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(cRepo, qRepo, newMockUserRepository(), tRepo)

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		CallerID:   "admin-user",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Fatalf("admin（チームメンバー）による削除に失敗しました: %v", err)
	}
}

// TestCommentUseCase_DeleteComment_AdminNonMemberAllowed は admin がチーム非メンバーでも他者のコメントを削除できるテストです。
func TestCommentUseCase_DeleteComment_AdminNonMemberAllowed(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	// admin-user はチームメンバーではない
	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(cRepo, qRepo, newMockUserRepository(), tRepo)

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		CallerID:   "admin-user",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Errorf("admin はチーム非メンバーでも削除できるべきです: %v", err)
	}
}

// TestCommentUseCase_DeleteComment_NotOwnerForbidden は投稿者以外（非 admin）の削除が拒否されることのテストです。
func TestCommentUseCase_DeleteComment_NotOwnerForbidden(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(cRepo, qRepo, newMockUserRepository(), tRepo)

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		CallerID:   "user-2", // 非投稿者、チームメンバーだが admin でもない
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCommentUseCase_DeleteComment_CommentNotFound は存在しないコメントの削除テストです。
func TestCommentUseCase_DeleteComment_CommentNotFound(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uc := newCommentUseCase(newMockCommentRepository(), qRepo, newMockUserRepository())

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "no-such-comment",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrCommentNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrCommentNotFound)
	}
}

// TestCommentUseCase_UpdateComment_AdminCanUpdateOthersComment は
// admin が他人のコメントを更新できることのテストです。
func TestCommentUseCase_UpdateComment_AdminCanUpdateOthersComment(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	uRepo := newMockUserRepository()

	cRepo := newMockCommentRepository()
	// testCallerID が投稿したコメント
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "元のコメント"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	// admin-user はチームメンバーではない（admin はメンバーチェックをスキップ）

	uc := newCommentUseCaseWithTeam(cRepo, qRepo, uRepo, tRepo)

	// admin-user（作成者とは別の admin）が編集
	result, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		Body:       "adminが編集したコメント",
		CallerID:   "admin-user",
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Fatalf("admin は他人のコメントを更新できるべきです: %v", err)
	}
	if result.Body != "adminが編集したコメント" {
		t.Errorf("Bodyが期待値と異なります: got %s, want adminが編集したコメント", result.Body)
	}
}

// TestCommentUseCase_UpdateComment_UserCannotUpdateOthersComment は
// 一般 user が他人のコメントを更新しようとすると ErrPermissionDenied になることのテストです。
func TestCommentUseCase_UpdateComment_UserCannotUpdateOthersComment(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(cRepo, qRepo, newMockUserRepository(), tRepo)

	_, err := uc.UpdateComment(context.Background(), usecase.UpdateCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		Body:       "不正な編集",
		CallerID:   "user-2", // 非投稿者・非 admin
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("一般 user が他人のコメントを更新しようとすると ErrPermissionDenied になるべきです: got %v", err)
	}
}

// TestCommentUseCase_DeleteComment_AdminCanDeleteOthersComment は
// admin が他人のコメントを削除できることのテストです（チーム非メンバー）。
func TestCommentUseCase_DeleteComment_AdminCanDeleteOthersComment(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	// admin-user はチームメンバーではない
	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(cRepo, qRepo, newMockUserRepository(), tRepo)

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		CallerID:   "admin-user", // 非チームメンバーの admin
		CallerRole: domain.RoleAdmin,
	})

	if err != nil {
		t.Errorf("admin は他人のコメントを削除できるべきです: %v", err)
	}
}

// TestCommentUseCase_DeleteComment_UserCannotDeleteOthersComment は
// 一般 user が他人のコメントを削除しようとすると ErrPermissionDenied になることのテストです。
func TestCommentUseCase_DeleteComment_UserCannotDeleteOthersComment(t *testing.T) {
	qRepo := newMockQuestionRepository()
	qRepo.addQuestion(testPublishedQuestion_comment("q-1", "owner-1"))

	cRepo := newMockCommentRepository()
	cRepo.addComment(testComment("c-1", "q-1", testCallerID, "コメント"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newCommentUseCaseWithTeam(cRepo, qRepo, newMockUserRepository(), tRepo)

	err := uc.DeleteComment(context.Background(), usecase.DeleteCommentInput{
		QuestionID: "q-1",
		TeamID:     testTeamID,
		CommentID:  "c-1",
		CallerID:   "user-2", // 非投稿者・非 admin
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("一般 user が他人のコメントを削除しようとすると ErrPermissionDenied になるべきです: got %v", err)
	}
}

// --- 可視性チェックのテスト（チームスコープ化後） ---

// TestCommentUseCase_VisibilityCheck_DraftAndPrivate_OnlyOwnerCanComment は
// draft/private 問題へのコメントが作成者のみに許可されるテストです（チームメンバー前提）。
func TestCommentUseCase_VisibilityCheck_DraftAndPrivate_OnlyOwnerCanComment(t *testing.T) {
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
			q := testQuestion("q-1", tt.name, testCallerID) // testCallerID が作成者
			q.Status = tt.status
			qRepo.addQuestion(q)

			tRepo := newMockTeamRepository()
			_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
			_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "other-member", Role: domain.MemberRoleMember})

			uc := newCommentUseCaseWithTeam(newMockCommentRepository(), qRepo, newMockUserRepository(), tRepo)

			// 他のチームメンバーは拒否
			_, err := uc.CreateComment(context.Background(), usecase.CreateCommentInput{
				QuestionID: "q-1",
				TeamID:     testTeamID,
				Body:       "コメント",
				CallerID:   "other-member",
				CallerRole: domain.RoleUser,
			})

			if tt.status == domain.QuestionStatusDraft {
				if !errors.Is(err, domain.ErrPermissionDenied) {
					t.Errorf("[%s] 非作成者は拒否されるべきです: got %v, want %v", tt.name, err, domain.ErrPermissionDenied)
				}
			} else {
				// private は published と同様にメンバー全員に見える
				if err != nil {
					t.Errorf("[%s] private 問題はメンバーに許可されるべきです: got %v", tt.name, err)
				}
			}

			// 作成者本人は許可
			_, err = uc.CreateComment(context.Background(), usecase.CreateCommentInput{
				QuestionID: "q-1",
				TeamID:     testTeamID,
				Body:       "コメント",
				CallerID:   testCallerID,
				CallerRole: domain.RoleUser,
			})

			if err != nil {
				t.Errorf("[%s] 作成者本人のコメントが拒否されました: %v", tt.name, err)
			}
		})
	}
}
