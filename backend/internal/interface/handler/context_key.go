package handler

import "github.com/kalKun24/cert-study-base/backend/internal/contextkey"

// コンテキストキーの定数は contextkey パッケージで一元管理します。
// infrastructure/auth パッケージと handler パッケージが同一キーを
// 互いに依存せず参照できるようにするためです。
var (
	// ContextKeyUserID はコンテキストに格納するユーザーIDのキーです。
	ContextKeyUserID = contextkey.UserID
	// ContextKeyUserRole はコンテキストに格納するユーザーロールのキーです。
	ContextKeyUserRole = contextkey.UserRole
	// ContextKeyIsActive はコンテキストに格納するユーザー有効状態のキーです。
	ContextKeyIsActive = contextkey.IsActive
)
