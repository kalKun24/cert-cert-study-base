# TICKET-020 フロントエンドユニットテスト整備

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-020 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/frontend-unit-tests` |
| PR番号 | - |
| PRリンク | （PR作成後に記入） |

---

## 概要

バックエンドはユースケース層のユニットテストが充実しているが、フロントエンドにはテスト基盤が存在しない。Vitest + React Testing Library でテスト環境を整備し、主要コンポーネント・フックのユニットテストを追加する。

---

## 背景・目的

フロントエンドのコンポーネントは手動確認のみで担保されており、リグレッションを検知できない。特に認証フロー・API クライアント・権限制御はロジックが複雑なため自動テストが有効。

---

## 受け入れ条件

- [ ] `npm test` でテストが実行できる
- [ ] Vitest + React Testing Library のテスト基盤が整備されている
- [ ] `AuthContext`・`PrivateRoute`・`CommentSection`・`apiClient` のテストが存在する
- [ ] CI（GitHub Actions）でフロントエンドテストが自動実行される

---

## サブチケット（コミット単位）

- [ ] `chore(frontend): Vitest + React Testing Library のテスト環境を構築`
- [ ] `test(frontend): AuthContext・PrivateRoute のユニットテストを作成`
- [ ] `test(frontend): CommentSection・apiClient のユニットテストを作成`
- [ ] `chore(ci): GitHub Actions にフロントエンドテストを追加`

---

## 関連情報

- 関連チケット: TICKET-007（コメント機能）、TICKET-009（問題管理フロントエンド）、TICKET-011（フロントエンド認証基盤）
- 備考: E2E テスト（Playwright 等）は別チケットで検討。本チケットはユニットテストの基盤整備にスコープを絞る
