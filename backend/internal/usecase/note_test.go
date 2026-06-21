package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// --- モック定義 ---

// mockNoteRepository は domain.NoteRepository のモックです。
type mockNoteRepository struct {
	notes     map[string]*domain.Note
	saveErr   error
	deleteErr error
}

func newMockNoteRepository() *mockNoteRepository {
	return &mockNoteRepository{
		notes: make(map[string]*domain.Note),
	}
}

// addNote はテスト用のノートをモックに追加します。
func (m *mockNoteRepository) addNote(n *domain.Note) {
	m.notes[n.ID] = n
}

func (m *mockNoteRepository) FindByID(_ context.Context, teamID, id string) (*domain.Note, error) {
	if n, ok := m.notes[id]; ok {
		if n.TeamID == teamID {
			return n, nil
		}
	}
	return nil, domain.ErrNoteNotFound
}

func (m *mockNoteRepository) ListByTeam(_ context.Context, teamID string) ([]*domain.Note, error) {
	notes := make([]*domain.Note, 0)
	for _, n := range m.notes {
		if n.TeamID == teamID {
			notes = append(notes, n)
		}
	}
	return notes, nil
}

func (m *mockNoteRepository) SearchByTeam(_ context.Context, teamID string, filter domain.NoteSearchFilter) ([]*domain.Note, error) {
	var result []*domain.Note
	for _, n := range m.notes {
		if n.TeamID != teamID {
			continue
		}
		// タグANDフィルタリング
		if len(filter.TagIDs) > 0 {
			tagSet := make(map[string]struct{}, len(n.Tags))
			for _, t := range n.Tags {
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
			if !noteContains(n.Title, kw) && !noteContains(n.Body, kw) &&
				!noteContains(n.DiscussionPoints, kw) && !noteContains(n.Memo, kw) {
				continue
			}
		}
		result = append(result, n)
	}
	return result, nil
}

func (m *mockNoteRepository) Save(_ context.Context, note *domain.Note) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.notes[note.ID] = note
	return nil
}

func (m *mockNoteRepository) Delete(_ context.Context, teamID, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	n, ok := m.notes[id]
	if !ok || n.TeamID != teamID {
		return domain.ErrNoteNotFound
	}
	delete(m.notes, id)
	return nil
}

func (m *mockNoteRepository) FindByTagID(_ context.Context, teamID, tagID string) ([]*domain.Note, error) {
	var result []*domain.Note
	for _, n := range m.notes {
		if n.TeamID != teamID {
			continue
		}
		for _, tid := range n.Tags {
			if tid == tagID {
				result = append(result, n)
				break
			}
		}
	}
	return result, nil
}

// noteContains は s が substr を含むかどうかを返すヘルパーです。
func noteContains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- テストヘルパー ---

// testNote はテスト用のノートエンティティを生成します（testTeamID に所属）。
func testNote(id, title, createdBy string) *domain.Note {
	return &domain.Note{
		ID:               id,
		TeamID:           testTeamID,
		Title:            title,
		Body:             "## 本文\nテスト本文",
		DiscussionPoints: "## 議論点\nテスト議論点",
		Memo:             "## メモ\nテストメモ",
		Tags:             []string{"テスト"},
		Status:           domain.NoteStatusDraft,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}
}

// testPublishedNote は公開済みノートエンティティを生成します（testTeamID に所属）。
func testPublishedNote(id, createdBy string) *domain.Note {
	n := testNote(id, "公開ノート", createdBy)
	n.Status = domain.NoteStatusPublished
	return n
}

// newNoteUseCase はテスト用の NoteUseCase を生成します。
// testCallerID が testTeamID のメンバーとして登録された teamRepo を使います。
func newNoteUseCase(nRepo *mockNoteRepository) *usecase.NoteUseCase {
	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID: testTeamID,
		UserID: testCallerID,
		Role:   domain.MemberRoleMember,
	})
	return usecase.NewNoteUseCase(nRepo, tRepo)
}

// newNoteUseCaseWithTeam はチームとメンバーを自由に設定できる NoteUseCase を生成します。
func newNoteUseCaseWithTeam(nRepo *mockNoteRepository, tRepo *mockTeamRepository) *usecase.NoteUseCase {
	return usecase.NewNoteUseCase(nRepo, tRepo)
}

// --- CreateNote テスト ---

