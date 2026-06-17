package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/usecase"
)

// --- モック定義 ---

// mockUserRepository は domain.UserRepository のモックです。
type mockUserRepository struct {
	users       map[string]*domain.User
	byUsername  map[string]*domain.User
	byEmail     map[string]*domain.User
	saveErr     error
	deleteErr   error
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:      make(map[string]*domain.User),
		byUsername: make(map[string]*domain.User),
		byEmail:    make(map[string]*domain.User),
	}
}

// addUser はテスト用のユーザーをモックに追加します。
func (m *mockUserRepository) addUser(u *domain.User) {
	m.users[u.ID] = u
	m.byUsername[u.Username] = u
	m.byEmail[u.Email] = u
}

func (m *mockUserRepository) FindByID(id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) FindByUsername(username string) (*domain.User, error) {
	if u, ok := m.byUsername[username]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) FindByEmail(email string) (*domain.User, error) {
	if u, ok := m.byEmail[email]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) List() ([]*domain.User, error) {
	users := make([]*domain.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

func (m *mockUserRepository) Save(user *domain.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.users[user.ID] = user
	m.byUsername[user.Username] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *mockUserRepository) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if u, ok := m.users[id]; ok {
		delete(m.users, id)
		delete(m.byUsername, u.Username)
		delete(m.byEmail, u.Email)
		return nil
	}
	return domain.ErrUserNotFound
}

// mockPasswordHasher は usecase.PasswordHasher のモックです。
type mockPasswordHasher struct {
	hashResult  string
	hashErr     error
	verifyResult bool
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	if m.hashErr != nil {
		return "", m.hashErr
	}
	if m.hashResult != "" {
		return m.hashResult, nil
	}
	return "hashed:" + password, nil
}

func (m *mockPasswordHasher) Verify(password, hash string) bool {
	return m.verifyResult
}

// mockTokenGenerator は usecase.TokenGenerator のモックです。
type mockTokenGenerator struct {
	token    string
	tokenErr error
}

func (m *mockTokenGenerator) Generate(user *domain.User) (string, error) {
	if m.tokenErr != nil {
		return "", m.tokenErr
	}
	if m.token != "" {
		return m.token, nil
	}
	return "token:" + user.ID, nil
}

// --- テストヘルパー ---

// testUser はテスト用のユーザーエンティティを生成します。
func testUser(id, username, email string, role domain.Role, isActive bool) *domain.User {
	return &domain.User{
		ID:           id,
		Username:     username,
		DisplayName:  "テストユーザー",
		Email:        email,
		PasswordHash: "hashed:password123",
		Role:         role,
		IsActive:     isActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// --- AuthUseCase のテスト ---

func TestAuthUseCase_Login_Success(t *testing.T) {
	repo := newMockUserRepository()
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(user)

	hasher := &mockPasswordHasher{verifyResult: true}
	tokenGen := &mockTokenGenerator{token: "jwt-token"}
	uc := usecase.NewAuthUseCase(repo, hasher, tokenGen)

	out, err := uc.Login(usecase.LoginInput{
		Username: "alice",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("ログインに失敗しました: %v", err)
	}
	if out.Token != "jwt-token" {
		t.Errorf("トークンが期待値と異なります: got %s, want jwt-token", out.Token)
	}
	if out.User.ID != "id-1" {
		t.Errorf("ユーザーIDが期待値と異なります: got %s, want id-1", out.User.ID)
	}
}

func TestAuthUseCase_Login_UserNotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}
	uc := usecase.NewAuthUseCase(repo, hasher, tokenGen)

	_, err := uc.Login(usecase.LoginInput{
		Username: "nonexistent",
		Password: "password123",
	})

	// ユーザー未存在は ErrInvalidCredentials を返す（情報漏洩防止）
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidCredentials)
	}
}

