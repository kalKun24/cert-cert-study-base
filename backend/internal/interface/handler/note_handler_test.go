package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// ---- ノートリポジトリのスタブ ----

type stubNoteRepo struct {
	saved *domain.Note
}

func (s *stubNoteRepo) FindByID(_ context.Context, _, _ string) (*domain.Note, error) {
	return nil, domain.ErrNoteNotFound
}
func (s *stubNoteRepo) ListByTeam(_ context.Context, _ string) ([]*domain.Note, error) {
	return nil, nil
}
func (s *stubNoteRepo) SearchByTeam(_ context.Context, _ string, _ domain.NoteSearchFilter) ([]*domain.Note, error) {
	return nil, nil
}
func (s *stubNoteRepo) Save(_ context.Context, n *domain.Note) error {
	s.saved = n
	return nil
}
func (s *stubNoteRepo) Delete(_ context.Context, _, _ string) error { return nil }
func (s *stubNoteRepo) FindByTagID(_ context.Context, _, _ string) ([]*domain.Note, error) {
	return nil, nil
}

// ---- ヘルパー ----

const (
	testNoteHandlerTeamID   = "11111111-1111-1111-1111-111111111111"
	testNoteHandlerCallerID = "22222222-2222-2222-2222-222222222222"
	testNoteHandlerNoteID   = "55555555-5555-5555-5555-555555555555"
)

func newTestNoteHandler(isMember bool) *NoteHandler {
	nRepo := &stubNoteRepo{}
	tRepo := &stubTeamRepo{isMember: isMember}
	uc := usecase.NewNoteUseCase(nRepo, tRepo)
	return NewNoteHandler(uc)
}

// authedNoteRequest は認証済みコンテキストを付与したリクエストを返します。
func authedNoteRequest(method, target string, body []byte) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), ContextKeyUserID, testNoteHandlerCallerID)
	ctx = context.WithValue(ctx, ContextKeyUserRole, string(domain.RoleUser))
	return r.WithContext(ctx)
}

// stubNoteRepoWithNote は指定ノートを返す stubNoteRepo です。
type stubNoteRepoWithNote struct {
	note *domain.Note
}

func (s *stubNoteRepoWithNote) FindByID(_ context.Context, teamID, id string) (*domain.Note, error) {
	if s.note != nil && s.note.ID == id && s.note.TeamID == teamID {
		return s.note, nil
	}
	return nil, domain.ErrNoteNotFound
}
func (s *stubNoteRepoWithNote) ListByTeam(_ context.Context, _ string) ([]*domain.Note, error) {
	return nil, nil
}
func (s *stubNoteRepoWithNote) SearchByTeam(_ context.Context, _ string, _ domain.NoteSearchFilter) ([]*domain.Note, error) {
	return nil, nil
}
func (s *stubNoteRepoWithNote) Save(_ context.Context, n *domain.Note) error {
	s.note = n
	return nil
}
func (s *stubNoteRepoWithNote) Delete(_ context.Context, _, _ string) error { return nil }
func (s *stubNoteRepoWithNote) FindByTagID(_ context.Context, _, _ string) ([]*domain.Note, error) {
	return nil, nil
}

// ---- UpdateNoteVisibility のステータスバリデーションテスト ----

// TestUpdateNoteVisibility_InvalidStatus は不正なステータス文字列で400が返ることを確認します。
func TestUpdateNoteVisibility_InvalidStatus(t *testing.T) {
	h := newTestNoteHandler(true)

	body, _ := json.Marshal(map[string]any{
		"status": "invalid_status",
	})

	req := authedNoteRequest(http.MethodPatch, "/", body)
	req.SetPathValue("team_id", testNoteHandlerTeamID)
	req.SetPathValue("note_id", testNoteHandlerNoteID)

	w := httptest.NewRecorder()
	h.HandleUpdateNoteVisibility(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("不正なステータスで 400 が期待されますが %d でした（レスポンス: %s）", w.Code, w.Body.String())
	}
}

// TestUpdateNoteVisibility_EmptyStatus は status が空文字列で400が返ることを確認します。
func TestUpdateNoteVisibility_EmptyStatus(t *testing.T) {
	h := newTestNoteHandler(true)

	body, _ := json.Marshal(map[string]any{
		"status": "",
	})

	req := authedNoteRequest(http.MethodPatch, "/", body)
	req.SetPathValue("team_id", testNoteHandlerTeamID)
	req.SetPathValue("note_id", testNoteHandlerNoteID)

	w := httptest.NewRecorder()
	h.HandleUpdateNoteVisibility(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("空ステータスで 400 が期待されますが %d でした", w.Code)
	}
}

