// Package auth はJWT認証・bcryptパスワード管理の具体実装を提供します。
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost はbcryptのコストパラメータです。
// OWASP推奨は12以上。パフォーマンスとセキュリティのバランスを取り12を採用します。
const bcryptCost = 12

// BcryptHasher はbcryptを使用したパスワードハッシュ化の実装です。
// usecase.PasswordHasher インターフェースを実装します。
type BcryptHasher struct{}

// NewBcryptHasher は BcryptHasher を生成します。
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

// Hash はパスワードをbcryptでハッシュ化します。
// 平文パスワードは呼び出し後に廃棄してください。
func (h *BcryptHasher) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("bcryptハッシュ化に失敗しました: %w", err)
	}
	return string(hashed), nil
}

// Verify はパスワードとbcryptハッシュが一致するか検証します。
// 一致する場合は true、一致しない場合は false を返します。
func (h *BcryptHasher) Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
