package storage

import (
	"context"
	"fmt"
	"io"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCSStorageClient は StorageClient インターフェースのGCS実装です。
type GCSStorageClient struct {
	client *gcs.Client
}

// NewGCSStorageClient は GCSStorageClient を生成します。
func NewGCSStorageClient(client *gcs.Client) *GCSStorageClient {
	return &GCSStorageClient{client: client}
}

// Read はバケット内の指定オブジェクトを読み込みます。
func (c *GCSStorageClient) Read(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
	rc, err := c.client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクトの読み込みに失敗しました (bucket=%s, object=%s): %w", bucket, object, err)
	}
	return rc, nil
}

// Write はバケット内の指定オブジェクトに書き込みます。
func (c *GCSStorageClient) Write(ctx context.Context, bucket, object string, r io.Reader) error {
	wc := c.client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		_ = wc.Close()
		return fmt.Errorf("GCS オブジェクトへの書き込みに失敗しました (bucket=%s, object=%s): %w", bucket, object, err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("GCS Writer のクローズに失敗しました (bucket=%s, object=%s): %w", bucket, object, err)
	}
	return nil
}

// Delete はバケット内の指定オブジェクトを削除します。
func (c *GCSStorageClient) Delete(ctx context.Context, bucket, object string) error {
	if err := c.client.Bucket(bucket).Object(object).Delete(ctx); err != nil {
		return fmt.Errorf("GCS オブジェクトの削除に失敗しました (bucket=%s, object=%s): %w", bucket, object, err)
	}
	return nil
}

// Exists はバケット内の指定オブジェクトが存在するかを確認します。
func (c *GCSStorageClient) Exists(ctx context.Context, bucket, object string) (bool, error) {
	_, err := c.client.Bucket(bucket).Object(object).Attrs(ctx)
	if err != nil {
		if err == gcs.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("GCS オブジェクトの存在確認に失敗しました (bucket=%s, object=%s): %w", bucket, object, err)
	}
	return true, nil
}

// List はバケット内の指定プレフィックスに一致するオブジェクト名一覧を返します。
func (c *GCSStorageClient) List(ctx context.Context, bucket, prefix string) ([]string, error) {
	var names []string

	it := c.client.Bucket(bucket).Objects(ctx, &gcs.Query{Prefix: prefix})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("GCS オブジェクト一覧取得に失敗しました (bucket=%s, prefix=%s): %w", bucket, prefix, err)
		}
		names = append(names, attrs.Name)
	}

	return names, nil
}

// GCSStorageClient が StorageClient を実装していることをコンパイル時に保証します。
var _ StorageClient = (*GCSStorageClient)(nil)
