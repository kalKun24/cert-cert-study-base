package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/contextkey"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// ---- 最小限のモック実装 ----

type stubQuestionRepo struct {
	saved *domain.Question
}

func (s *stubQuestionRepo) FindByID(_ context.Context, _, _ string) (*domain.Question, error) {
	return nil, domain.ErrQuestionNotFound
}
func (s *stubQuestionRepo) ListByTeam(_ context.Context, _ string) ([]*domain.Question, error) {
	return nil, nil
}
func (s *stubQuestionRepo) SearchByTeam(_ context.Context, _ string, _ domain.QuestionSearchFilter) ([]*domain.Question, error) {
	return nil, nil
}
func (s *stubQuestionRepo) Save(_ context.Context, q *domain.Question) error {
	s.saved = q
	return nil
}
func (s *stubQuestionRepo) FindByTagID(_ context.Context, _, _ string) ([]*domain.Question, error) {
	return nil, nil
}
func (s *stubQuestionRepo) Delete(_ context.Context, _, _ string) error { return nil }

type stubTeamRepo struct {
	isMember bool
}

func (s *stubTeamRepo) FindByID(_ context.Context, _ string) (*domain.Team, error) {
	return nil, domain.ErrTeamNotFound
}
func (s *stubTeamRepo) FindByName(_ context.Context, _ string) (*domain.Team, error) {
	return nil, domain.ErrTeamNotFound
}
func (s *stubTeamRepo) List(_ context.Context) ([]*domain.Team, error) { return nil, nil }
func (s *stubTeamRepo) ListByOwnerOrMember(_ context.Context, _ string) ([]*domain.Team, error) {
	return nil, nil
}
func (s *stubTeamRepo) Save(_ context.Context, _ *domain.Team) error            { return nil }
func (s *stubTeamRepo) Delete(_ context.Context, _ string) error                { return nil }
func (s *stubTeamRepo) AddMember(_ context.Context, _ *domain.TeamMember) error { return nil }
func (s *stubTeamRepo) RemoveMember(_ context.Context, _, _ string) error       { return nil }
func (s *stubTeamRepo) ListMembers(_ context.Context, _ string) ([]*domain.TeamMember, error) {
	return nil, nil
}
func (s *stubTeamRepo) IsMember(_ context.Context, _, _ string) (bool, error) {
	return s.isMember, nil
}
func (s *stubTeamRepo) FindOwners(_ context.Context, _ string) ([]*domain.TeamMember, error) {
	return nil, nil
}
func (s *stubTeamRepo) UpdateMemberRole(_ context.Context, _, _ string, _ domain.MemberRole) error {
	return nil
}

// ---- ヘルパー ----

const (
	testHandlerTeamID   = "11111111-1111-1111-1111-111111111111"
	testHandlerCallerID = "22222222-2222-2222-2222-222222222222"
	testValidTagID      = "33333333-3333-3333-3333-333333333333"
)

func newTestQuestionHandler(isMember bool) *QuestionHandler {
	qRepo := &stubQuestionRepo{}
	tRepo := &stubTeamRepo{isMember: isMember}
	uc := usecase.NewQuestionUseCase(qRepo, tRepo)
	return NewQuestionHandler(uc)
}

