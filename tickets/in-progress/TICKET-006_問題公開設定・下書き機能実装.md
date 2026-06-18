# TICKET-006 問題公開設定・下書き機能実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-006 |
| ステータス | 🟡 作業中 |
| 作成日 | 2026-06-17 |
| 着手日 | 2026-06-18 |
| 完了日 | - |
| ブランチ名 | `feature/question-visibility` |
| PR番号 | - |
| PRリンク | （PR作成後に記入） |

---

## 概要

問題の作成者が「公開」「チーム公開」「非公開」「下書き」を選択できる公開設定機能を実装する。Webアプリ内のログイン済みユーザ間での可視性を制御する。外部への共有URLは提供しない。

---

## 背景・目的

CLAUDE.mdに「Markdown形式でのテキストの共有」が主な機能として掲げられているが、共有はWebアプリ内のログイン済みユーザ間に限定する。作成途中の問題を下書きとして保存し、完成後に全体またはチーム単位で公開できるワークフローを提供する。

---

## ステータス・公開範囲の定義

### status フィールド

| 値 | 表示名 | 説明 |
|---|---|---|
| `draft` | 下書き | 作成途中。作成者のみ閲覧可能 |
| `private` | 非公開 | 完成済みだが非公開。作成者のみ閲覧可能 |
| `published` | 公開 | `visibility_scope` に従って閲覧範囲を制御 |

### visibility_scope フィールド（status が `published` のときのみ有効）

| 値 | 説明 |
|---|---|
| `all` | 全ログインユーザが閲覧可能 |
| `team` | `published_team_ids` に指定したチームのメンバーのみ閲覧可能（複数チーム指定可） |

---

## 受け入れ条件

- [ ] Questionエンティティに `status` / `visibility_scope` / `published_team_ids` フィールドが存在する
- [ ] 作成時のデフォルトは `status: draft`
- [ ] `PATCH /api/v1/questions/{id}/visibility` でステータス・公開範囲を変更できる
- [ ] `GET /api/v1/questions` 一覧取得ルール:
  - `status: published` かつ `visibility_scope: all` → 全ログインユーザに返す
  - `status: published` かつ `visibility_scope: team` → リクエストユーザが `published_team_ids` のいずれかに所属する場合のみ返す
  - `status: draft` / `private` → 作成者本人のみ返す（`admin` は全件取得可）
- [ ] `GET /api/v1/questions/{id}` でも上記と同じ可視性ルールを適用し、閲覧不可の場合は404を返す
- [ ] `published_team_ids` には複数のチームIDを指定できる
- [ ] ステータス・公開範囲の変更は作成者本人のみ可能（`admin` は全件変更可）
- [ ] `openapi.yaml` に公開設定変更エンドポイントおよび各フィールドのSwagger定義が存在する
- [ ] ユースケース層のユニットテストが作成されている

---

## サブチケット（コミット単位）

- [ ] `docs(api): 公開設定フィールドと変更エンドポイントをopenapi.yamlに追加`
- [ ] `feat(domain): Questionエンティティにvisibility_scope・published_team_idsを追加`
- [ ] `feat(usecase): 公開設定変更ユースケースと可視性フィルタロジックを実装`
- [ ] `feat(interface): 公開設定変更ハンドラとDTOを実装`
- [ ] `feat(infrastructure): GCSリポジトリの可視性フィルタ対応`
- [ ] `test(usecase): 公開設定ユースケースのユニットテストを作成`

---

## 関連情報

- 関連チケット: TICKET-004（Questionエンティティの拡張）、TICKET-003（チーム管理・メンバーシップ判定の前提）、TICKET-009（フロントエンドに公開設定UIを追加）、TICKET-007（コメント投稿可否の判定に閲覧権限ロジックを再利用）
- 参考: CLAUDE.md「主な機能」（Markdownテキストの共有）
- 備考: 可視性チェックはユースケース層で行う。`published_team_ids` に含まれるチームが削除された場合の扱い（自動除外 or エラー）はTICKET-003と設計を合わせること
