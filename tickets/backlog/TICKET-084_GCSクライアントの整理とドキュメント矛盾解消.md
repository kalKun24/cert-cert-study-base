# TICKET-084 GCSクライアントの整理とドキュメント矛盾解消

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-084 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-07-17 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | - |
| PR番号 | - |
| PRリンク | （PR作成後に記入） |

---

## 概要

どこからも参照されていない GCS クライアント実装（`backend/internal/infrastructure/storage/`）の扱いを決定（削除 or 添付ファイル機能用に温存）し、GCS の扱いについて矛盾している各ドキュメント（CLAUDE.md / README.md / .env.example）の記述を実態に合わせて整合させる。

---

## 背景・目的

Firestore 移行（TICKET-062）後、GCS は実際には使われていないが、コードとドキュメントに以下の不整合が残っている。

- `backend/internal/infrastructure/storage/` には `gcs_client.go`（Read/Write/Delete/Exists/List のフル実装）と `client_factory.go`（エミュレータ対応のクライアント生成）が存在するが、**バックエンドのどこからも参照されていない dead code** になっている
- CLAUDE.md は「GCS クライアントはインターフェース定義のみで現状未使用」と記載しており、フル実装が存在する実態と不一致。`storage.go` の docstring（「具体的な実装は本チケットのスコープ外」）も陳腐化している
- README.md は GCS を技術スタックの正式構成要素として記載し、`GCS_BUCKET` を「必須」環境変数としている（56, 108, 184, 213, 358 行付近）
- `.env.example` は `GCS_BUCKET`（必須扱い）と `GCS_EMULATOR_HOST`（fake-gcs-server 前提）を記載しているが、`docker-compose.yml` に fake-gcs-server サービスは存在しない

ドキュメント間で「未使用」と「必須・稼働中」が混在しており、新規参画者・エージェント双方の誤解の元になるため解消する。

---

## 受け入れ条件

- [ ] GCS クライアントの扱い（削除 or 添付ファイル機能用に温存）が決定され、チケットに記録されている
- [ ] 決定に沿ってコードが整理されている（削除の場合: `storage/` パッケージの除去。温存の場合: `storage.go` の陳腐化した docstring の更新）
- [ ] CLAUDE.md・README.md・`.env.example` の GCS 関連記述が実態と一致している（削除の場合: GCS 記述の除去、温存の場合: 「将来の添付ファイル用途・現状未配線」と明記）
- [ ] `make test` が正常終了する

---

## サブチケット（コミット単位）

- [ ] `refactor(storage): 未使用のGCSクライアントを整理`（削除 or docstring 更新）
- [ ] `docs(readme): GCS関連の記述を実態に合わせて整合`

---

## 関連情報

- 関連チケット: TICKET-062（Firestore移行。GCS が未使用になった起点）
- 参考: `backend/internal/infrastructure/storage/`（storage.go / gcs_client.go / client_factory.go）
- 備考: 温存する場合、docker-compose への GCS エミュレータ（fake-gcs-server）追加は添付ファイル機能の実装チケットで対応する（本チケットのスコープ外）
