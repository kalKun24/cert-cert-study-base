package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// --- タグ用モック定義 ---

// mockTagRepository は domain.TagRepository のモックです。
type mockTagRepository struct {
	tags      map[string]*domain.Tag
	saveErr   error
	deleteErr error
}

func newMockTagRepository() *mockTagRepository {
	return &mockTagRepository{
		tags: make(map[string]*domain.Tag),
	}
}

// addTag はテスト用のタグをモックに追加します。
func (m *mockTagRepository) addTag(t *domain.Tag) {
	m.tags[t.ID] = t
}

func (m *mockTagRepository) FindByID(_ context.Context, id string) (*domain.Tag, error) {
	if t, ok := m.tags[id]; ok {
		return t, nil
	}
	return nil, domain.ErrTagNotFound
}

func (m *mockTagRepository) FindByName(_ context.Context, name string) (*domain.Tag, error) {
	for _, t := range m.tags {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, domain.ErrTagNotFound
}

func (m *mockTagRepository) List(_ context.Context) ([]*domain.Tag, error) {
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
	delete(m.tags, id)
	return nil
}

// --- テストヘルパー ---

// testTag はテスト用のタグエンティティを生成します。
func testTag(id, name string) *domain.Tag {
	return &domain.Tag{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
}

// testQuestionWithTags はタグIDを持つテスト用の問題エンティティを生成します。
func testQuestionWithTags(id string, tagIDs []string) *domain.Question {
	return &domain.Question{
		ID:               id,
		Title:            "テスト問題",
		Body:             "## 問題\nテスト問題文",
		Tags:             tagIDs,
		Status:           domain.QuestionStatusDraft,
		VisibilityScope:  domain.VisibilityScopeAll,
		PublishedTeamIDs: []string{},
		CreatedBy:        "user-1",
	}
}

// --- TagUseCase のテスト ---

// TestTagUseCase_ListTags_Success は正常系のタグ一覧取得テストです。
func TestTagUseCase_ListTags_Success(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))
	tagRepo.addTag(testTag("tag-2", "ドメイン1"))

	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	tags, err := uc.ListTags(context.Background())
	if err != nil {
		t.Fatalf("タグ一覧取得に失敗しました: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(tags))
	}
}

// TestTagUseCase_ListTags_Empty はタグが0件の場合のテストです。
func TestTagUseCase_ListTags_Empty(t *testing.T) {
	tagRepo := newMockTagRepository()
	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	tags, err := uc.ListTags(context.Background())
	if err != nil {
		t.Fatalf("タグ一覧取得に失敗しました: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 0", len(tags))
	}
}

// TestTagUseCase_CreateTag_Success は正常系のタグ作成テストです。
func TestTagUseCase_CreateTag_Success(t *testing.T) {
	tagRepo := newMockTagRepository()
	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	tag, err := uc.CreateTag(context.Background(), usecase.CreateTagInput{Name: "CISSP"})
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

// TestTagUseCase_CreateTag_EmptyName はタグ名が空の場合のテストです。
func TestTagUseCase_CreateTag_EmptyName(t *testing.T) {
	tagRepo := newMockTagRepository()
	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	_, err := uc.CreateTag(context.Background(), usecase.CreateTagInput{Name: ""})
	if err == nil {
		t.Fatal("タグ名が空の場合はエラーが返されるべきです")
	}
}

// TestTagUseCase_CreateTag_DuplicateName はタグ名が重複している場合のテストです。
func TestTagUseCase_CreateTag_DuplicateName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	_, err := uc.CreateTag(context.Background(), usecase.CreateTagInput{Name: "CISSP"})
	if !errors.Is(err, domain.ErrTagNameAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagNameAlreadyExists)
	}
}

// TestTagUseCase_UpdateTag_Success は正常系のタグ更新テストです。
func TestTagUseCase_UpdateTag_Success(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	newName := "CISSP 2024"
	tag, err := uc.UpdateTag(context.Background(), "tag-1", usecase.UpdateTagInput{Name: &newName})
	if err != nil {
		t.Fatalf("タグ更新に失敗しました: %v", err)
	}
	if tag.Name != "CISSP 2024" {
		t.Errorf("タグ名が期待値と異なります: got %s, want CISSP 2024", tag.Name)
	}
}