func TestAuthUseCase_Login_InvalidPassword(t *testing.T) {
	repo := newMockUserRepository()
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(user)

	hasher := &mockPasswordHasher{verifyResult: false} // パスワード不一致
	tokenGen := &mockTokenGenerator{}
	uc := usecase.NewAuthUseCase(repo, hasher, tokenGen)

	_, err := uc.Login(usecase.LoginInput{
		Username: "alice",
		Password: "wrongpassword",
	})

	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidCredentials)
	}
}

func TestAuthUseCase_Login_InactiveUser(t *testing.T) {
	repo := newMockUserRepository()
	// 停止中ユーザー
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, false)
	repo.addUser(user)

	hasher := &mockPasswordHasher{verifyResult: true}
	tokenGen := &mockTokenGenerator{}
	uc := usecase.NewAuthUseCase(repo, hasher, tokenGen)

	_, err := uc.Login(usecase.LoginInput{
		Username: "alice",
		Password: "password123",
	})

	if !errors.Is(err, domain.ErrUserInactive) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrUserInactive)
	}
}

func TestAuthUseCase_Login_TokenGenerationError(t *testing.T) {
	repo := newMockUserRepository()
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(user)

	hasher := &mockPasswordHasher{verifyResult: true}
	tokenGen := &mockTokenGenerator{tokenErr: errors.New("JWT生成エラー")}
	uc := usecase.NewAuthUseCase(repo, hasher, tokenGen)

	_, err := uc.Login(usecase.LoginInput{
		Username: "alice",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("エラーが返されませんでした")
	}
}

// --- UserUseCase のテスト ---

func TestUserUseCase_CreateUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	user, err := uc.CreateUser(usecase.CreateUserInput{
		Username:    "bob",
		DisplayName: "Bob",
		Email:       "bob@example.com",
		Password:    "password123",
		Role:        domain.RoleUser,
	})

	if err != nil {
		t.Fatalf("ユーザー作成に失敗しました: %v", err)
	}
	if user.Username != "bob" {
		t.Errorf("usernameが期待値と異なります: got %s, want bob", user.Username)
	}
	if user.IsActive != true {
		t.Errorf("is_activeが期待値と異なります: got %v, want true", user.IsActive)
	}
	// パスワードがハッシュ化されていること（平文でないこと）を確認
	if user.PasswordHash == "password123" {
		t.Error("パスワードが平文で保存されています。ハッシュ化が必要です")
	}
}

func TestUserUseCase_CreateUser_DuplicateUsername(t *testing.T) {
	repo := newMockUserRepository()
	existing := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(existing)

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	_, err := uc.CreateUser(usecase.CreateUserInput{
		Username:    "alice", // 既存のusername
		DisplayName: "Alice2",
		Email:       "alice2@example.com",
		Password:    "password123",
		Role:        domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrUsernameAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrUsernameAlreadyExists)
	}
}

func TestUserUseCase_CreateUser_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepository()
	existing := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(existing)

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	_, err := uc.CreateUser(usecase.CreateUserInput{
		Username:    "alice2", // 別のusername
		DisplayName: "Alice2",
		Email:       "alice@example.com", // 既存のemail
		Password:    "password123",
		Role:        domain.RoleUser,
	})

	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrEmailAlreadyExists)
	}
}

func TestUserUseCase_CreateUser_InvalidRole(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	_, err := uc.CreateUser(usecase.CreateUserInput{
		Username:    "bob",
		DisplayName: "Bob",
		Email:       "bob@example.com",
		Password:    "password123",
		Role:        domain.Role("invalid-role"),
	})

	if !errors.Is(err, domain.ErrInvalidRole) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidRole)
	}
}

func TestUserUseCase_GetUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleAdmin, true)
	repo.addUser(user)

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	got, err := uc.GetUser("id-1")
	if err != nil {
		t.Fatalf("ユーザー取得に失敗しました: %v", err)
	}
	if got.ID != "id-1" {
		t.Errorf("IDが期待値と異なります: got %s, want id-1", got.ID)
	}
}

