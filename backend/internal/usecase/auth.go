// Package usecase はビジネスロジック（ユースケース）を実装します。
// このパッケージは domain パッケージのみに依存します。
package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

// PasswordHasher はパスワードのハッシュ化と検証を抽象化するインターフェースです。
// 具体的な実装（bcrypt）はinfrastructure層に置きます。
type PasswordHasher interface {
	// Hash はパスワードをハッシュ化して返します。
	Hash(password string) (string, error)
	// Verify はパスワードとハッシュが一致するか検証します。
	Verify(password, hash string) bool
}

// TokenGenerator はJWTトークンの生成と検証を抽象化するインターフェースです。
// 具体的な実装はinfrastructure層に置きます。
type TokenGenerator interface {
	// Generate はユーザー情報からJWTトークンを生成します。
	Generate(user *domain.User) (string, error)
}

// AuthUseCase は認証に関するユースケースを実装します。
type AuthUseCase struct {
	userRepo domain.UserRepository
	hasher   PasswordHasher
	tokenGen TokenGenerator
}

// NewAuthUseCase は AuthUseCase を生成します（コンストラクタインジェクション）。
func NewAuthUseCase(
	userRepo domain.UserRepository,
	hasher PasswordHasher,
	tokenGen TokenGenerator,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
		hasher:   hasher,
		tokenGen: tokenGen,
	}
}

// LoginInput はログインユースケースの入力です。
type LoginInput struct {
	Username string
	Password string
}

// LoginOutput はログインユースケースの出力です。
type LoginOutput struct {
	Token string
	User  *domain.User
}

// Login はusername/パスワードで認証し、JWTトークンを返します。
// - is_active: false のユーザーは ErrUserInactive を返します。
// - 認証失敗時は ErrInvalidCredentials を返します。
func (uc *AuthUseCase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	user, err := uc.userRepo.FindByUsername(ctx, input.Username)
	if err != nil {
		// ユーザーが存在しない場合も認証失敗として扱い、情報漏洩を防ぐ
		return nil, domain.ErrInvalidCredentials
	}

	// 停止中ユーザーのログインを拒否
	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	// パスワード検証
	if !uc.hasher.Verify(input.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// JWTトークン生成
	token, err := uc.tokenGen.Generate(user)
	if err != nil {
		return nil, fmt.Errorf("JWTトークン生成に失敗しました: %w", err)
	}

	// 最終ログイン日時を更新して保存
	// 保存に失敗してもログイン自体は成功とみなし、警告ログのみ記録する
	now := time.Now().UTC()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	if err := uc.userRepo.Save(ctx, user); err != nil {
		// GCS の一時的な障害時もログインを失敗させないため、エラーは握り潰してログのみ出力する
		slog.Warn("最終ログイン日時の更新に失敗しました", "user_id", user.ID, "error", err)
	}

	return &LoginOutput{
		Token: token,
		User:  user,
	}, nil
}

// UserUseCase はユーザー管理に関するユースケースを実装します。
type UserUseCase struct {
	userRepo domain.UserRepository
	hasher   PasswordHasher
}

// NewUserUseCase は UserUseCase を生成します（コンストラクタインジェクション）。
func NewUserUseCase(userRepo domain.UserRepository, hasher PasswordHasher) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

// CreateUserInput はユーザー作成ユースケースの入力です。
type CreateUserInput struct {
	Username    string
	DisplayName string
	Email       string
	Password    string
	Role        domain.Role
}

// CreateUser は新しいユーザーを作成します（admin のみ呼び出し可）。
// - username または email の重複時は対応するドメインエラーを返します。
// - パスワードはbcryptでハッシュ化して保存します。平文は保存しません。
func (uc *UserUseCase) CreateUser(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	// ロール検証
	if !input.Role.IsValid() {
		return nil, domain.ErrInvalidRole
	}

	// username 重複チェック
	if _, err := uc.userRepo.FindByUsername(ctx, input.Username); err == nil {
		return nil, domain.ErrUsernameAlreadyExists
	}

	// email 重複チェック
	if _, err := uc.userRepo.FindByEmail(ctx, input.Email); err == nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	// パスワードをbcryptでハッシュ化（平文保存禁止）
	hash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.NewString(),
		Username:     input.Username,
		DisplayName:  input.DisplayName,
		Email:        input.Email,
		PasswordHash: hash,
		Role:         input.Role,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}

	return user, nil
}

// GetUser はIDでユーザーを取得します（admin のみ呼び出し可）。
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ユーザー取得に失敗しました: %w", err)
	}
	return user, nil
}

// ListUsers は全ユーザーを取得します（admin のみ呼び出し可）。
func (uc *UserUseCase) ListUsers(ctx context.Context) ([]*domain.User, error) {
	users, err := uc.userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("ユーザー一覧取得に失敗しました: %w", err)
	}
	return users, nil
}

