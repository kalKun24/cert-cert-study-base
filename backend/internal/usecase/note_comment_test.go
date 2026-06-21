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

// mockNoteCommentRepository は domain.NoteCommentRepository のモックです。
type mockNoteCommentRepository struct {
	comments  map[string]*domain.NoteComment // key: noteID + "/" + commentID
	saveErr   error
	deleteErr error
}

func newMockNoteCommentRepository() *mockNoteCommentRepository {
	return &mockNoteCommentRepository{
		comments: make(map[string]*domain.NoteComment),
	}
}

func noteCommentKey(noteID, commentID string) string {
	return noteID + "/" + commentID
}

func (m *mockNoteCommentRepository) addNoteComment(c *domain.NoteComment) {
	m.comments[noteCommentKey(c.NoteID, c.ID)] = c
}

func (m *mockNoteCommentRepository) FindByID(_ context.Context, teamID, noteID, commentID string) (*domain.NoteComment, error) {
	_ = teamID // テストではチームIDを区別しない
	if c, ok := m.comments[noteCommentKey(noteID, commentID)]; ok {
		return c, nil
	}
	return nil, domain.ErrNoteCommentNotFound
}

func (m *mockNoteCommentRepository) ListByNoteID(_ context.Context, teamID, noteID string) ([]*domain.NoteComment, error) {
	_ = teamID // テストではチームIDを区別しない
	var result []*domain.NoteComment
	for _, c := range m.comments {
		if c.NoteID == noteID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockNoteCommentRepository) Save(_ context.Context, teamID string, comment *domain.NoteComment) error {
	_ = teamID // テストではチームIDを区別しない
	if m.saveErr != nil {
		return m.saveErr
	}
	m.comments[noteCommentKey(comment.NoteID, comment.ID)] = comment
	return nil
}

func (m *mockNoteCommentRepository) Delete(_ context.Context, teamID, noteID, commentID string) error {
	_ = teamID // テストではチームIDを区別しない
	if m.deleteErr != nil {
		return m.deleteErr
	}
	key := noteCommentKey(noteID, commentID)
	if _, ok := m.comments[key]; !ok {
		return domain.ErrNoteCommentNotFound
	}
	delete(m.comments, key)
	return nil
}

// --- テストヘルパー ---

// testDraftNote はコメントテスト用の draft ノートを生成します（testTeamID に所属）。
func testDraftNote(id, createdBy string) *domain.Note {
	n := testNote(id, "下書きノート", createdBy)
	n.Status = domain.NoteStatusDraft
	return n
}

// testNoteComment はテスト用のノートコメントエンティティを生成します。
func testNoteComment(id, noteID, createdBy, body string) *domain.NoteComment {
	now := time.Now().UTC()
	return &domain.NoteComment{
		ID:        id,
		NoteID:    noteID,
		Body:      body,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// newNoteCommentUseCase はテスト用の NoteCommentUseCase を生成します。
// testCallerID が testTeamID のメンバーとして登録されたチームリポジトリを使います。
// userRepo は nil を渡し、テスト内では display_name の解決をスキップします。
func newNoteCommentUseCase(
	ncRepo *mockNoteCommentRepository,
	nRepo *mockNoteRepository,
) *usecase.NoteCommentUseCase {
	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID: testTeamID,
		UserID: testCallerID,
		Role:   domain.MemberRoleMember,
	})
	return usecase.NewNoteCommentUseCase(ncRepo, nRepo, tRepo, nil)
}

// newNoteCommentUseCaseWithTeam は自由にチームリポジトリを指定できる NoteCommentUseCase を生成します。
// userRepo は nil を渡し、テスト内では display_name の解決をスキップします。
func newNoteCommentUseCaseWithTeam(
	ncRepo *mockNoteCommentRepository,
	nRepo *mockNoteRepository,
	tRepo *mockTeamRepository,
) *usecase.NoteCommentUseCase {
	return usecase.NewNoteCommentUseCase(ncRepo, nRepo, tRepo, nil)
}

// --- CreateNoteComment テスト ---

// TestCreateNoteComment_Success は正常系のノートコメント投稿テストです。
func TestCreateNoteComment_Success(t *testing.T) {
	nRepo := newMockNoteRepository()
	nRepo.addNote(testPublishedNote("note-1", "owner-1"))

	ncRepo := newMockNoteCommentRepository()
	uc := newNoteCommentUseCase(ncRepo, nRepo)

	result, err := uc.CreateNoteComment(context.Background(), usecase.CreateNoteCommentInput{
		TeamID:     testTeamID,
		NoteID:     "note-1",
		Body:       "テストコメント",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("コメント投稿に失敗しました: %v", err)
	}
	if result.Body != "テストコメント" {
		t.Errorf("Bodyが期待値と異なります: got %s, want テストコメント", result.Body)
	}
	if result.CreatedBy != testCallerID {
		t.Errorf("CreatedByが期待値と異なります: got %s, want %s", result.CreatedBy, testCallerID)
	}
	if result.ID == "" {
		t.Error("IDが生成されていません")
	}
	if result.NoteID != "note-1" {
		t.Errorf("NoteIDが期待値と異なります: got %s, want note-1", result.NoteID)
	}
}

// TestCreateNoteComment_NotMember はチーム非メンバーのコメント投稿が拒否されることのテストです。
func TestCreateNoteComment_NotMember(t *testing.T) {
	nRepo := newMockNoteRepository()
	nRepo.addNote(testPublishedNote("note-1", "owner-1"))

	ncRepo := newMockNoteCommentRepository()
	uc := newNoteCommentUseCase(ncRepo, nRepo)

	// "non-member-user" はチームメンバーではない
	_, err := uc.CreateNoteComment(context.Background(), usecase.CreateNoteCommentInput{
		TeamID:     testTeamID,
		NoteID:     "note-1",
		Body:       "コメント",
		CallerID:   "non-member-user",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCreateNoteComment_EmptyBody はコメント本文が空の場合のバリデーションテストです。
func TestCreateNoteComment_EmptyBody(t *testing.T) {
	nRepo := newMockNoteRepository()
	nRepo.addNote(testPublishedNote("note-1", "owner-1"))

	ncRepo := newMockNoteCommentRepository()
	uc := newNoteCommentUseCase(ncRepo, nRepo)

	_, err := uc.CreateNoteComment(context.Background(), usecase.CreateNoteCommentInput{
		TeamID:     testTeamID,
		NoteID:     "note-1",
		Body:       "", // 空
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrCommentBodyEmpty) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrCommentBodyEmpty)
	}
}

// TestCreateNoteComment_DraftNoteNotVisible は draft ノートへのコメントが作成者以外に拒否されることのテストです。
func TestCreateNoteComment_DraftNoteNotVisible(t *testing.T) {
	nRepo := newMockNoteRepository()
	// draft ノート: "owner-1" が作成
	nRepo.addNote(testDraftNote("note-1", "owner-1"))

	tRepo := newMockTeamRepository()
	// testCallerID と "owner-1" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "owner-1", Role: domain.MemberRoleMember})

	ncRepo := newMockNoteCommentRepository()
	uc := newNoteCommentUseCaseWithTeam(ncRepo, nRepo, tRepo)

	// testCallerID（非作成者、チームメンバー）はコメント不可
	_, err := uc.CreateNoteComment(context.Background(), usecase.CreateNoteCommentInput{
		TeamID:     testTeamID,
		NoteID:     "note-1",
		Body:       "コメント",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestCreateNoteComment_DraftNoteOwnerAllowed は draft ノートへのコメントが作成者本人に許可されることのテストです。
func TestCreateNoteComment_DraftNoteOwnerAllowed(t *testing.T) {
	nRepo := newMockNoteRepository()
	// draft ノート: testCallerID が作成者
	nRepo.addNote(testDraftNote("note-1", testCallerID))

	ncRepo := newMockNoteCommentRepository()
	uc := newNoteCommentUseCase(ncRepo, nRepo)

	// testCallerID（作成者本人）はコメント可
	_, err := uc.CreateNoteComment(context.Background(), usecase.CreateNoteCommentInput{
		TeamID:     testTeamID,
		NoteID:     "note-1",
		Body:       "コメント",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("作成者本人のコメント投稿に失敗しました: %v", err)
	}
}

// --- ListNoteComments テスト ---

// TestListNoteComments_Sorted はコメント一覧が投稿日時の昇順で返ることのテストです。
func TestListNoteComments_Sorted(t *testing.T) {
	nRepo := newMockNoteRepository()
	nRepo.addNote(testPublishedNote("note-1", "owner-1"))

	ncRepo := newMockNoteCommentRepository()
	now := time.Now().UTC()
	// 投稿順を逆にして挿入し、ソートが正しく機能することを確認
	ncRepo.addNoteComment(&domain.NoteComment{
		ID: "nc-2", NoteID: "note-1", Body: "コメント2", CreatedBy: "user-2",
		CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute),
	})
	ncRepo.addNoteComment(&domain.NoteComment{
		ID: "nc-1", NoteID: "note-1", Body: "コメント1", CreatedBy: testCallerID,
		CreatedAt: now, UpdatedAt: now,
	})

	uc := newNoteCommentUseCase(ncRepo, nRepo)

	result, err := uc.ListNoteComments(context.Background(), testTeamID, "note-1", testCallerID, domain.RoleUser)

	if err != nil {
		t.Fatalf("コメント一覧取得に失敗しました: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("取得件数が期待値と異なります: got %d, want 2", len(result))
	}
	// 昇順ソートの確認
	if result[0].ID != "nc-1" {
		t.Errorf("1件目のIDが期待値と異なります: got %s, want nc-1", result[0].ID)
	}
	if result[1].ID != "nc-2" {
		t.Errorf("2件目のIDが期待値と異なります: got %s, want nc-2", result[1].ID)
	}
}

// --- UpdateNoteComment テスト ---

// TestUpdateNoteComment_OnlyAuthorCanEdit は投稿者本人以外の編集が拒否されることのテストです。
func TestUpdateNoteComment_OnlyAuthorCanEdit(t *testing.T) {
	nRepo := newMockNoteRepository()
	nRepo.addNote(testPublishedNote("note-1", "owner-1"))

	ncRepo := newMockNoteCommentRepository()
	ncRepo.addNoteComment(testNoteComment("nc-1", "note-1", testCallerID, "コメント"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newNoteCommentUseCaseWithTeam(ncRepo, nRepo, tRepo)

	// "user-2"（チームメンバーだが非投稿者）は編集不可
	_, err := uc.UpdateNoteComment(context.Background(), usecase.UpdateNoteCommentInput{
		TeamID:     testTeamID,
		NoteID:     "note-1",
		CommentID:  "nc-1",
		Body:       "編集後のコメント",
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}

	// 投稿者本人は編集可能
	result, err := uc.UpdateNoteComment(context.Background(), usecase.UpdateNoteCommentInput{
		TeamID:     testTeamID,
		NoteID:     "note-1",
		CommentID:  "nc-1",
		Body:       "編集後のコメント",
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("投稿者本人の編集に失敗しました: %v", err)
	}
	if result.Body != "編集後のコメント" {
		t.Errorf("Bodyが期待値と異なります: got %s, want 編集後のコメント", result.Body)
	}
}

// --- DeleteNoteComment テスト ---

// TestDeleteNoteComment_AdminCanDelete は admin が投稿者以外のコメントも削除できることのテストです。
func TestDeleteNoteComment_AdminCanDelete(t *testing.T) {
	nRepo := newMockNoteRepository()
	nRepo.addNote(testPublishedNote("note-1", "owner-1"))

	ncRepo := newMockNoteCommentRepository()
	ncRepo.addNoteComment(testNoteComment("nc-1", "note-1", testCallerID, "コメント"))

	// admin はどのチームにもメンバー登録されていない
	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})

	uc := newNoteCommentUseCaseWithTeam(ncRepo, nRepo, tRepo)

	err := uc.DeleteNoteComment(context.Background(), testTeamID, "note-1", "nc-1", "admin-user", domain.RoleAdmin)

	if err != nil {
		t.Errorf("admin は他者のコメントを削除できるべきです: %v", err)
	}
}

// TestDeleteNoteComment_AuthorCanDelete は投稿者本人が自身のコメントを削除できることのテストです。
func TestDeleteNoteComment_AuthorCanDelete(t *testing.T) {
	nRepo := newMockNoteRepository()
	nRepo.addNote(testPublishedNote("note-1", "owner-1"))

	ncRepo := newMockNoteCommentRepository()
	ncRepo.addNoteComment(testNoteComment("nc-1", "note-1", testCallerID, "コメント"))

	uc := newNoteCommentUseCase(ncRepo, nRepo)

	err := uc.DeleteNoteComment(context.Background(), testTeamID, "note-1", "nc-1", testCallerID, domain.RoleUser)

	if err != nil {
		t.Fatalf("投稿者本人による削除に失敗しました: %v", err)
	}
}

// TestDeleteNoteComment_NotOwnerForbidden は投稿者以外（非 admin）の削除が拒否されることのテストです。
func TestDeleteNoteComment_NotOwnerForbidden(t *testing.T) {
	nRepo := newMockNoteRepository()
	nRepo.addNote(testPublishedNote("note-1", "owner-1"))

	ncRepo := newMockNoteCommentRepository()
	ncRepo.addNoteComment(testNoteComment("nc-1", "note-1", testCallerID, "コメント"))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newNoteCommentUseCaseWithTeam(ncRepo, nRepo, tRepo)

	err := uc.DeleteNoteComment(context.Background(), testTeamID, "note-1", "nc-1", "user-2", domain.RoleUser)

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}
