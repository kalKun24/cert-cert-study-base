// Package seed はサーバー起動時の初期データ投入処理を提供します。
package seed

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
)

// PasswordHasher はパスワードのハッシュ化を抽象化するインターフェースです。
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// SeedAdminIfNeeded は初回 admin ユーザーの seed を実行します。
//
// 以下の条件をすべて満たす場合のみ admin を作成します:
//   - SEED_ADMIN_USERNAME / SEED_ADMIN_PASSWORD / SEED_ADMIN_EMAIL がすべて設定されている
//   - ユーザーストアが空（既存ユーザーが 0 件）
//
// どちらかが満たされない場合はスキップします（冪等）。
func SeedAdminIfNeeded(ctx context.Context, userRepo domain.UserRepository, hasher PasswordHasher) error {
	username := os.Getenv("SEED_ADMIN_USERNAME")
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	email := os.Getenv("SEED_ADMIN_EMAIL")
	displayName := os.Getenv("SEED_ADMIN_DISPLAY_NAME")

	if username == "" || password == "" || email == "" {
		slog.Info("SEED_ADMIN_* 環境変数が未設定のため admin seed をスキップします")
		return nil
	}

	if displayName == "" {
		displayName = "Administrator"
	}

	users, err := userRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("ユーザー一覧の取得に失敗しました: %w", err)
	}

	if len(users) > 0 {
		slog.Info("ユーザーが既に存在するため admin seed をスキップします", "count", len(users))
		return nil
	}

	hash, err := hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
	}

	now := time.Now().UTC()
	admin := &domain.User{
		ID:           uuid.NewString(),
		Username:     username,
		DisplayName:  displayName,
		Email:        email,
		PasswordHash: hash,
		Role:         domain.RoleAdmin,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := userRepo.Save(ctx, admin); err != nil {
		return fmt.Errorf("admin ユーザーの保存に失敗しました: %w", err)
	}

	slog.Info("初回 admin ユーザーを作成しました", "username", username, "email", email)
	return nil
}
