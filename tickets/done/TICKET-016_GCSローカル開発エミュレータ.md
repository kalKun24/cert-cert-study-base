# TICKET-016 GCS ローカル開発エミュレータ

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-016 |
| ステータス | 🟢 完了 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-18 |
| 完了日 | 2026-06-18 |
| ブランチ名 | `feature/gcs-local-emulator` |
| PR番号 | #4 |
| PRリンク | https://github.com/kalKun24/cert-cert-study-base/pull/4 |

---

## 概要

`fake-gcs-server` を docker-compose に組み込み、ローカル開発・CI テストで実 GCS を使わずに開発・テストできる環境を整備する。CLAUDE.md の未決定事項「GCS ローカル開発時のエミュレータ方式」を解決する。

---

## 背景・目的

現状、ローカルで動かすには実際の GCS バケットと GCP 認証情報が必要なため、開発者全員が GCP アカウントを持たないと開発できない。また CI でも実 GCS を叩くのはコスト・速度・副作用の観点でリスクがある。

---

## 方針

`fake-gcs-server`（Docker イメージ: `fsouza/fake-gcs-server`）を採用する。

- `docker-compose.yml` に `gcs-emulator` サービスを追加
- バックエンドの GCS クライアント初期化時に、環境変数 `GCS_EMULATOR_HOST` が設定されている場合はエミュレータに向ける
- CI（GitHub Actions）でも同様に `GCS_EMULATOR_HOST` を設定してテストを実行

```
環境変数:
  GCS_EMULATOR_HOST=http://gcs-emulator:4443  # ローカル
  GCS_EMULATOR_HOST=http://localhost:4443       # CI
```

---

## 受け入れ条件

- [ ] `make up` で `fake-gcs-server` が起動し、バックエンドがエミュレータ経由で GCS 操作できる
- [ ] `GCS_EMULATOR_HOST` 未設定時は実 GCS クライアントとして動作する（本番互換）
- [ ] `make test` がエミュレータを使った状態で全テスト通過する
- [ ] GitHub Actions CI でも `fake-gcs-server` を使ったインテグレーションテストが実行される
- [ ] `.env.example` に `GCS_EMULATOR_HOST` の説明が追加されている

---

## サブチケット（コミット単位）

- [x] `chore(infra): docker-compose に fake-gcs-server サービスを追加`
- [x] `feat(infrastructure): GCS_EMULATOR_HOST によるエミュレータ切り替えを実装`
- [x] `test(infrastructure): GCS エミュレータを使ったリポジトリ統合テストを追加`
- [x] `chore(ci): GitHub Actions に fake-gcs-server を組み込む`
- [x] `docs(readme): ローカル開発環境セットアップ手順を更新`

---

## 関連情報

- 関連チケット: TICKET-001（docker-compose 基盤）、TICKET-010（本番 GCS バケット）
- 参考: [fake-gcs-server](https://github.com/fsouza/fake-gcs-server)
- 備考: CLAUDE.md TODO「GCS ローカル開発時のエミュレータ方式（`fake-gcs-server` or ローカルファイルフォールバック）」を本チケットで解決する
