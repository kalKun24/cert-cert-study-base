# TICKET-039 チームREADME機能

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-039 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-20 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/team-readme` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

チームエンティティに Markdown 形式の README フィールドを追加し、
チームオーナーがチーム作成・編集画面で自由に記載できるようにする。

---

## 背景・目的

チームの目的・ルールを共有できる場所として README が必要。
ホーム画面ダッシュボード（TICKET-040）でも表示する。

---

## 受け入れ条件

- [ ] チームエンティティに `readme: string`（Markdown）フィールドが追加されている
- [ ] `PUT /api/v1/teams/{id}` が `readme` フィールドの更新に対応している
- [ ] チーム作成・編集画面に Markdown エディタ（`@uiw/react-md-editor`）で README を入力できる（チームオーナーのみ）
- [ ] `api/openapi.yaml` が更新されている

---

## サブチケット（コミット単位）

- [ ] `feat(domain): チームエンティティに readme フィールドを追加`
- [ ] `feat(usecase): チーム更新ユースケースで readme を扱えるようにする`
- [ ] `docs(api): openapi.yaml を更新`
- [ ] `feat(page): チーム作成・編集画面に Markdown README 入力欄を追加`

---

## 関連情報

- 関連チケット: TICKET-040（ダッシュボード表示）
- 備考:
  - README の表示は `react-markdown` + `rehype-sanitize` で XSS 対策を施す
  - エディタ・プレビュー UI の配色は `global.css` の CSS カスタムプロパティと既存の Markdown エディタスタイルに従う