// TestTagUseCase_UpdateTag_SameName は同じ名前に更新する場合のテストです（自分自身への重複は許可）。
func TestTagUseCase_UpdateTag_SameName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	sameName := "CISSP"
	tag, err := uc.UpdateTag(context.Background(), "tag-1", usecase.UpdateTagInput{Name: &sameName})
	if err != nil {
		t.Fatalf("同名更新に失敗しました: %v", err)
	}
	if tag.Name != "CISSP" {
		t.Errorf("タグ名が期待値と異なります: got %s, want CISSP", tag.Name)
	}
}

// TestTagUseCase_UpdateTag_DuplicateName は他タグと重複する名前への更新テストです。
func TestTagUseCase_UpdateTag_DuplicateName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))
	tagRepo.addTag(testTag("tag-2", "ドメイン1"))

	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	duplicateName := "ドメイン1"
	_, err := uc.UpdateTag(context.Background(), "tag-1", usecase.UpdateTagInput{Name: &duplicateName})
	if !errors.Is(err, domain.ErrTagNameAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagNameAlreadyExists)
	}
}

// TestTagUseCase_UpdateTag_NotFound は存在しないIDで更新した場合のテストです。
func TestTagUseCase_UpdateTag_NotFound(t *testing.T) {
	tagRepo := newMockTagRepository()
	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	newName := "CISSP"
	_, err := uc.UpdateTag(context.Background(), "nonexistent-id", usecase.UpdateTagInput{Name: &newName})
	if !errors.Is(err, domain.ErrTagNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagNotFound)
	}
}

// TestTagUseCase_UpdateTag_EmptyName はタグ名を空にする更新テストです。
func TestTagUseCase_UpdateTag_EmptyName(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	emptyName := ""
	_, err := uc.UpdateTag(context.Background(), "tag-1", usecase.UpdateTagInput{Name: &emptyName})
	if err == nil {
		t.Fatal("タグ名を空にする場合はエラーが返されるべきです")
	}
}

// TestTagUseCase_DeleteTag_Success は正常系のタグ削除テストです。
func TestTagUseCase_DeleteTag_Success(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	err := uc.DeleteTag(context.Background(), "tag-1")
	if err != nil {
		t.Fatalf("タグ削除に失敗しました: %v", err)
	}

	// 削除後に取得しようとすると NotFound になること
	_, err = tagRepo.FindByID(context.Background(), "tag-1")
	if !errors.Is(err, domain.ErrTagNotFound) {
		t.Error("削除後もタグが存在しています")
	}
}

// TestTagUseCase_DeleteTag_NotFound は存在しないIDで削除した場合のテストです。
func TestTagUseCase_DeleteTag_NotFound(t *testing.T) {
	tagRepo := newMockTagRepository()
	questionRepo := newMockQuestionRepository()
	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	err := uc.DeleteTag(context.Background(), "nonexistent-id")
	if !errors.Is(err, domain.ErrTagNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagNotFound)
	}
}

// TestTagUseCase_DeleteTag_InUse は使用中タグを削除しようとした場合のテストです。
func TestTagUseCase_DeleteTag_InUse(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))

	questionRepo := newMockQuestionRepository()
	// tag-1 を使用している問題を追加
	questionRepo.addQuestion(testQuestionWithTags("q-1", []string{"tag-1", "tag-2"}))

	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	err := uc.DeleteTag(context.Background(), "tag-1")
	if !errors.Is(err, domain.ErrTagInUse) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrTagInUse)
	}
}

// TestTagUseCase_DeleteTag_NotInUse は他の問題が別タグを使用している場合の削除テストです。
func TestTagUseCase_DeleteTag_NotInUse(t *testing.T) {
	tagRepo := newMockTagRepository()
	tagRepo.addTag(testTag("tag-1", "CISSP"))
	tagRepo.addTag(testTag("tag-2", "ドメイン1"))

	questionRepo := newMockQuestionRepository()
	// tag-2 のみを使用している問題を追加
	questionRepo.addQuestion(testQuestionWithTags("q-1", []string{"tag-2"}))

	uc := usecase.NewTagUseCase(tagRepo, questionRepo)

	// tag-1 は使用されていないので削除可能
	err := uc.DeleteTag(context.Background(), "tag-1")
	if err != nil {
		t.Fatalf("未使用タグの削除に失敗しました: %v", err)
	}
}
