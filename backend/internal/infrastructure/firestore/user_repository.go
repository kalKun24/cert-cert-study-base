// Package firestore はドメインリポジトリインターフェースの Cloud Firestore 実装を提供します。
package firestore

import (
	"context"
	"fmt"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// userRecord はFirestoreに保存するユーザーレコードです。
type userRecord struct {
	ID           string     `firestore:"id"`
	Username     string     `firestore:"username"`
	DisplayName  string     `firestore:"display_name"`
	Email        string     `firestore:"email"`
	PasswordHash string     `firestore:"password_hash"`
	Role         string     `firestore:"role"`
	IsActive     bool       `firestore:"is_active"`
	IsTeamOwner  bool       `firestore:"is_team_owner"`
	MaxTeams     int        `firestore:"max_teams"`
	CreatedAt    time.Time  `firestore:"created_at"`
	UpdatedAt    time.Time  `firestore:"updated_at"`
	LastLoginAt  *time.Time `firestore:"last_login_at"`
}

func toUserRecord(u *domain.User) userRecord {
	return userRecord{
		ID:           u.ID,
		Username:     u.Username,
		DisplayName:  u.DisplayName,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         string(u.Role),
		IsActive:     u.IsActive,
		IsTeamOwner:  u.IsTeamOwner,
		MaxTeams:     u.MaxTeams,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		LastLoginAt:  u.LastLoginAt,
	}
}

func toUser(r userRecord) *domain.User {
	return &domain.User{
		ID:           r.ID,
		Username:     r.Username,
		DisplayName:  r.DisplayName,
		Email:        r.Email,
		PasswordHash: r.PasswordHash,
		Role:         domain.Role(r.Role),
		IsActive:     r.IsActive,
		IsTeamOwner:  r.IsTeamOwner,
		MaxTeams:     r.MaxTeams,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		LastLoginAt:  r.LastLoginAt,
	}
}

// FirestoreUserRepository はCloud Firestoreにユーザーデータを永続化するリポジトリです。
// domain.UserRepository インターフェースを実装します。
// コレクション: users/{userID}
type FirestoreUserRepository struct {
	client *fs.Client
}

// NewFirestoreUserRepository は FirestoreUserRepository を生成します。
func NewFirestoreUserRepository(client *fs.Client) *FirestoreUserRepository {
	return &FirestoreUserRepository{client: client}
}

// FindByID はIDでユーザーを検索します。
func (r *FirestoreUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	doc, err := r.client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("ユーザーの取得に失敗しました: %w", err)
	}

	var rec userRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("ユーザーデータのデコードに失敗しました: %w", err)
	}

	return toUser(rec), nil
}

// FindByUsername はusernameでユーザーを検索します。
func (r *FirestoreUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	iter := r.client.Collection("users").Where("username", "==", username).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ユーザー検索に失敗しました: %w", err)
	}

	var rec userRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("ユーザーデータのデコードに失敗しました: %w", err)
	}

	return toUser(rec), nil
}

// FindByEmail はメールアドレスでユーザーを検索します。
func (r *FirestoreUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	iter := r.client.Collection("users").Where("email", "==", email).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("ユーザー検索に失敗しました: %w", err)
	}

	var rec userRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("ユーザーデータのデコードに失敗しました: %w", err)
	}

	return toUser(rec), nil
}

// List は全ユーザーの一覧を返します。
func (r *FirestoreUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	iter := r.client.Collection("users").Documents(ctx)
	defer iter.Stop()

	var users []*domain.User
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ユーザー一覧の取得に失敗しました: %w", err)
		}

		var rec userRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("ユーザーデータのデコードに失敗しました: %w", err)
		}
		users = append(users, toUser(rec))
	}

	if users == nil {
		users = []*domain.User{}
	}
	return users, nil
}

// Save はユーザーを新規作成または更新します（Upsert）。
func (r *FirestoreUserRepository) Save(ctx context.Context, user *domain.User) error {
	rec := toUserRecord(user)
	_, err := r.client.Collection("users").Doc(user.ID).Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("ユーザーの保存に失敗しました: %w", err)
	}
	return nil
}

// Delete はIDで指定したユーザーを削除します。
func (r *FirestoreUserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection("users").Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("ユーザーの存在確認に失敗しました: %w", err)
	}

	_, err = r.client.Collection("users").Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("ユーザーの削除に失敗しました: %w", err)
	}
	return nil
}

// FirestoreUserRepository が domain.UserRepository を実装していることをコンパイル時に保証します。
var _ domain.UserRepository = (*FirestoreUserRepository)(nil)
