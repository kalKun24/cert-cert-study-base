package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// NewClientFromEnv は GCS_EMULATOR_HOST 環境変数に基づいて GCS クライアントを生成します。
// GCS_EMULATOR_HOST が設定されている場合はエミュレータへ、未設定の場合は実 GCS へ接続します。
func NewClientFromEnv(ctx context.Context) (*gcs.Client, error) {
	emulatorHost := os.Getenv("GCS_EMULATOR_HOST")
	if emulatorHost != "" {
		emulatorURL, err := url.Parse(emulatorHost)
		if err != nil {
			return nil, fmt.Errorf("GCS_EMULATOR_HOST のパースに失敗しました: %w", err)
		}
		// option.WithEndpoint / STORAGE_EMULATOR_HOST は JSON API のみ対象で
		// NewReader が使うダウンロードパスがリダイレクトされない。
		// カスタム RoundTripper でリクエストのホストを書き換えることで
		// JSON API・ダウンロードを含む全リクエストをエミュレータへ向ける。
		httpClient := &http.Client{
			Transport: &emulatorRoundTripper{
				base:   http.DefaultTransport,
				scheme: emulatorURL.Scheme,
				host:   emulatorURL.Host,
			},
		}
		client, err := gcs.NewClient(ctx,
			option.WithHTTPClient(httpClient),
			option.WithoutAuthentication(),
		)
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

// emulatorRoundTripper はすべての HTTP リクエストのホストをエミュレータへ書き換えます。
type emulatorRoundTripper struct {
	base   http.RoundTripper
	scheme string
	host   string
}

func (t *emulatorRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.URL.Scheme = t.scheme
	cloned.URL.Host = t.host
	cloned.Host = t.host
	return t.base.RoundTrip(cloned)
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
