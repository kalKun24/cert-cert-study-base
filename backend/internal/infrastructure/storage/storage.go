// Package storage はGCSクライアントのDIインターフェースを定義します。
// 具体的な実装（GCS接続・ローカルフォールバック）は本チケットのスコープ外です。
package storage

import (
	"context"
	"io"
)

// StorageClient はオブジェクトストレージに対する操作を抽象化するインターフェースです。
// GCS・ローカルファイルシステム・フェイクGCSなど、実装を差し替え可能にします。
type StorageClient interface {
	// Read はバケット内の指定オブジェクトを読み込みます。
	Read(ctx context.Context, bucket, object string) (io.ReadCloser, error)

	// Write はバケット内の指定オブジェクトに書き込みます。
	Write(ctx context.Context, bucket, object string, r io.Reader) error

	// Delete はバケット内の指定オブジェクトを削除します。
	Delete(ctx context.Context, bucket, object string) error

	// Exists はバケット内の指定オブジェクトが存在するかを確認します。
	Exists(ctx context.Context, bucket, object string) (bool, error)

	// List はバケット内の指定プレフィックスに一致するオブジェクト名一覧を返します。
	List(ctx context.Context, bucket, prefix string) ([]string, error)
}
