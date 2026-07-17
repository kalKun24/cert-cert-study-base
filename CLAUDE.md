# CLAUDE.md

このファイルはClaude（AIアシスタント）がこのリポジトリで作業する際のガイドラインを定義します。

---

## プロジェクト概要

**cert-study-base**

> リポジトリ: https://github.com/kalKun24/cert-study-base

CISSPや情報処理安全確保支援士などのセキュリティ資格取得を目標とした勉強会を、効率的に運営・管理するためのWebアプリケーションです。

主な機能: Markdown形式での"問題・解答・解説・議論点メモ(議論点・知らない知識などをメモする)"の作成・編集と、そのmarkdownテキストの共有、タグ（フラット・複数付与可）による分類・検索・フィルタリング。チーム単位での運用（メンバー招待・チーム切り替え）、ノート・コメント機能を含む。

主要エンティティ: `question` / `tag` / `team` / `user` / `invitation` / `note` / `comment`（`backend/internal/domain/` 参照）

---

## 技術スタック

| レイヤー | 技術 |
|---|---|
| バックエンド | Go 1.25 |
| フロントエンド | React 18 + TypeScript + Vite（状態管理: React Context） |
| データ形式 | JSON / Markdown |
| API設計 | REST API / Swagger (OpenAPI 3.0) |
| 永続化 | Cloud Firestore（全エンティティ。※`infrastructure/storage/` の GCS クライアントはインターフェース定義のみで現状未使用） |
| インフラ | GCP Cloud Run（リージョン: asia-northeast1） |
| CI/CD | GitHub Actions（main / develop マージトリガー、2環境デプロイ） |

---

## ディレクトリ構成

```
.
├── backend/internal/
│   ├── domain/          # エンティティ層: ビジネスエンティティ・ルール（依存なし）
│   ├── usecase/         # ユースケース層: ビジネスロジック（domain のみ依存可）
│   ├── interface/       # インターフェース層: ハンドラ・DTO・Repositoryインターフェース
│   ├── infrastructure/  # インフラ層: Firestore・認証・ミドルウェアの具体実装
│   └── contextkey/      # context.Context のキー定義（層をまたぐ共有定数）
├── frontend/src/
├── api/openapi.yaml     # OpenAPI 3.0 仕様書（API定義の単一管理元）
├── docs/                # ドキュメント・手順書
├── scripts/             # セットアップ・運用スクリプト
├── tickets/             # チケット管理（backlog/ in-progress/ done/）
├── .github/workflows/
└── CLAUDE.md
```

---

## 開発ガイドライン

### クリーンアーキテクチャ

依存の方向は **外側 → 内側のみ** 許可。逆方向の依存は禁止。

```
[Infrastructure] → [Interface] → [Usecase] → [Domain]
```

- 層の境界はinterfaceで定義し、具体実装はInfrastructure層に置く（DI）
- エラーハンドリングは `fmt.Errorf("...: %w", err)` で文脈を付与する

### REST API

- エンドポイントは `/api/v1/` プレフィックス、リソース名は複数形・名詞
- レスポンスは `{ "data": ..., "error": ... }` の統一フォーマット（`error` は文字列 / null）
- 一覧APIのページネーションはオフセット方式（`page` / `per_page`、レスポンスに `total_pages` を含む）
- HTTPステータスコードを適切に使用（200, 201, 400, 404, 500 など）

### Swagger（OpenAPI 3.0）

- **API定義は `api/openapi.yaml` で一元管理**（手動編集で管理する）
- 新しいエンドポイントは **Swagger定義を先に更新してから実装**（API First原則）
- Swagger UI の提供（`make swagger`）は未実装・将来対応

### 認証

- 認証方式: ID / パスワード認証（JWTトークン発行）
- ユーザデータは Firestore に永続化する
- パスワードは **bcrypt** でハッシュ化して保存。平文保存は絶対禁止
- ロール: `admin`（ユーザ管理・全機能） / `user`（自身の認証情報変更・問題CRUD）
- 認証・認可はInfrastructure層の `auth/` にミドルウェアとして実装

### ログ

- 形式: **構造化ログ（JSON）**、ライブラリ: `log/slog`（Go標準、1.21+）
- フィールド: `severity` / `time` / `message` / `request_id` / `user_id` / `method` / `path` / `status` / `latency_ms`
- レベル: `DEBUG` / `INFO` / `WARN` / `ERROR`（本番は `INFO` 以上）

### ローカル開発環境

`make` コマンドで統一（`Makefile` はリポジトリルートに配置）。

