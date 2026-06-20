// Package repository はドメインリポジトリインターフェースの具体実装を提供します。
package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/storage"
)

// usersObjectName はGCSバケット内でユーザーデータを保存するオブジェクト名です。
const usersObjectName = "users.json"

// userRecord はGCS上のJSONファイルに保存するユーザーレコードです。
// domain.User と対応しており、JSON直列化のための構造体です。
type userRecord struct {
	ID           string     `json:"id"`
	Username     string     `json:"username"`
	DisplayName  string     `json:"display_name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"password_hash"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	IsTeamOwner  bool       `json:"is_team_owner"`
	MaxTeams     int        `json:"max_teams"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// toUserRecord はドメインエンティティをJSONレコードに変換します。
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

// toUser はJSONレコードをドメインエンティティに変換します。
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

// GCSUserRepository はGCS上のJSONファイルにユーザーデータを永続化するリポジトリです。
// domain.UserRepository インターフェースを実装します。
// ユーザーデータは GCS バケット内の users.json に保存します。
type GCSUserRepository struct {
	mu      sync.RWMutex
	storage storage.StorageClient
	bucket  string
}

// NewGCSUserRepository は GCSUserRepository を生成します。
func NewGCSUserRepository(sc storage.StorageClient, bucket string) *GCSUserRepository {
	return &GCSUserRepository{
		storage: sc,
		bucket:  bucket,
	}
}

// load はGCSからユーザーデータを読み込みます。
// オブジェクトが存在しない場合は空のスライスを返します。
func (r *GCSUserRepository) load(ctx context.Context) ([]userRecord, error) {
	exists, err := r.storage.Exists(ctx, r.bucket, usersObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}

	if !exists {
		return []userRecord{}, nil
	}

	rc, err := r.storage.Read(ctx, r.bucket, usersObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS からの読み込みに失敗しました: %w", err)
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			slog.Warn("GCS ReadCloser のクローズに失敗しました", "error", cerr)
		}
	}()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("GCS データの読み取りに失敗しました: %w", err)
	}

	var records []userRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("ユーザーデータのJSONデコードに失敗しました: %w", err)
	}

	return records, nil
}

// save はユーザーデータをGCSに書き込みます。
func (r *GCSUserRepository) save(ctx context.Context, records []userRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("ユーザーデータのJSONエンコードに失敗しました: %w", err)
	}

	if err := r.storage.Write(ctx, r.bucket, usersObjectName, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}

	return nil
}

// FindByID はIDでユーザーを検索します。
func (r *GCSUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.load(ctx)
	if err != nil {
		return nil, fmt.Errorf("ユーザーデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.ID == id {
			return toUser(rec), nil
		}
	}

	return nil, domain.ErrUserNotFound
}

// FindByUsername はusernameでユーザーを検索します。
func (r *GCSUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.load(ctx)
	if err != nil {
		return nil, fmt.Errorf("ユーザーデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.Username == username {
			return toUser(rec), nil
		}
	}

	return nil, domain.ErrUserNotFound
}

// FindByEmail はメールアドレスでユーザーを検索します。
func (r *GCSUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.load(ctx)
	if err != nil {
		return nil, fmt.Errorf("ユーザーデータ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.Email == email {
			return toUser(rec), nil
		}
	}

	return nil, domain.ErrUserNotFound
}

// List は全ユーザーを返します。
func (r *GCSUserRepository) List(ctx context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.load(ctx)
	if err != nil {
		return nil, fmt.Errorf("ユーザーデータ読み込みに失敗しました: %w", err)
	}

	users := make([]*domain.User, 0, len(records))
	for _, rec := range records {
		users = append(users, toUser(rec))
	}

	return users, nil
}

// Save はユーザーを新規作成または更新します。
// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
func (r *GCSUserRepository) Save(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.load(ctx)
	if err != nil {
		return fmt.Errorf("ユーザーデータ読み込みに失敗しました: %w", err)
	}

	rec := toUserRecord(user)
	updated := false
	for i, existing := range records {
		if existing.ID == user.ID {
			records[i] = rec
			updated = true
			break
		}
	}

	if !updated {
		records = append(records, rec)
	}

	if err := r.save(ctx, records); err != nil {
		return fmt.Errorf("ユーザーデータ保存に失敗しました: %w", err)
	}

	return nil
}

// Delete はIDで指定したユーザーを削除します。
func (r *GCSUserRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.load(ctx)
	if err != nil {
		return fmt.Errorf("ユーザーデータ読み込みに失敗しました: %w", err)
	}

	newRecords := make([]userRecord, 0, len(records))
	found := false
	for _, rec := range records {
		if rec.ID == id {
			found = true
			continue
		}
		newRecords = append(newRecords, rec)
	}

	if !found {
		return domain.ErrUserNotFound
	}

	if err := r.save(ctx, newRecords); err != nil {
		return fmt.Errorf("ユーザーデータ保存に失敗しました: %w", err)
	}

	return nil
}

// GCSUserRepository が domain.UserRepository を実装していることをコンパイル時に保証します。
var _ domain.UserRepository = (*GCSUserRepository)(nil)
