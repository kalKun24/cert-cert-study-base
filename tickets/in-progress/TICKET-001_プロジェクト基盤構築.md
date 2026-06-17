# TICKET-001 プロジェクト基盤構築

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-001 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-17 |
| 完了日 | - |
| ブランチ名 | `feature/project-scaffold` |
| PR番号 | - |
| PRリンク | （PR作成後に記入） |

---

## 概要

Go（バックエンド）・React（フロントエンド）のプロジェクト骨格を構築し、ローカル開発環境・CI/CDパイプラインを整備する。このチケットが完了することで、以降のチケットが独立して実装に着手できる状態になる。

---

## 背景・目的

現状リポジトリにはCLAUDE.md・チケットテンプレートしか存在しない。まず全機能の土台となるディレクトリ構成・Dockerize・Makefile・GitHub Actions を整備し、開発フローを確立する必要がある。

---

## 受け入れ条件

- [ ] `backend/` に Go モジュール（go.mod）が初期化されており、`make up` でサーバーが起動する
- [ ] `frontend/` に React + TypeScript プロジェクトが初期化されており、`make up` で開発サーバーが起動する
- [ ] `docker-compose.yml` でバックエンド・フロントエンドをまとめて起動できる
- [ ] `Makefile` に `up` / `down` / `test` / `lint` / `swagger` / `build` コマンドが定義されている
- [ ] `api/openapi.yaml` に OpenAPI 3.0 のスケルトン定義（info・servers・paths の空定義）が存在する
- [ ] GitHub Actions ワークフロー（テスト・lint）が `main` マージ時に実行される
- [ ] `make lint`（golangci-lint）・`make test` がエラーなく完了する

---

## サブチケット（コミット単位）

- [x] `chore(backend): Goモジュールの初期化とディレクトリ構成を作成`
- [x] `chore(frontend): React + TypeScriptプロジェクトを初期化`
- [x] `chore(infra): docker-compose.ymlとDockerfileを作成`
- [x] `chore(make): Makefileにローカル開発コマンドを定義`
- [x] `docs(api): openapi.yamlのスケルトンを作成`
- [x] `chore(ci): GitHub ActionsのCI/CDワークフローを追加`

---

## 関連情報

- 関連チケット: TICKET-002〜009（本チケット完了が着手前提）
- 参考: CLAUDE.md「ローカル開発環境」「CI/CD」セクション
- 備考: GCSエミュレータ方式（fake-gcs-server or ローカルファイルフォールバック）はTODOのため、本チケットではGCSクライアントのDIインターフェース定義のみ行う