// authedRequest は認証済みコンテキストを付与したリクエストを返します。
func authedRequest(method, target string, body []byte) *http.Request {
	r := httptest.NewRequest(method, target, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(r.Context(), contextkey.UserID, testHandlerCallerID)
	ctx = context.WithValue(ctx, contextkey.UserRole, string(domain.RoleUser))
	return r.WithContext(ctx)
}

// ---- テスト ----

// TestCreateQuestion_TagsMustBeUUID はタグにUUID以外の文字列を送ると400になることを確認します。
// このテストがあれば「タグ名を送ったらフィルタが壊れる」バグをAPIレベルで早期に検出できます。
func TestCreateQuestion_TagsMustBeUUID(t *testing.T) {
	h := newTestQuestionHandler(true)

	body, _ := json.Marshal(map[string]any{
		"title": "テスト問題",
		"body":  "本文",
		"tags":  []string{"暗号化"}, // UUID ではなくタグ名を送る（過去のバグのパターン）
	})

	req := authedRequest(http.MethodPost, "/", body)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleCreateQuestion(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("タグ名を送ったとき 400 が期待されますが %d でした", w.Code)
	}
}

// TestCreateQuestion_ValidTagUUID はUUID形式のタグIDを送ると正常に作成されることを確認します。
func TestCreateQuestion_ValidTagUUID(t *testing.T) {
	h := newTestQuestionHandler(true)

	body, _ := json.Marshal(map[string]any{
		"title": "テスト問題",
		"body":  "本文",
		"tags":  []string{testValidTagID},
	})

	req := authedRequest(http.MethodPost, "/", body)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleCreateQuestion(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("有効なタグIDを送ったとき 201 が期待されますが %d でした（レスポンス: %s）", w.Code, w.Body.String())
	}
}

// TestCreateQuestion_EmptyTags はタグなしで問題を作成できることを確認します。
func TestCreateQuestion_EmptyTags(t *testing.T) {
	h := newTestQuestionHandler(true)

	body, _ := json.Marshal(map[string]any{
		"title": "テスト問題",
		"body":  "本文",
		"tags":  []string{},
	})

	req := authedRequest(http.MethodPost, "/", body)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleCreateQuestion(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("タグなしで 201 が期待されますが %d でした", w.Code)
	}
}

// TestCreateQuestion_MultipleTagsMixedFormat は有効なUUIDと無効な文字列が混在するとき400になることを確認します。
func TestCreateQuestion_MultipleTagsMixedFormat(t *testing.T) {
	h := newTestQuestionHandler(true)

	body, _ := json.Marshal(map[string]any{
		"title": "テスト問題",
		"body":  "本文",
		"tags":  []string{testValidTagID, "アクセス制御"}, // 1つ目は正常、2つ目が名前
	})

	req := authedRequest(http.MethodPost, "/", body)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleCreateQuestion(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UUID混在時に 400 が期待されますが %d でした", w.Code)
	}
}

// ---- UpdateQuestionVisibility のテスト ----

// TestUpdateQuestionVisibility_InvalidStatus は不正なステータス文字列で400が返ることを確認します。
func TestUpdateQuestionVisibility_InvalidStatus(t *testing.T) {
	h := newTestQuestionHandler(true)

	body, _ := json.Marshal(map[string]any{
		"status": "invalid_status",
	})

	req := authedRequest(http.MethodPatch, "/", body)
	req.SetPathValue("team_id", testHandlerTeamID)
	req.SetPathValue("id", "44444444-4444-4444-4444-444444444444")

	w := httptest.NewRecorder()
	h.HandleUpdateQuestionVisibility(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("不正なステータスで 400 が期待されますが %d でした（レスポンス: %s）", w.Code, w.Body.String())
	}
}

// TestUpdateQuestionVisibility_EmptyStatus は status が空文字列で400が返ることを確認します。
func TestUpdateQuestionVisibility_EmptyStatus(t *testing.T) {
	h := newTestQuestionHandler(true)

	body, _ := json.Marshal(map[string]any{
		"status": "",
	})

	req := authedRequest(http.MethodPatch, "/", body)
	req.SetPathValue("team_id", testHandlerTeamID)
	req.SetPathValue("id", "44444444-4444-4444-4444-444444444444")

	w := httptest.NewRecorder()
	h.HandleUpdateQuestionVisibility(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("空ステータスで 400 が期待されますが %d でした", w.Code)
	}
}

// TestUpdateQuestionVisibility_ValidStatus は有効なステータスでハンドラーを通過することを確認します。
// stubQuestionRepo.FindByID は ErrQuestionNotFound を返すため、ハンドラーは 404 を返します。
// これによりバリデーション自体は通過していることが確認できます。
func TestUpdateQuestionVisibility_ValidStatus(t *testing.T) {
	h := newTestQuestionHandler(true)

	body, _ := json.Marshal(map[string]any{
		"status": "published",
	})

	req := authedRequest(http.MethodPatch, "/", body)
	req.SetPathValue("team_id", testHandlerTeamID)
	req.SetPathValue("id", "44444444-4444-4444-4444-444444444444")

	w := httptest.NewRecorder()
	h.HandleUpdateQuestionVisibility(w, req)

	// バリデーションは通過するので 400 にはならない（FindByID が ErrQuestionNotFound を返すので 404 になる）
	if w.Code == http.StatusBadRequest {
		t.Errorf("有効なステータス 'published' で 400 は不正です（レスポンス: %s）", w.Body.String())
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("FindByID が NotFound を返すため 404 が期待されますが %d でした", w.Code)
	}
}

// ---- ページネーション境界値テスト ----

// TestListQuestions_PageZero は page=0 のとき 400 が返ることを確認します。
func TestListQuestions_PageZero(t *testing.T) {
	h := newTestQuestionHandler(true)

	req := authedRequest(http.MethodGet, "/?page=0", nil)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListQuestions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("page=0 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestListQuestions_PageNegative は page=-1 のとき 400 が返ることを確認します。
func TestListQuestions_PageNegative(t *testing.T) {
	h := newTestQuestionHandler(true)

	req := authedRequest(http.MethodGet, "/?page=-1", nil)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListQuestions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("page=-1 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestListQuestions_PerPageZero は per_page=0 のとき 400 が返ることを確認します。
func TestListQuestions_PerPageZero(t *testing.T) {
	h := newTestQuestionHandler(true)

	req := authedRequest(http.MethodGet, "/?per_page=0", nil)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListQuestions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("per_page=0 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestListQuestions_PerPageNegative は per_page=-5 のとき 400 が返ることを確認します。
func TestListQuestions_PerPageNegative(t *testing.T) {
	h := newTestQuestionHandler(true)

	req := authedRequest(http.MethodGet, "/?per_page=-5", nil)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListQuestions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("per_page=-5 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestListQuestions_PerPageExceedsMax は per_page=101 のとき 400 が返ることを確認します。
func TestListQuestions_PerPageExceedsMax(t *testing.T) {
	h := newTestQuestionHandler(true)

	req := authedRequest(http.MethodGet, "/?per_page=101", nil)
	req.SetPathValue("team_id", testHandlerTeamID)

	w := httptest.NewRecorder()
	h.HandleListQuestions(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("per_page=101 で 400 が期待されますが %d でした", w.Code)
	}
}

// TestUpdateQuestion_TagsMustBeUUID は更新時もタグのUUID検証が効くことを確認します。
func TestUpdateQuestion_TagsMustBeUUID(t *testing.T) {
	// FindByID が返すダミー問題を設定
	qID := "44444444-4444-4444-4444-444444444444"
	dummyQ := &domain.Question{
		ID:        qID,
		TeamID:    testHandlerTeamID,
		Title:     "既存問題",
		CreatedBy: testHandlerCallerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      []string{},
		Status:    domain.QuestionStatusPublished,
	}
	// stubQuestionRepo の FindByID が返せるよう直接セット
	qRepo := &stubQuestionRepo{saved: dummyQ}
	tRepo := &stubTeamRepo{isMember: true}
	uc := usecase.NewQuestionUseCase(qRepo, tRepo)
	h2 := NewQuestionHandler(uc)

	body, _ := json.Marshal(map[string]any{
		"title": "更新後タイトル",
		"tags":  []string{"暗号化"},
	})

	req := authedRequest(http.MethodPut, "/", body)
	req.SetPathValue("team_id", testHandlerTeamID)
	req.SetPathValue("id", qID)

	w := httptest.NewRecorder()
	h2.HandleUpdateQuestion(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("更新時タグ名を送ったとき 400 が期待されますが %d でした", w.Code)
	}
}
