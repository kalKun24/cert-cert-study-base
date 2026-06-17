package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// NewClientFromEnv は GCS_EMULATOR_HOST 環境変数に基づいて GCS クライアントを生成します。
// GCS_EMULATOR_HOST が設定されている場合はエミュレータへ、未設定の場合は実 GCS へ接続します。
func NewClientFromEnv(ctx context.Context) (*gcs.Client, error) {
	emulatorHost := os.Getenv("GCS_EMULATOR_HOST")
	if emulatorHost != "" {
		// GCS SDK は STORAGE_EMULATOR_HOST（host:port 形式、スキームなし）を natively サポートする。
		// option.WithEndpoint は JSON API のみ対象で NewReader が使う download path が
		// リダイレクトされないため、SDK 標準の env var を経由して全 API をリダイレクトする。
		hostPort := strings.TrimPrefix(strings.TrimPrefix(emulatorHost, "https://"), "http://")
		os.Setenv("STORAGE_EMULATOR_HOST", hostPort) //nolint:errcheck
		client, err := gcs.NewClient(ctx, option.WithoutAuthentication())
		if err != nil {
			return nil, fmt.Errorf("GCS エミュレータクライアントの初期化に失敗しました: %w", err)
		}
		return client, nil
	}
	client, err := gcs.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCS クライアントの初期化に失敗しました: %w", err)
	}
	return client, nil
}

// EnsureBucketExists はバケットが存在しない場合に作成します。
// エミュレータ起動直後はバケットが空のため、初期化時に呼び出します。
func EnsureBucketExists(ctx context.Context, client *gcs.Client, bucket string) error {
	if err := client.Bucket(bucket).Create(ctx, "local", nil); err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == http.StatusConflict {
			return nil // 既に存在する
		}
		return fmt.Errorf("バケットの作成に失敗しました (bucket=%s): %w", bucket, err)
	}
	return nil
}
