package repository_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/repository"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/storage"
)

// setupIntegrationTest はエミュレータへの接続とテスト用バケットの準備を行います。
// GCS_EMULATOR_HOST が未設定の場合はテストをスキップします。
func setupIntegrationTest(t *testing.T) *repository.GCSUserRepository {
	t.Helper()

	emulatorHost := os.Getenv("GCS_EMULATOR_HOST")
	if emulatorHost == "" {
		t.Skip("GCS_EMULATOR_HOST が未設定のためスキップします（エミュレータ起動後に再実行してください）")
	}

	ctx := context.Background()

	gcsClient, err := storage.NewClientFromEnv(ctx)
	if err != nil {
		t.Fatalf("GCS クライアントの初期化に失敗しました: %v", err)
	}
	t.Cleanup(func() { _ = gcsClient.Close() })

	// テストごとに独立したバケットを使用して競合を避ける
	bucket := fmt.Sprintf("test-bucket-%d", time.Now().UnixNano())
	if err := storage.EnsureBucketExists(ctx, gcsClient, bucket); err != nil {
		t.Fatalf("テスト用バケットの作成に失敗しました: %v", err)
	}

	sc := storage.NewGCSStorageClient(gcsClient)
	return repository.NewGCSUserRepository(sc, bucket)
}

func TestGCSUserRepository_SaveAndFindByID(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	user := newTestUser("user-001")
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save に失敗しました: %v", err)
	}

	got, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID に失敗しました: %v", err)
	}
	assertUserEqual(t, user, got)
}

func TestGCSUserRepository_FindByUsername(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	user := newTestUser("user-002")
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save に失敗しました: %v", err)
	}

	got, err := repo.FindByUsername(ctx, user.Username)
	if err != nil {
		t.Fatalf("FindByUsername に失敗しました: %v", err)
	}
	assertUserEqual(t, user, got)
}

func TestGCSUserRepository_FindByEmail(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	user := newTestUser("user-003")
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save に失敗しました: %v", err)
	}

	got, err := repo.FindByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("FindByEmail に失敗しました: %v", err)
	}
	assertUserEqual(t, user, got)
}

func TestGCSUserRepository_List(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	users := []*domain.User{
		newTestUser("user-101"),
		newTestUser("user-102"),
		newTestUser("user-103"),
	}
	for _, u := range users {
		if err := repo.Save(ctx, u); err != nil {
			t.Fatalf("Save に失敗しました: %v", err)
		}
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List に失敗しました: %v", err)
	}
	if len(list) != len(users) {
		t.Errorf("List の件数 = %d, want %d", len(list), len(users))
	}
}

func TestGCSUserRepository_Update(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	user := newTestUser("user-201")
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save に失敗しました: %v", err)
	}

	user.DisplayName = "Updated Name"
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Update（Save）に失敗しました: %v", err)
	}

	got, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("FindByID に失敗しました: %v", err)
	}
	if got.DisplayName != "Updated Name" {
		t.Errorf("DisplayName = %q, want %q", got.DisplayName, "Updated Name")
	}
}

func TestGCSUserRepository_Delete(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	user := newTestUser("user-301")
	if err := repo.Save(ctx, user); err != nil {
		t.Fatalf("Save に失敗しました: %v", err)
	}

	if err := repo.Delete(ctx, user.ID); err != nil {
		t.Fatalf("Delete に失敗しました: %v", err)
	}

	if _, err := repo.FindByID(ctx, user.ID); err != domain.ErrUserNotFound {
		t.Errorf("削除後は ErrUserNotFound が返るはず, got: %v", err)
	}
}

func TestGCSUserRepository_FindByID_NotFound(t *testing.T) {
	repo := setupIntegrationTest(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "non-existent-id")
	if err != domain.ErrUserNotFound {
		t.Errorf("ErrUserNotFound が返るはず, got: %v", err)
	}
}

// --- ヘルパー ---

func newTestUser(id string) *domain.User {
	return &domain.User{
		ID:           id,
		Username:     "user-" + id,
		DisplayName:  "Test User " + id,
		Email:        id + "@example.com",
		PasswordHash: "$2a$12$dummyhash",
		Role:         domain.RoleUser,
		IsActive:     true,
		CreatedAt:    time.Now().UTC().Truncate(time.Second),
		UpdatedAt:    time.Now().UTC().Truncate(time.Second),
	}
}

func assertUserEqual(t *testing.T, want, got *domain.User) {
	t.Helper()
	if got.ID != want.ID {
		t.Errorf("ID = %q, want %q", got.ID, want.ID)
	}
	if got.Username != want.Username {
		t.Errorf("Username = %q, want %q", got.Username, want.Username)
	}
	if got.Email != want.Email {
		t.Errorf("Email = %q, want %q", got.Email, want.Email)
	}
	if got.Role != want.Role {
		t.Errorf("Role = %q, want %q", got.Role, want.Role)
	}
	if got.IsActive != want.IsActive {
		t.Errorf("IsActive = %v, want %v", got.IsActive, want.IsActive)
	}
}
