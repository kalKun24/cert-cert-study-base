package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// --- テスト定数 ---

const (
	testCallerID = "user-1"
	testTeamID   = "team-1"
)

// --- タグ用モック定義 ---

type mockTagRepository struct {
	tags      map[string]*domain.Tag
	questions map[string]*domain.Question
	saveErr   error
	deleteErr error
}

func newMockTagRepository() *mockTagRepository {
	return &mockTagRepository{
		tags:      make(map[string]*domain.Tag),
		questions: make(map[string]*domain.Question),
	}
}

func (m *mockTagRepository) addTag(t *domain.Tag) {
	m.tags[t.ID] = t
}

func (m *mockTagRepository) addQuestionForTagCheck(q *domain.Question) {
	m.questions[q.ID] = q
}

func (m *mockTagRepository) FindByID(_ context.Context, id string) (*domain.Tag, error) {
	if t, ok := m.tags[id]; ok {
		return t, nil
	}
	return nil, domain.ErrTagNotFound
}

func (m *mockTagRepository) FindByName(_ context.Context, _ string, name string) (*domain.Tag, error) {
	for _, t := range m.tags {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, domain.ErrTagNotFound
}

func (m *mockTagRepository) ListByTeam(_ context.Context, _ string) ([]*domain.Tag, error) {
	tags := make([]*domain.Tag, 0, len(m.tags))
	for _, t := range m.tags {
		tags = append(tags, t)
	}
	return tags, nil
}

func (m *mockTagRepository) Save(_ context.Context, tag *domain.Tag) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.tags[tag.ID] = tag
	return nil
}

func (m *mockTagRepository) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.tags[id]; !ok {
		return domain.ErrTagNotFound
	}
	for _, q := range m.questions {
		for _, tid := range q.Tags {
			if tid == id {
				return domain.ErrTagInUse
			}
		}
	}
	delete(m.tags, id)
	return nil
}

// --- テストヘルパー ---

// newTagTeamRepo はテスト用チームリポジトリを返します。
// testCallerID が testTeamID のメンバーとして登録されています。
func newTagTeamRepo() *mockTeamRepository {
	teamRepo := newMockTeamRepository()
	_ = teamRepo.AddMember(context.Background(), &domain.TeamMember{
		TeamID: testTeamID,
		UserID: testCallerID,
		Role:   domain.MemberRoleMember,
	})
	return teamRepo
}

func newTagUseCase(tagRepo *mockTagRepository) *usecase.TagUseCase {
	return usecase.NewTagUseCase(tagRepo, newTagTeamRepo())
}

func testTag(id, name string) *domain.Tag {
	return &domain.Tag{
		ID:        id,
		TeamID:    testTeamID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
}

func testQuestionWithTags(id string, tagIDs []string) *domain.Question {
	return &domain.Question{
		ID:               id,
		Title:            "テスト問題",
		Body:             "## 問題\nテスト問題文",
		Tags:             tagIDs,
		Status:           domain.QuestionStatusDraft,
		VisibilityScope:  domain.VisibilityScopeAll,
		PublishedTeamIDs: []string{},
		CreatedBy:        testCallerID,
	}
}

// --- TagUseCase のテスト ---

func TestTagUseCase_ListTags_Success(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))
	tagRepo.addTag(testTag("tag-2", "ドメイン1"))

	uc := newTagUseCase(tagRepo)

	tags, err := uc.ListTags(context.Background(), testCallerID, domain.RoleUser, testTeamID)
	if err != nil {
		t.Fatalf("タグ一覧取得に失敗しました: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(tags))
	}
}

func TestTagUseCase_ListTags_Empty(t *testing.T) {
	uc := newTagUseCase(newMockTagRepository())

	tags, err := uc.ListTags(context.Background(), testCallerID, domain.RoleUser, testTeamID)
	if err != nil {
		t.Fatalf("タグ一覧取得に失敗しました: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 0", len(tags))
	}
}

func TestTagUseCase_CreateTag_Success(t *testing.T) {
	uc := newTagUseCase(newMockTagRepository())

	tag, err := uc.CreateTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, usecase.CreateTagInput{Name: "CISSP"})
	if err != nil {
		t.Fatalf("タグ作成に失敗しました: %v", err)
	}
	if tag.Name != "CISSP" {
		t.Errorf("タグ名が期待値と異なります: got %s, want CISSP", tag.Name)
	}
	if tag.ID == "" {
		t.Error("IDが生成されていません")
	}
	if tag.CreatedAt.IsZero() {
		t.Error("CreatedAtが設定されていません")
	}
}

func TestTagUseCase_CreateTag_EmptyName(t *testing.T) {
	uc := newTagUseCase(newMockTagRepository())

	_, err := uc.CreateTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, usecase.CreateTagInput{Name: ""})
	if err == nil {
		t.Fatal("タグ名が空の場合はエラーが返されるべきです")
	}
}

func TestTagUseCase_CreateTag_WhitespaceOnly(t *testing.T) {
	uc := newTagUseCase(newMockTagRepository())

	_, err := uc.CreateTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, usecase.CreateTagInput{Name: "   "})
	if err == nil {
		t.Fatal("タグ名が空白のみの場合はエラーが返されるべきです")
	}
}

func TestTagUseCase_CreateTag_TooLongName(t *testing.T) {
	uc := newTagUseCase(newMockTagRepository())

	longName := string(make([]rune, 51))
	for i := range longName {
		longName = longName[:i] + "a" + longName[i+1:]
	}
	_, err := uc.CreateTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, usecase.CreateTagInput{Name: longName})
	if err == nil {
		t.Fatal("タグ名が50文字を超える場合はエラーが返されるべきです")
	}
}