// TestCreateNote_Success は正常系のノート作成テストです。
func TestCreateNote_Success(t *testing.T) {
	repo := newMockNoteRepository()
	uc := newNoteUseCase(repo)

	n, err := uc.CreateNote(context.Background(), testTeamID, usecase.CreateNoteInput{
		CallerID:         testCallerID,
		Title:            "テストノート",
		Body:             "## 本文\n本文",
		DiscussionPoints: "## 議論点\n議論点",
		Memo:             "## メモ\nメモ",
		Tags:             []string{"CISSP"},
	})

	if err != nil {
		t.Fatalf("ノート作成に失敗しました: %v", err)
	}
	if n.Title != "テストノート" {
		t.Errorf("タイトルが期待値と異なります: got %s, want テストノート", n.Title)
	}
	if n.TeamID != testTeamID {
		t.Errorf("teamIDが期待値と異なります: got %s, want %s", n.TeamID, testTeamID)
	}
	if n.Status != domain.NoteStatusDraft {
		t.Errorf("statusが期待値と異なります: got %s, want draft", n.Status)
	}
	if n.CreatedBy != testCallerID {
		t.Errorf("created_byが期待値と異なります: got %s, want %s", n.CreatedBy, testCallerID)
	}
	if n.ID == "" {
		t.Error("IDが生成されていません")
	}
}

// TestCreateNote_NotMember はチーム非メンバーが作成できないテストです。
func TestCreateNote_NotMember(t *testing.T) {
	repo := newMockNoteRepository()
	uc := newNoteUseCase(repo) // testCallerID のみメンバー登録済み

	_, err := uc.CreateNote(context.Background(), testTeamID, usecase.CreateNoteInput{
		CallerID: "user-not-member",
		Title:    "テストノート",
	})

	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("チーム非メンバーは 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestCreateNote_EmptyTitle はタイトルが空の場合のバリデーションテストです。
func TestCreateNote_EmptyTitle(t *testing.T) {
	repo := newMockNoteRepository()
	uc := newNoteUseCase(repo)

	_, err := uc.CreateNote(context.Background(), testTeamID, usecase.CreateNoteInput{
		CallerID: testCallerID,
		Title:    "", // 空タイトル
	})

	if err == nil {
		t.Fatal("タイトルが空の場合はエラーが返されるべきです")
	}
}

// --- GetNote テスト ---

// TestGetNote_Success は正常系のノート詳細取得テストです。
func TestGetNote_Success(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	uc := newNoteUseCase(repo)

	n, err := uc.GetNote(context.Background(), "n-1", testTeamID, usecase.GetNoteInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Fatalf("ノート取得に失敗しました: %v", err)
	}
	if n.ID != "n-1" {
		t.Errorf("IDが期待値と異なります: got %s, want n-1", n.ID)
	}
}

// TestGetNote_DraftVisibility は draft ノートが作成者本人のみ閲覧可能なテストです。
func TestGetNote_DraftVisibility(t *testing.T) {
	repo := newMockNoteRepository()
	// testCallerID が作成した draft ノート
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID と "user-2" を両方メンバーとして登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	// user-2 には見えない（draft は作成者本人のみ）
	_, err := uc.GetNote(context.Background(), "n-1", testTeamID, usecase.GetNoteInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
	})
	if !errors.Is(err, domain.ErrNoteNotFound) {
		t.Errorf("draft ノートは作成者以外には404になるべきです: got %v, want %v", err, domain.ErrNoteNotFound)
	}

	// testCallerID には見える
	n, err := uc.GetNote(context.Background(), "n-1", testTeamID, usecase.GetNoteInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
	})
	if err != nil {
		t.Errorf("作成者本人は draft ノートを閲覧できるべきです: %v", err)
	}
	if n == nil || n.ID != "n-1" {
		t.Error("取得したノートが期待値と異なります")
	}
}

// --- UpdateNote テスト ---

// TestUpdateNote_OwnerCanEdit はチームオーナーが他人のノートを編集できるテストです。
func TestUpdateNote_OwnerCanEdit(t *testing.T) {
	repo := newMockNoteRepository()
	// testCallerID が作成したノート
	repo.addNote(testNote("n-1", "元のタイトル", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID を member、"team-owner-1" を owner として登録
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "team-owner-1", Role: domain.MemberRoleOwner})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	newTitle := "オーナーが更新"
	n, err := uc.UpdateNote(context.Background(), "n-1", testTeamID, usecase.UpdateNoteInput{
		CallerID:   "team-owner-1", // チームオーナー（作成者ではない）
		CallerRole: domain.RoleUser,
		Title:      &newTitle,
	})

	if err != nil {
		t.Fatalf("チームオーナーはノートを編集できるべきです: %v", err)
	}
	if n.Title != "オーナーが更新" {
		t.Errorf("タイトルが期待値と異なります: got %s, want オーナーが更新", n.Title)
	}
}

