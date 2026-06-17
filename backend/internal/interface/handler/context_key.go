package handler

// contextKey はコンテキストキーの型です。
// 文字列ではなく専用型を使うことで他パッケージとの衝突を防ぎます。
type contextKey string

const (
	// ContextKeyUserID はコンテキストに格納するユーザーIDのキーです。
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyUserRole はコンテキストに格納するユーザーロールのキーです。
	ContextKeyUserRole contextKey = "user_role"
	// ContextKeyIsActive はコンテキストに格納するユーザー有効状態のキーです。
	ContextKeyIsActive contextKey = "is_active"
)