// TestUpdateNoteVisibility_ValidStatus は有効なステータスでバリデーションを通過することを確認します。
// stubNoteRepo.FindByID は ErrNoteNotFound を返すため 404 になります。
// これによりバリデーション自体は通過していることが確認できます。
func TestUpdateNoteVisibility_ValidStatus(t *testing.T) {
	h := newTestNoteHandler(true)

	body, _ := json.Marshal(map[string]any{
		"status": "published",
	})

	req := authedNoteRequest(http.MethodPatch, "/", body)
	req.SetPathValue("team_id", testNoteHandlerTeamID)
	req.SetPathValue("note_id", testNoteHandlerNoteID)

	w := httptest.NewRecorder()
	h.HandleUpdateNoteVisibility(w, req)

	// バリデーションは通過するので 400 にはならない
	if w.Code == http.StatusBadRequest {
		t.Errorf("有効なステータス 'published' で 400 は不正です（レスポンス: %s）", w.Body.String())
	}
}

// TestUpdateNoteVisibility_ValidStatus_WithExistingNote はノートが存在する場合の正常系を確認します。
func TestUpdateNoteVisibility_ValidStatus_WithExistingNote(t *testing.T) {
	existingNote := &domain.Note{
		ID:        testNoteHandlerNoteID,
		TeamID:    testNoteHandlerTeamID,
		Title:     "テストノート",
		CreatedBy: testNoteHandlerCallerID,
		Status:    domain.NoteStatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      []string{},
	}
	nRepo := &stubNoteRepoWithNote{note: existingNote}
	tRepo := &stubTeamRepo{isMember: true}
	uc := usecase.NewNoteUseCase(nRepo, tRepo)
	h := NewNoteHandler(uc)

	body, _ := json.Marshal(map[string]any{
		"status": "published",
	})

	req := authedNoteRequest(http.MethodPatch, "/", body)
	req.SetPathValue("team_id", testNoteHandlerTeamID)
	req.SetPathValue("note_id", testNoteHandlerNoteID)

	w := httptest.NewRecorder()
	h.HandleUpdateNoteVisibility(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("有効なステータスで既存ノートを更新すると 200 が期待されますが %d でした（レスポンス: %s）", w.Code, w.Body.String())
	}
}

// ---- ページネーション境界値テスト ----

// TestListNotes_PageZero は page=0 のとき 400 が返ることを確認します。
func TestListNotes_PageZero(t *testing.T) {
	h := newTestNoteHandler(true)

	req := authedNoteRequest(http.MethodGet, "/?page=0", nil)
	req.SetPathValue("team_id", testNoteHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListNotes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("page=0 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestListNotes_PageNegative は page=-1 のとき 400 が返ることを確認します。
func TestListNotes_PageNegative(t *testing.T) {
	h := newTestNoteHandler(true)

	req := authedNoteRequest(http.MethodGet, "/?page=-1", nil)
	req.SetPathValue("team_id", testNoteHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListNotes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("page=-1 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestListNotes_PerPageZero は per_page=0 のとき 400 が返ることを確認します。
func TestListNotes_PerPageZero(t *testing.T) {
	h := newTestNoteHandler(true)

	req := authedNoteRequest(http.MethodGet, "/?per_page=0", nil)
	req.SetPathValue("team_id", testNoteHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListNotes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("per_page=0 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestListNotes_PerPageNegative は per_page=-5 のとき 400 が返ることを確認します。
func TestListNotes_PerPageNegative(t *testing.T) {
	h := newTestNoteHandler(true)

	req := authedNoteRequest(http.MethodGet, "/?per_page=-5", nil)
	req.SetPathValue("team_id", testNoteHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListNotes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("per_page=-5 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestListNotes_PerPageExceedsMax は per_page=101 のとき 400 が返ることを確認します。
func TestListNotes_PerPageExceedsMax(t *testing.T) {
	h := newTestNoteHandler(true)

	req := authedNoteRequest(http.MethodGet, "/?per_page=101", nil)
	req.SetPathValue("team_id", testNoteHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListNotes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("per_page=101 で 400 が期待されますが %d でした", w.Code)
	}
}

// ---- CreateNote のタグ UUID バリデーション ----

// TestCreateNote_TagsMustBeUUID はタグにUUID以外の文字列を送ると400になることを確認します。
func TestCreateNote_TagsMustBeUUID(t *testing.T) {
	h := newTestNoteHandler(true)

	body, _ := json.Marshal(map[string]any{
		"title": "テストノート",
		"body":  "本文",
		"tags":  []string{"暗号化"}, // UUID ではなくタグ名を送る
	})

	req := authedNoteRequest(http.MethodPost, "/", body)
	req.SetPathValue("team_id", testNoteHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleCreateNote(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("タグ名を送ったとき 400 が期待されますが %d でした", w.Code)
	}
}