func TestTagUseCase_CreateTag_DuplicateName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	uc := newTagUseCase(tagRepo)

	_, err := uc.CreateTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, usecase.CreateTagInput{Name: "CISSP"})
	if !errors.Is(err, domain.ErrTagNameAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagNameAlreadyExists)
	}
}

func TestTagUseCase_UpdateTag_Success(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	uc := newTagUseCase(tagRepo)

	newName := "CISSP 2024"
	tag, err := uc.UpdateTag(context.Background(), domain.RoleAdmin, testTeamID, "tag-1", usecase.UpdateTagInput{Name: &newName})
	if err != nil {
		t.Fatalf("タグ更新に失敗しました: %v", err)
	}
	if tag.Name != "CISSP 2024" {
		t.Errorf("タグ名が期待値と異なります: got %s, want CISSP 2024", tag.Name)
	}
}

func TestTagUseCase_UpdateTag_SameName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	uc := newTagUseCase(tagRepo)

	sameName := "CISSP"
	tag, err := uc.UpdateTag(context.Background(), domain.RoleAdmin, testTeamID, "tag-1", usecase.UpdateTagInput{Name: &sameName})
	if err != nil {
		t.Fatalf("同名更新に失敗しました: %v", err)
	}
	if tag.Name != "CISSP" {
		t.Errorf("タグ名が期待値と異なります: got %s, want CISSP", tag.Name)
	}
}

func TestTagUseCase_UpdateTag_DuplicateName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))
	tagRepo.addTag(testTag("tag-2", "ドメイン1"))

	uc := newTagUseCase(tagRepo)

	duplicateName := "ドメイン1"
	_, err := uc.UpdateTag(context.Background(), domain.RoleAdmin, testTeamID, "tag-1", usecase.UpdateTagInput{Name: &duplicateName})
	if !errors.Is(err, domain.ErrTagNameAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagNameAlreadyExists)
	}
}

func TestTagUseCase_UpdateTag_NotFound(t *testing.T) {
	uc := newTagUseCase(newMockTagRepository())

	newName := "CISSP"
	_, err := uc.UpdateTag(context.Background(), domain.RoleAdmin, testTeamID, "nonexistent-id", usecase.UpdateTagInput{Name: &newName})
	if !errors.Is(err, domain.ErrTagNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagNotFound)
	}
}

func TestTagUseCase_UpdateTag_EmptyName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	uc := newTagUseCase(tagRepo)

	emptyName := ""
	_, err := uc.UpdateTag(context.Background(), domain.RoleAdmin, testTeamID, "tag-1", usecase.UpdateTagInput{Name: &emptyName})
	if err == nil {
		t.Fatal("タグ名を空にする場合はエラーが返されるべきです")
	}
}

func TestTagUseCase_UpdateTag_WhitespaceOnly(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	uc := newTagUseCase(tagRepo)

	whitespaceName := "   "
	_, err := uc.UpdateTag(context.Background(), domain.RoleAdmin, testTeamID, "tag-1", usecase.UpdateTagInput{Name: &whitespaceName})
	if err == nil {
		t.Fatal("タグ名が空白のみの場合はエラーが返されるべきです")
	}
}

func TestTagUseCase_UpdateTag_TooLongName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	uc := newTagUseCase(tagRepo)

	longName := string(make([]rune, 51))
	for i := range longName {
		longName = longName[:i] + "a" + longName[i+1:]
	}
	_, err := uc.UpdateTag(context.Background(), domain.RoleAdmin, testTeamID, "tag-1", usecase.UpdateTagInput{Name: &longName})
	if err == nil {
		t.Fatal("タグ名が50文字を超える場合はエラーが返されるべきです")
	}
}

func TestTagUseCase_DeleteTag_Success(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	uc := newTagUseCase(tagRepo)

	err := uc.DeleteTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, "tag-1")
	if err != nil {
		t.Fatalf("タグ削除に失敗しました: %v", err)
	}

	_, err = tagRepo.FindByID(context.Background(), "tag-1")
	if !errors.Is(err, domain.ErrTagNotFound) {
		t.Error("削除後もタグが存在しています")
	}
}

func TestTagUseCase_DeleteTag_NotFound(t *testing.T) {
	uc := newTagUseCase(newMockTagRepository())

	err := uc.DeleteTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, "nonexistent-id")
	if !errors.Is(err, domain.ErrTagNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagNotFound)
	}
}

func TestTagUseCase_DeleteTag_InUse(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))
	tagRepo.addQuestionForTagCheck(testQuestionWithTags("q-1", []string{"tag-1", "tag-2"}))

	uc := newTagUseCase(tagRepo)

	err := uc.DeleteTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, "tag-1")
	if !errors.Is(err, domain.ErrTagInUse) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagInUse)
	}
}

func TestTagUseCase_DeleteTag_NotInUse(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))
	tagRepo.addTag(testTag("tag-2", "ドメイン1"))
	tagRepo.addQuestionForTagCheck(testQuestionWithTags("q-1", []string{"tag-2"}))

	uc := newTagUseCase(tagRepo)

	err := uc.DeleteTag(context.Background(), testCallerID, domain.RoleUser, testTeamID, "tag-1")
	if err != nil {
		t.Fatalf("未使用タグの削除に失敗しました: %v", err)
	}
}