| コマンド | 内容 |
|---|---|
| `make up` | バックエンド・フロントエンドをまとめて起動（Docker Compose） |
| `make down` | 全サービスを停止 |
| `make test` | 全テストを実行 |
| `make fmt` | `gofmt` でバックエンドのコードをフォーマット |
| `make lint` | `golangci-lint` を実行 |
| `make hooks` | gitフック（`.githooks/`）を有効化。初回クローン後に一度実行 |
| `make swagger` | Swagger UIを起動・`openapi.yaml` を反映（将来実装） |
| `make build` | 本番用Dockerイメージをビルド |

Firestore を使う統合テストはエミュレータ（`FIRESTORE_EMULATOR_HOST`）を前提とする。

---

## ブランチ戦略

3層構成（`main` / `develop` / `feature/*`）。`main`・`develop` への直接pushは禁止。

- `feature/*` ブランチで開発し、`develop` へ PR を作成してマージする（dev 環境へデプロイ）
- `develop` → `main` への PR をマージすると prod 環境へデプロイされる
- PRは承認者（リポジトリオーナー）による **1名承認が必須**
- セルフマージ禁止。PRタイトルはコミットメッセージ規約に準拠

---

## CI/CD

| ブランチ | トリガー | 動作 | 環境 |
|---|---|---|---|
| `feature/*` | PR作成/更新 | テスト・Lint・セキュリティスキャン（ci.yml） | - |
| `develop` | マージ | CI 通過後にビルド → Cloud Run デプロイ（cd-dev.yml） | dev（`cert-study-base-dev`） |
| `main` | マージ | CI 通過後にビルド → Cloud Run デプロイ（cd.yml） | prod（`cert-study-base`） |

シークレットは GCP Secret Manager で管理。GCP 認証は Workload Identity Federation を使用。

---

## コーディング規約

### コミットメッセージ

[Semantic Commit Messages](https://gist.github.com/joshbuchea/6f47e86d2510bce28f8e7f42ae84c716) 準拠。**件名は日本語**で記述。

```
<type>(<scope>): <件名>
```

| type | 用途 |
|---|---|
| `feat` | 新機能追加 |
| `fix` | バグ修正 |
| `docs` | ドキュメント変更 |
| `style` | フォーマット修正（動作に影響しない） |
| `refactor` | リファクタリング |
| `test` | テストの追加・修正 |
| `chore` | ビルドツール・タスクランナーの更新 |

例: `feat(question): 問題作成APIを追加`

### Go
- `gofmt` / `golangci-lint` を通すこと
- テストファイルは `_test.go` サフィックスで同パッケージに配置

### React
- 関数コンポーネント + Hooks で統一、ESLint / Prettier に従う
- コンポーネント: `PascalCase.tsx`、ユーティリティ: `camelCase.ts`

---

## チケット管理

- **チケット = PR単位**、**サブチケット = コミット単位**
- ステータスはファイルの配置ディレクトリで管理: `backlog/` → `in-progress/` → `done/`
- ファイル命名: `TICKET-{連番3桁}_{概要}.md`（例: `TICKET-001_問題作成API実装.md`）
- テンプレート: `tickets/TICKET_TEMPLATE.md` を参照

**Claudeの作業フロー:**
1. 着手前にチケットを `backlog/` で確認、なければ起票する
2. 着手時に `in-progress/` へ移動し、ブランチ名を記入する
3. サブチケット消化のたびにチェックボックスをONにする
4. PR作成時にPR番号・リンクを記入する
5. マージ後に `done/` へ移動し、完了日を記入する

---

## TODO / 未決定事項

- [ ] `make swagger`（Swagger UI 起動）の実装
- [ ] 未使用の GCS クライアント（`backend/internal/infrastructure/storage/`）の扱い（削除 or 添付ファイル用途で活用）

※ 過去の未決定事項は解決済み: 永続化=Firestore・ページネーション=オフセット方式・状態管理=React Context・Go 1.25・リージョン=asia-northeast1（各セクションに反映済み）

---

## Claudeへの作業指示

1. **言語**: コメント・ドキュメントは日本語を基本とする
2. **技術選定**: 上記スタックを前提とし、勝手に別技術を導入しない
3. **アーキテクチャ遵守**: クリーンアーキテクチャの層間依存ルールを厳守する。依存の逆転が発生するコードは書かない
4. **API First**: 新しいAPIを追加・変更する場合は、必ず `api/openapi.yaml` を先に更新してから実装する
5. **チケット駆動**: タスクに着手する前に必ずチケットを確認・起票し、`tickets/` のライフサイクルに従って管理する
6. **変更提案**: 設計変更を伴う実装は、コードを書く前に方針を提示し確認を取る
7. **テスト**: 新機能追加時は必ずテストコードを合わせて作成する。特にユースケース層はユニットテストを必須とする
8. **セキュリティ**: 入力バリデーションやXSS対策を怠らない
