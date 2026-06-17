// Package contextkey はHTTPコンテキストに格納するキーの型と定数を定義します。
// このパッケージを共有することで、infrastructure層とinterface/handler層が
// 互いに依存せずに同一のキーを参照できます。
package contextkey

// Key はコンテキストキーの型です。
// 文字列ではなく専用型を使うことで他パッケージとの衝突を防ぎます。
type Key string

const (
	// UserID はコンテキストに格納するユーザーIDのキーです。
	UserID Key = "user_id"
	// UserRole はコンテキストに格納するユーザーロールのキーです。
	UserRole Key = "user_role"
	// IsActive はコンテキストに格納するユーザー有効状態のキーです。
	IsActive Key = "is_active"
)