// TestUpdateNote_NotOwnerCannotEdit は非オーナー・非作成者は編集不可のテストです。
func TestUpdateNote_NotOwnerCannotEdit(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	tRepo := newMockTeamRepository()
	// testCallerID と "user-2" を両方 member として登録（owner なし）
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	newTitle := "不正な更新"
	_, err := uc.UpdateNote(context.Background(), "n-1", testTeamID, usecase.UpdateNoteInput{
		CallerID:   "user-2", // チームメンバーだが作成者でも owner でも admin でもない
		CallerRole: domain.RoleUser,
		Title:      &newTitle,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// --- DeleteNote テスト ---

// TestDeleteNote_Success は作成者本人によるノート削除テストです。
func TestDeleteNote_Success(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	uc := newNoteUseCase(repo)

	err := uc.DeleteNote(context.Background(), "n-1", testTeamID, testCallerID, domain.RoleUser)
	if err != nil {
		t.Fatalf("ノート削除に失敗しました: %v", err)
	}

	// 削除後に取得しようとすると NotFound になること
	_, err = repo.FindByID(context.Background(), testTeamID, "n-1")
	if !errors.Is(err, domain.ErrNoteNotFound) {
		t.Error("削除後もノートが存在しています")
	}
}

// TestDeleteNote_OwnerCanDelete はチームオーナーが他人のノートを削除できるテストです。
func TestDeleteNote_OwnerCanDelete(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "team-owner-1", Role: domain.MemberRoleOwner})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	err := uc.DeleteNote(context.Background(), "n-1", testTeamID, "team-owner-1", domain.RoleUser)
	if err != nil {
		t.Fatalf("チームオーナーはノートを削除できるべきです: %v", err)
	}
}

// TestDeleteNote_PermissionDenied は権限のないユーザーによる削除テストです。
func TestDeleteNote_PermissionDenied(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	err := uc.DeleteNote(context.Background(), "n-1", testTeamID, "user-2", domain.RoleUser)
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// --- UpdateNoteVisibility テスト ---

// TestUpdateNoteVisibility_Success は作成者による公開設定変更テストです。
func TestUpdateNoteVisibility_Success(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	uc := newNoteUseCase(repo)

	n, err := uc.UpdateNoteVisibility(context.Background(), "n-1", testTeamID, usecase.UpdateNoteVisibilityInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Status:     domain.NoteStatusPublished,
	})

	if err != nil {
		t.Fatalf("公開設定変更に失敗しました: %v", err)
	}
	if n.Status != domain.NoteStatusPublished {
		t.Errorf("statusが期待値と異なります: got %s, want published", n.Status)
	}
}

// TestUpdateNoteVisibility_OwnerCanChange はチームオーナーによる公開設定変更テストです。
func TestUpdateNoteVisibility_OwnerCanChange(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "team-owner-1", Role: domain.MemberRoleOwner})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	n, err := uc.UpdateNoteVisibility(context.Background(), "n-1", testTeamID, usecase.UpdateNoteVisibilityInput{
		CallerID:   "team-owner-1",
		CallerRole: domain.RoleUser,
		Status:     domain.NoteStatusPrivate,
	})

	if err != nil {
		t.Fatalf("チームオーナーによる公開設定変更に失敗しました: %v", err)
	}
	if n.Status != domain.NoteStatusPrivate {
		t.Errorf("statusが期待値と異なります: got %s, want private", n.Status)
	}
}

// TestUpdateNoteVisibility_PermissionDenied は権限のないユーザーによる変更テストです。
func TestUpdateNoteVisibility_PermissionDenied(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	_, err := uc.UpdateNoteVisibility(context.Background(), "n-1", testTeamID, usecase.UpdateNoteVisibilityInput{
		CallerID:   "user-2", // チームメンバーだが作成者でも owner でも admin でもない
		CallerRole: domain.RoleUser,
		Status:     domain.NoteStatusPublished,
	})

	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrPermissionDenied)
	}
}

// TestUpdateNoteVisibility_InvalidStatus は無効なステータスのテストです。
func TestUpdateNoteVisibility_InvalidStatus(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testNote("n-1", "ノート1", testCallerID))

	uc := newNoteUseCase(repo)

	_, err := uc.UpdateNoteVisibility(context.Background(), "n-1", testTeamID, usecase.UpdateNoteVisibilityInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Status:     domain.NoteStatus("invalid"),
	})

	if !errors.Is(err, domain.ErrInvalidNoteStatus) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidNoteStatus)
	}
}

