# TICKET-062 Firestore へのデータストア移行

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-062 |
| ステータス | 🟡 進行中 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | - |
| ブランチ名 | feature/firestore-migration |
| PR番号 | - |
| PRリンク | - |

---

## 概要

GCS の JSON ファイルをデータベース代わりに使う現状の設計から、Cloud Firestore へ移行する。
マルチインスタンス競合・バケット誤設定時の全漏洩・アトミック性欠如という GCS データストア固有のリスクを解消する。

クリーンアーキテクチャの恩恵を活かし、**ドメイン層・ユースケース層・インターフェース層は変更しない**。
`infrastructure/repository/` 層の実装を差し替えることで移行を完結させる。

---

## 背景・目的

- GCS の `sync.RWMutex` はプロセス内排他制御のみで、Cloud Run 複数インスタンス時に競合が発生する
- `teams/{id}/questions.json` のように全件を1ファイルに保持する設計は I/O コストと競合窓口を拡大させる
- Firestore はドキュメント単位のアトミック更新・トランザクションをサポートし、Cloud Run との相性が良い
- Firestore Emulator で `FIRESTORE_EMULATOR_HOST` を設定するだけでローカル・CI 問わず接続先を切り替えられる（現在の `GCS_EMULATOR_HOST` と同じパターン）

---

## Firestore コレクション設計

```
users/{userID}
teams/{teamID}
teams/{teamID}/members/{userID}
teams/{teamID}/tags/{tagID}
teams/{teamID}/questions/{questionID}
teams/{teamID}/questions/{questionID}/comments/{commentID}
teams/{teamID}/notes/{noteID}
teams/{teamID}/notes/{noteID}/comments/{commentID}
invitations/{invitationID}
```

---

## 受け入れ条件

### リポジトリ実装
- [ ] `backend/internal/infrastructure/firestore/` パッケージに以下を新規作成
  - `user_repository.go`
  - `team_repository.go`
  - `question_repository.go`
  - `comment_repository.go`
  - `note_repository.go`
  - `note_comment_repository.go`
  - `tag_repository.go`
  - `invitation_repository.go`
- [ ] 各実装が既存のドメインリポジトリインターフェースを満たすことをコンパイル時チェック（`var _ domain.XxxRepository = (*FirestoreXxxRepository)(nil)`）
- [ ] 旧 GCS リポジトリ（`infrastructure/repository/*.go`）は削除する
- [ ] `infrastructure/storage/` パッケージは将来のファイルストレージ用途として残す（main.go の DI からは外す）

### 依存・環境
- [ ] `backend/go.mod` に `cloud.google.com/go/firestore` を追加
- [ ] `backend/go.mod` から `github.com/fsouza/fake-gcs-server` を削除（統合テストも Firestore Emulator に移行するため）
- [ ] `backend/cmd/server/main.go` の DI を Firestore リポジトリに差し替え
- [ ] 必要な環境変数: `GCP_PROJECT_ID`（Firestore プロジェクト ID）、`FIRESTORE_EMULATOR_HOST`（エミュレータ使用時）
- [ ] GCS 関連環境変数（`GCS_BUCKET`, `GCS_EMULATOR_HOST`）を main.go・.env から除去

### ローカル開発環境
- [ ] `docker-compose.yml` の `gcs-emulator` サービスを Firestore Emulator に差し替え
  - イメージ: `gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine`
  - 起動コマンド: `gcloud emulators firestore start --host-port=0.0.0.0:8080 --project=local-project`
  - backend サービスへ `FIRESTORE_EMULATOR_HOST=firestore-emulator:8080` と `GCP_PROJECT_ID=local-project` を渡す
- [ ] `Makefile` を更新（コメント・ヘルプを Firestore Emulator に合わせる）

### CI
- [ ] `.github/workflows/ci.yml` のバックエンドテストジョブに Firestore Emulator の起動ステップを追加
  - `gcloud components install cloud-firestore-emulator` → `gcloud emulators firestore start` をバックグラウンド起動
  - `FIRESTORE_EMULATOR_HOST: localhost:8080` と `GCP_PROJECT_ID: test-project` を環境変数に追加

### 統合テスト
- [ ] `backend/internal/infrastructure/repository/user_repository_integration_test.go` を `firestore/` パッケージ以下に移行
  - fake-gcs-server の代わりに `FIRESTORE_EMULATOR_HOST` を使用（Go テスト内から `os.Setenv` でポートを指定）
  - Firestore クライアントをテスト内で初期化してテスト用コレクションに書き込み・読み取りを検証

### ビルド・品質
- [ ] `go build ./...` が通ること
- [ ] `golangci-lint run` が通ること
- [ ] 既存のユースケースユニットテスト（`usecase/*_test.go`）が変更なしにパスすること（リポジトリモックはそのまま）

---

## サブチケット（コミット単位）

- [x] `chore(deps): cloud.google.com/go/firestoreを追加・fake-gcs-serverを削除`
- [x] `feat(infra): Firestoreリポジトリ実装を追加（user・team・invitation）`
- [x] `feat(infra): Firestoreリポジトリ実装を追加（question・comment・tag）`
- [x] `feat(infra): Firestoreリポジトリ実装を追加（note・note_comment）`
- [x] `refactor(main): DIをFirestoreリポジトリに差し替え・GCS依存を除去`
- [x] `chore(infra): docker-composeをFirestore Emulatorに更新`
- [x] `ci: GitHub ActionsにFirestore Emulatorの起動ステップを追加`
- [x] `test(infra): ユーザーリポジトリ統合テストをFirestoreに移行`

---

## 関連情報

- 参考実装（移行元）: `backend/internal/infrastructure/repository/`
- 参考ドキュメント: `cloud.google.com/go/firestore` GoDoc
- 既存ドメインインターフェース: `backend/internal/domain/*.go`（変更なし）
- 備考:
  - Firestore の Go クライアントは `FIRESTORE_EMULATOR_HOST` 環境変数を自動で参照する。コードの分岐不要
  - エミュレータ使用時もプロジェクト ID が必要（値は任意）
  - `sync.RWMutex` は Firestore リポジトリでは不要（Firestore がアトミック性を保証するため）
  - TagRepository は QuestionRepository への依存を持つ（タグ削除時の使用中チェック）。この依存は Firestore 版でも同様に保持する
