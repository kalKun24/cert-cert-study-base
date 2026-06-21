package firestore_test

import (
	"context"
	"os"
	"testing"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	firestoreRepo "github.com/kalKun24/cert-study-base/backend/internal/infrastructure/firestore"
)

// setupFirestoreTest は Firestore エミュレータに接続してテスト用リポジトリを返します。
// FIRESTORE_EMULATOR_HOST が未設定の場合はテストをスキップします。
func setupFirestoreTest(t *testing.T) *firestoreRepo.FirestoreUserRepository {
	t.Helper()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST が設定されていないためスキップします")
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "test-project"
	}

	ctx := context.Background()
	client, err := fs.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Logf("Firestoreクライアントのクローズに失敗しました: %v", err)
		}
	})

	return firestoreRepo.NewFirestoreUserRepository(client)
}

func TestFirestoreUserRepository_SaveAndFindByID(t *testing.T) {
	repo := setupFirestoreTest(t)
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

func TestFirestoreUserRepository_FindByUsername(t *testing.T) {
	repo := setupFirestoreTest(t)
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

func TestFirestoreUserRepository_FindByEmail(t *testing.T) {
	repo := setupFirestoreTest(t)
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

func TestFirestoreUserRepository_List(t *testing.T) {
	repo := setupFirestoreTest(t)
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
	// 他のテストとの競合があり得るため >= で検証
	if len(list) < len(users) {
		t.Errorf("List の件数 = %d, want >= %d", len(list), len(users))
	}
}

func TestFirestoreUserRepository_Update(t *testing.T) {
	repo := setupFirestoreTest(t)
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

func TestFirestoreUserRepository_Delete(t *testing.T) {
	repo := setupFirestoreTest(t)
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

func TestFirestoreUserRepository_FindByID_NotFound(t *testing.T) {
	repo := setupFirestoreTest(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "non-existent-id-999")
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