// UpdateUserInput はユーザー更新ユースケースの入力です。
// 各フィールドは nil ポインタの場合は更新しません。
type UpdateUserInput struct {
	DisplayName *string
	Email       *string
	Role        *domain.Role
	Password    *string
}

// UpdateUser は指定IDのユーザー情報を更新します（admin のみ呼び出し可）。
func (uc *UserUseCase) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ユーザー取得に失敗しました: %w", err)
	}

	if input.DisplayName != nil {
		user.DisplayName = *input.DisplayName
	}

	if input.Email != nil {
		// 別ユーザーがそのemailを既に使っていないか確認
		existing, err := uc.userRepo.FindByEmail(ctx, *input.Email)
		if err == nil && existing.ID != id {
			return nil, domain.ErrEmailAlreadyExists
		}
		user.Email = *input.Email
	}

	if input.Role != nil {
		if !input.Role.IsValid() {
			return nil, domain.ErrInvalidRole
		}
		user.Role = *input.Role
	}

	if input.Password != nil {
		hash, err := uc.hasher.Hash(*input.Password)
		if err != nil {
			return nil, fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
		}
		user.PasswordHash = hash
	}

	user.UpdatedAt = time.Now().UTC()

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}

	return user, nil
}

// DeleteUser は指定IDのユーザーを削除します（admin のみ呼び出し可）。
func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	// ユーザーが存在するか確認
	if _, err := uc.userRepo.FindByID(ctx, id); err != nil {
		return fmt.Errorf("ユーザー取得に失敗しました: %w", err)
	}

	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("ユーザーの削除に失敗しました: %w", err)
	}

	return nil
}

// UpdateProfileInput はプロフィール編集ユースケースの入力です。
type UpdateProfileInput struct {
	// UserID は編集対象のユーザーID（ログイン中のユーザー自身）
	UserID string
	// DisplayName は新しい表示名
	DisplayName string
}

// UpdateProfile はログイン中のユーザーの display_name を更新します（全ロール可）。
func (uc *UserUseCase) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*domain.User, error) {
	if input.DisplayName == "" {
		return nil, fmt.Errorf("display_name は必須です")
	}

	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー取得に失敗しました: %w", err)
	}

	user.DisplayName = input.DisplayName
	user.UpdatedAt = time.Now().UTC()

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}

	return user, nil
}

// ChangePasswordInput はパスワード変更ユースケースの入力です。
type ChangePasswordInput struct {
	// UserID は変更対象のユーザーID（ログイン中のユーザー自身）
	UserID string
	// CurrentPassword は現在のパスワード（平文）
	CurrentPassword string
	// NewPassword は新しいパスワード（平文）
	NewPassword string
}

// ChangePassword はログイン中のユーザーのパスワードを変更します（全ロール可）。
// 現在のパスワードが誤っている場合は ErrCurrentPasswordIncorrect を返します。
func (uc *UserUseCase) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("ユーザー取得に失敗しました: %w", err)
	}

	// 現在のパスワードを検証
	if !uc.hasher.Verify(input.CurrentPassword, user.PasswordHash) {
		return domain.ErrCurrentPasswordIncorrect
	}

	// 新しいパスワードをbcryptでハッシュ化
	hash, err := uc.hasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now().UTC()

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}

	return nil
}

// UpdateUserStatus はユーザーの有効/停止を切り替えます（admin のみ呼び出し可）。
func (uc *UserUseCase) UpdateUserStatus(ctx context.Context, id string, isActive bool) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ユーザー取得に失敗しました: %w", err)
	}

	user.IsActive = isActive
	user.UpdatedAt = time.Now().UTC()

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}

	return user, nil
}

// UpdateTeamOwnerStatusInput はグローバルチームオーナー権限更新ユースケースの入力です。
type UpdateTeamOwnerStatusInput struct {
	// UserID は更新対象のユーザーID
	UserID string
	// IsTeamOwner はチームを作成できるグローバル権限を付与する場合は true
	IsTeamOwner bool
	// MaxTeams は作成可能なチームの上限数（0 = 制限なし）
	MaxTeams int
}

// UpdateTeamOwnerStatus はユーザーのグローバルチームオーナー権限を更新します（admin のみ呼び出し可）。
// 指定ユーザーが存在しない場合は ErrUserNotFound を返します。
func (uc *UserUseCase) UpdateTeamOwnerStatus(ctx context.Context, input UpdateTeamOwnerStatusInput) (*domain.User, error) {
	if input.MaxTeams < 0 {
		return nil, fmt.Errorf("max_teams は0以上の値を指定してください")
	}

	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー取得に失敗しました: %w", err)
	}

	user.IsTeamOwner = input.IsTeamOwner
	user.MaxTeams = input.MaxTeams
	user.UpdatedAt = time.Now().UTC()

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}

	return user, nil
}
