# TICKET-069 CI/CD セキュリティスキャンとシークレット管理の整備

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-069 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-22 |
| 着手日 | 2026-06-22 |
| 完了日 | - |
| ブランチ名 | feature/TICKET-069 |
| PR番号 | #58 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/58 |

---

## 概要

CI パイプラインへのシークレット検出・SAST の追加、.env の JWT シークレット管理方針の明確化、`context.TODO()` 29箇所の適切なコンテキスト伝播への置き換えを行う。

---

## 背景・目的

- **.env の JWT シークレット共有化** (`.env`): `JWT_SECRET` と `SEED_ADMIN_PASSWORD` が `.env` に平文で存在する。`.gitignore` に登録されているため git 履歴への混入は現時点で防げているが、`git add .` による誤コミットを検出する仕組みがない。CI パイプラインにシークレット検出ステップを追加することでセーフティネットを構築する。
- **SAST 不在** (`.github/workflows/ci.yml`): CI にテスト・Lint はあるが、既知の CVE を持つ依存ライブラリや Go のセキュリティアンチパターンを検出するスキャンが存在しない。
- **context.TODO() 29箇所** (usecase 層全体): リクエストキャンセルシグナルが Firestore 操作に伝播しない。接続切断を繰り返すスロータック型 DoS に利用されるリスクがある。コード自体に `TODO: context.Context を追加` のコメントが記載されており、既知の技術的負債として認識されている。

---

## 受け入れ条件

- [x] `.github/workflows/ci.yml` に `gitleaks` または `truffleHog` によるシークレット検出ステップを追加する
- [x] CI パイプラインに `gosec` または Trivy による Go 依存関係スキャンステップを追加する
- [x] `.env.example` にダミー値を記載したテンプレートを整備し、README のセットアップ手順に `.env` の生成方法を追記する
- [x] `context.TODO()` を使用しているユースケース層のメソッドシグネチャに `ctx context.Context` を追加し、Firestore 操作に伝播させる

---

## サブチケット（コミット単位）

- [x] `ci: gitleaks によるシークレット検出ステップを CI に追加`
- [x] `ci: Trivy / gosec による依存関係スキャンステップを CI に追加`
- [x] `docs: .env.example を整備し README にセットアップ手順を追記`
- [x] `refactor(usecase): context.TODO() を ctx に置き換えてリクエストコンテキストを伝播`

---

## 関連情報

- 関連チケット: TICKET-065（セキュリティ基盤強化）と連動して対応すると効率的
- 参考: gitleaks https://github.com/gitleaks/gitleaks、gosec https://github.com/securego/gosec
- 備考: 将来対応。ただし .env の管理方針（gitleaks 追加部分）は CI 整備とセットで早期対応が望ましい。検出エージェント: Security Engineer（Critical #1、Medium #9、Low #13）。品質チェック日: 2026-06-22。