// --- SearchNotes ページネーションテスト ---

// TestSearchNotes_Pagination はページネーションのテストです。
func TestSearchNotes_Pagination(t *testing.T) {
	repo := newMockNoteRepository()
	// 公開ノートを5件作成
	for i := 1; i <= 5; i++ {
		n := testPublishedNote(fmt.Sprintf("n-%d", i), testCallerID)
		n.Title = fmt.Sprintf("ノート%d", i)
		repo.addNote(n)
	}

	uc := newNoteUseCase(repo)

	// 1ページあたり2件、2ページ目を取得
	result, err := uc.SearchNotes(context.Background(), testTeamID, usecase.SearchNotesInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Page:       2,
		PerPage:    2,
	})
	if err != nil {
		t.Fatalf("ノート検索に失敗しました: %v", err)
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

// TestSearchNotes_DraftVisibility は draft ノートが可視性フィルタされるテストです。
func TestSearchNotes_DraftVisibility(t *testing.T) {
	repo := newMockNoteRepository()
	// testCallerID の draft ノート（user-2 からは見えない）
	repo.addNote(testNote("n-1", "ドラフトノート", testCallerID))
	// testCallerID の公開ノート（全メンバーに見える）
	repo.addNote(testPublishedNote("n-2", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: "user-2", Role: domain.MemberRoleMember})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	// user-2 として検索 → 公開ノート n-2 のみ見える
	result, err := uc.SearchNotes(context.Background(), testTeamID, usecase.SearchNotesInput{
		CallerID:   "user-2",
		CallerRole: domain.RoleUser,
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("ノート検索に失敗しました: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("可視性フィルタ後の総件数が期待値と異なります: got %d, want 1", result.Total)
	}
}

// TestSearchNotes_KeywordSearch はキーワード検索テストです（title/body/discussion_points/memo）。
func TestSearchNotes_KeywordSearch(t *testing.T) {
	repo := newMockNoteRepository()
	n1 := testPublishedNote("n-1", testCallerID)
	n1.Title = "暗号化の基礎"
	n2 := testPublishedNote("n-2", testCallerID)
	n2.Title = "リスク管理"
	repo.addNote(n1)
	repo.addNote(n2)

	uc := newNoteUseCase(repo)

	result, err := uc.SearchNotes(context.Background(), testTeamID, usecase.SearchNotesInput{
		CallerID:   testCallerID,
		CallerRole: domain.RoleUser,
		Keyword:    "暗号化",
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("ノート検索に失敗しました: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("総件数が期待値と異なります: got %d, want 1", result.Total)
	}
}

// TestSearchNotes_NonMemberDenied はチーム非メンバーが検索できないテストです。
func TestSearchNotes_NonMemberDenied(t *testing.T) {
	repo := newMockNoteRepository()
	repo.addNote(testPublishedNote("n-1", testCallerID))

	uc := newNoteUseCase(repo) // testCallerID のみメンバー登録済み

	_, err := uc.SearchNotes(context.Background(), testTeamID, usecase.SearchNotesInput{
		CallerID:   "user-not-member",
		CallerRole: domain.RoleUser,
		Page:       1,
		PerPage:    20,
	})
	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("チーム非メンバーは 403 になるべきです: got %v, want %v", err, domain.ErrMemberNotFound)
	}
}

// TestSearchNotes_AdminSeesAll は admin が全ノート（draft含む）を閲覧できるテストです。
func TestSearchNotes_AdminSeesAll(t *testing.T) {
	repo := newMockNoteRepository()
	// testCallerID の draft ノート
	repo.addNote(testNote("n-1", "ドラフトノート", testCallerID))
	// testCallerID の公開ノート
	repo.addNote(testPublishedNote("n-2", testCallerID))

	tRepo := newMockTeamRepository()
	_ = tRepo.AddMember(context.Background(), &domain.TeamMember{TeamID: testTeamID, UserID: testCallerID, Role: domain.MemberRoleMember})

	uc := newNoteUseCaseWithTeam(repo, tRepo)

	// admin として検索 → 全件見える
	result, err := uc.SearchNotes(context.Background(), testTeamID, usecase.SearchNotesInput{
		CallerID:   "admin-1",
		CallerRole: domain.RoleAdmin,
		Page:       1,
		PerPage:    20,
	})
	if err != nil {
		t.Fatalf("ノート検索に失敗しました: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("admin は全ノートを閲覧できるべきです: got %d, want 2", result.Total)
	}
}