func TestUserUseCase_GetUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	_, err := uc.GetUser("nonexistent-id")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestUserUseCase_ListUsers_Success(t *testing.T) {
	repo := newMockUserRepository()
	repo.addUser(testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true))
	repo.addUser(testUser("id-2", "bob", "bob@example.com", domain.RoleTeamOwner, true))

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	users, err := uc.ListUsers()
	if err != nil {
		t.Fatalf("ユーザー一覧取得に失敗しました: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("取得件数が期待値と異なります: got %d, want 2", len(users))
	}
}

func TestUserUseCase_UpdateUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(user)

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	newDisplayName := "Alice Updated"
	got, err := uc.UpdateUser("id-1", usecase.UpdateUserInput{
		DisplayName: &newDisplayName,
	})

	if err != nil {
		t.Fatalf("ユーザー更新に失敗しました: %v", err)
	}
	if got.DisplayName != "Alice Updated" {
		t.Errorf("DisplayNameが期待値と異なります: got %s, want Alice Updated", got.DisplayName)
	}
}

func TestUserUseCase_UpdateUser_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepository()
	repo.addUser(testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true))
	repo.addUser(testUser("id-2", "bob", "bob@example.com", domain.RoleUser, true))

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	// bob のメールを alice と同じにしようとする
	aliceEmail := "alice@example.com"
	_, err := uc.UpdateUser("id-2", usecase.UpdateUserInput{
		Email: &aliceEmail,
	})

	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrEmailAlreadyExists)
	}
}

func TestUserUseCase_UpdateUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	name := "New Name"
	_, err := uc.UpdateUser("nonexistent-id", usecase.UpdateUserInput{
		DisplayName: &name,
	})

	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestUserUseCase_UpdateUser_InvalidRole(t *testing.T) {
	repo := newMockUserRepository()
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(user)

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	invalidRole := domain.Role("superuser")
	_, err := uc.UpdateUser("id-1", usecase.UpdateUserInput{
		Role: &invalidRole,
	})

	if !errors.Is(err, domain.ErrInvalidRole) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrInvalidRole)
	}
}

func TestUserUseCase_DeleteUser_Success(t *testing.T) {
	repo := newMockUserRepository()
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(user)

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	if err := uc.DeleteUser("id-1"); err != nil {
		t.Fatalf("ユーザー削除に失敗しました: %v", err)
	}

	// 削除後に取得しようとすると NotFound になること
	_, err := repo.FindByID("id-1")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("削除後もユーザーが存在しています")
	}
}

func TestUserUseCase_DeleteUser_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	err := uc.DeleteUser("nonexistent-id")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrUserNotFound)
	}
}

func TestUserUseCase_UpdateUserStatus_Deactivate(t *testing.T) {
	repo := newMockUserRepository()
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, true)
	repo.addUser(user)

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	got, err := uc.UpdateUserStatus("id-1", false)
	if err != nil {
		t.Fatalf("ユーザーステータス更新に失敗しました: %v", err)
	}
	if got.IsActive {
		t.Error("is_activeが true のままです。false に変わっているはずです")
	}
}

func TestUserUseCase_UpdateUserStatus_Reactivate(t *testing.T) {
	repo := newMockUserRepository()
	// 停止中ユーザーを再有効化
	user := testUser("id-1", "alice", "alice@example.com", domain.RoleUser, false)
	repo.addUser(user)

	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	got, err := uc.UpdateUserStatus("id-1", true)
	if err != nil {
		t.Fatalf("ユーザーステータス更新に失敗しました: %v", err)
	}
	if !got.IsActive {
		t.Error("is_activeが false のままです。true に変わっているはずです")
	}
}

func TestUserUseCase_UpdateUserStatus_NotFound(t *testing.T) {
	repo := newMockUserRepository()
	hasher := &mockPasswordHasher{}
	uc := usecase.NewUserUseCase(repo, hasher)

	_, err := uc.UpdateUserStatus("nonexistent-id", false)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("エラーが期待値と異なります: got %v, want %v", err, domain.ErrUserNotFound)
	}
}
