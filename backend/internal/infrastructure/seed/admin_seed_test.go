package seed_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/seed"
)

// --- モック定義 ---

type mockUserRepo struct {
	users  []*domain.User
	saved  *domain.User
	listErr error
	saveErr error
}

func (m *mockUserRepo) List(_ context.Context) ([]*domain.User, error) {
	return m.users, m.listErr
}

func (m *mockUserRepo) Save(_ context.Context, u *domain.User) error {
	m.saved = u
	return m.saveErr
}

func (m *mockUserRepo) FindByID(_ context.Context, _ string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) FindByUsername(_ context.Context, _ string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) FindByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) Delete(_ context.Context, _ string) error {
	return nil
}

type mockHasher struct {
	hashResult string
	hashErr    error
}

func (m *mockHasher) Hash(_ string) (string, error) {
	return m.hashResult, m.hashErr
}

// --- テスト ---

func TestSeedAdminIfNeeded_SkipsWhenEnvVarsNotSet(t *testing.T) {
	t.Setenv("SEED_ADMIN_USERNAME", "")
	t.Setenv("SEED_ADMIN_PASSWORD", "")
	t.Setenv("SEED_ADMIN_EMAIL", "")

	repo := &mockUserRepo{}
	hasher := &mockHasher{hashResult: "hashed"}

	if err := seed.SeedAdminIfNeeded(context.Background(), repo, hasher); err != nil {
		t.Fatalf("エラーが返らないはず: %v", err)
	}
	if repo.saved != nil {
		t.Fatal("環境変数未設定時は保存されないはず")
	}
}

func TestSeedAdminIfNeeded_SkipsWhenUsersExist(t *testing.T) {
	t.Setenv("SEED_ADMIN_USERNAME", "admin")
	t.Setenv("SEED_ADMIN_PASSWORD", "password123")
	t.Setenv("SEED_ADMIN_EMAIL", "admin@example.com")

	existing := &domain.User{ID: "existing-id", Username: "existing"}
	repo := &mockUserRepo{users: []*domain.User{existing}}
	hasher := &mockHasher{hashResult: "hashed"}

	if err := seed.SeedAdminIfNeeded(context.Background(), repo, hasher); err != nil {
		t.Fatalf("エラーが返らないはず: %v", err)
	}
	if repo.saved != nil {
		t.Fatal("既存ユーザーがいる場合は保存されないはず")
	}
}

func TestSeedAdminIfNeeded_CreatesAdminWhenStoreIsEmpty(t *testing.T) {
	t.Setenv("SEED_ADMIN_USERNAME", "admin")
	t.Setenv("SEED_ADMIN_PASSWORD", "password123")
	t.Setenv("SEED_ADMIN_EMAIL", "admin@example.com")
	t.Setenv("SEED_ADMIN_DISPLAY_NAME", "Administrator")

	repo := &mockUserRepo{users: []*domain.User{}}
	hasher := &mockHasher{hashResult: "bcrypt-hashed"}

	if err := seed.SeedAdminIfNeeded(context.Background(), repo, hasher); err != nil {
		t.Fatalf("エラーが返らないはず: %v", err)
	}
	if repo.saved == nil {
		t.Fatal("admin ユーザーが保存されるはず")
	}

	u := repo.saved
	if u.Username != "admin" {
		t.Errorf("Username = %q, want %q", u.Username, "admin")
	}
	if u.Email != "admin@example.com" {
		t.Errorf("Email = %q, want %q", u.Email, "admin@example.com")
	}
	if u.Role != domain.RoleAdmin {
		t.Errorf("Role = %q, want %q", u.Role, domain.RoleAdmin)
	}
	if !u.IsActive {
		t.Error("IsActive = false, want true")
	}
	if u.PasswordHash != "bcrypt-hashed" {
		t.Errorf("PasswordHash = %q, want %q", u.PasswordHash, "bcrypt-hashed")
	}
	if u.ID == "" {
		t.Error("ID が空")
	}
}

func TestSeedAdminIfNeeded_DefaultDisplayName(t *testing.T) {
	t.Setenv("SEED_ADMIN_USERNAME", "admin")
	t.Setenv("SEED_ADMIN_PASSWORD", "password123")
	t.Setenv("SEED_ADMIN_EMAIL", "admin@example.com")
	t.Setenv("SEED_ADMIN_DISPLAY_NAME", "")

	repo := &mockUserRepo{users: []*domain.User{}}
	hasher := &mockHasher{hashResult: "hashed"}

	if err := seed.SeedAdminIfNeeded(context.Background(), repo, hasher); err != nil {
		t.Fatalf("エラーが返らないはず: %v", err)
	}
	if repo.saved.DisplayName != "Administrator" {
		t.Errorf("DisplayName = %q, want %q", repo.saved.DisplayName, "Administrator")
	}
}

func TestSeedAdminIfNeeded_ReturnsErrorOnHashFailure(t *testing.T) {
	t.Setenv("SEED_ADMIN_USERNAME", "admin")
	t.Setenv("SEED_ADMIN_PASSWORD", "password123")
	t.Setenv("SEED_ADMIN_EMAIL", "admin@example.com")

	repo := &mockUserRepo{users: []*domain.User{}}
	hasher := &mockHasher{hashErr: errors.New("hash error")}

	if err := seed.SeedAdminIfNeeded(context.Background(), repo, hasher); err == nil {
		t.Fatal("ハッシュエラー時はエラーが返るはず")
	}
}

func TestSeedAdminIfNeeded_ReturnsErrorOnSaveFailure(t *testing.T) {
	t.Setenv("SEED_ADMIN_USERNAME", "admin")
	t.Setenv("SEED_ADMIN_PASSWORD", "password123")
	t.Setenv("SEED_ADMIN_EMAIL", "admin@example.com")

	repo := &mockUserRepo{users: []*domain.User{}, saveErr: errors.New("save error")}
	hasher := &mockHasher{hashResult: "hashed"}

	if err := seed.SeedAdminIfNeeded(context.Background(), repo, hasher); err == nil {
		t.Fatal("保存エラー時はエラーが返るはず")
	}
}
