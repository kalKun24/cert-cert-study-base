# バックログ優先順位

着手推奨順。依存関係を考慮した順番で並べている。
ステータス管理はファイルのディレクトリ（backlog/ → in-progress/ → done/）で行う。

---

## P0 — ないと動かない（最優先）

| 順位 | チケット | 理由 |
|---|---|---|
| 1 | [TICKET-017 初回 admin ユーザーセットアップ](backlog/TICKET-017_初回adminユーザーセットアップ.md) | 誰もログインできない状態を解消する |
| 2 | [TICKET-016 GCS ローカル開発エミュレータ](backlog/TICKET-016_GCSローカル開発エミュレータ.md) | 実 GCS なしでローカル開発・CI テストができる環境を整備する |
| 3 | [TICKET-011 フロントエンド認証基盤](backlog/TICKET-011_フロントエンド認証基盤.md) | ログイン画面と JWT 管理がないとフロントエンド開発が始められない |

---

## P1 — バックエンド機能実装

| 順位 | チケット | 依存 |
|---|---|---|
| 4 | [TICKET-003 チーム管理機能実装](backlog/TICKET-003_チーム管理機能実装.md) | TICKET-002（完了済み） |
| 5 | [TICKET-004 問題 CRUD API 実装](backlog/TICKET-004_問題CRUD_API実装.md) | TICKET-002（完了済み） |
| 6 | [TICKET-005 タグ管理 API 実装](backlog/TICKET-005_タグ管理API実装.md) | TICKET-004 |
| 7 | [TICKET-006 問題公開設定・下書き機能実装](backlog/TICKET-006_問題公開設定・下書き機能実装.md) | TICKET-003, 004 |
| 8 | [TICKET-007 問題コメント機能実装](backlog/TICKET-007_問題コメント機能実装.md) | TICKET-004 |
| 9 | [TICKET-008 タグ検索・フィルタリング API 実装](backlog/TICKET-008_タグ検索・フィルタリングAPI実装.md) | TICKET-005 |
| 10 | [TICKET-009 問題管理フロントエンド実装](backlog/TICKET-009_問題管理フロントエンド実装.md) | TICKET-004, 011 |

---

## P2 — インフラ整備

| 順位 | チケット | 依存 |
|---|---|---|
| 11 | [TICKET-010 GCP インフラ構築（Terraform）](backlog/TICKET-010_GCPインフラ構築_Terraform.md) | GCP プロジェクト作成（人間作業）|
| 12 | [TICKET-015 CD パイプライン（Cloud Run 自動デプロイ）](backlog/TICKET-015_CDパイプライン.md) | TICKET-010 |

---

## P3 — フロントエンド完成

| 順位 | チケット | 依存 |
|---|---|---|
| 13 | [TICKET-014 ユーザー管理フロントエンド](backlog/TICKET-014_ユーザー管理フロントエンド.md) | TICKET-011（API は TICKET-002 完了済み）|
| 14 | [TICKET-012 チーム管理フロントエンド](backlog/TICKET-012_チーム管理フロントエンド.md) | TICKET-003, 011 |
| 15 | [TICKET-013 タグ検索・フィルタフロントエンド](backlog/TICKET-013_タグ検索フィルタフロントエンド.md) | TICKET-005, 008, 009, 011 |

---

## 完了済み

| チケット |
|---|
| [TICKET-001 プロジェクト基盤構築](done/TICKET-001_プロジェクト基盤構築.md) |
| [TICKET-002 認証 API 実装](done/TICKET-002_認証API実装.md) |
