# webapp-dev-with-claude-template

Claude Code を使った Web アプリケーション開発をすぐに始められるテンプレートリポジトリです。

毎回ゼロから設定ファイルやエージェント定義を書く手間を省き、チーム開発・個人開発どちらでもそのまま流用できます。

---

## 含まれるもの

| ファイル / ディレクトリ | 役割 |
|---|---|
| `CLAUDE.md` | Claude への作業指示とプロジェクト設定のテンプレート |
| `.claude/agents/` | 役割別サブエージェント定義（4種） |
| `tickets/` | チケット駆動開発用のディレクトリ構成とチケットテンプレート |

### サブエージェント一覧

| ファイル | 役割 | 呼び出しタイミング |
|---|---|---|
| `project-setup.md` | clone 後の初期設定（CLAUDE.md のプレースホルダを埋める） | セットアップ時 |
| `engineering-backend-architect.md` | Go バックエンド設計・データベース・クラウドインフラ | バックエンド実装時 |
| `engineering-frontend-developer.md` | React / TypeScript フロントエンド実装 | フロントエンド実装時 |
| `engineering-code-reviewer.md` | コードレビュー（正確性・セキュリティ・保守性・パフォーマンス） | PR レビュー時 |
| `engineering-security-engineer.md` | 脅威モデリング・脆弱性診断・セキュアコードレビュー | セキュリティ確認時 |

---

## Getting Started

### 1. リポジトリを clone する

```bash
git clone https://github.com/<your-org>/webapp-dev-with-claude-template.git <your-project-name>
cd <your-project-name>
```

リモートを自分のリポジトリに向け直します。

```bash
git remote set-url origin https://github.com/<your-org>/<your-project-name>.git
```

### 2. Claude Code を起動してセットアップを依頼する

```bash
claude
```

起動後、以下のように依頼すると **Project Setup エージェント**が対話形式でプロジェクト名・技術スタックなどをヒアリングし、`CLAUDE.md` のプレースホルダを自動で埋めます。

```
プロジェクトのセットアップをして
```

セットアップ完了後は、Claude が `CLAUDE.md` を読み込みプロジェクトのルール・スタックに従って作業を進めます。

> `CLAUDE.md` を手動で編集したい場合は `{...}` のプレースホルダを直接書き換えてください。

---

## チケット管理フロー

チケット = PR 単位、サブチケット = コミット単位で管理します。

```
作成       着手          完了
backlog/ → in-progress/ → done/
```

**ファイル命名規則:** `TICKET-{連番3桁}_{概要}.md`（例: `TICKET-001_ログイン機能実装.md`）

テンプレートは `tickets/TICKET_TEMPLATE.md` を参照してください。

**Claude の作業フロー:**

1. 着手前に `backlog/` のチケットを確認（なければ起票）
2. 着手時に `in-progress/` へ移動し、ブランチ名を記入
3. サブチケット消化のたびにチェックボックスをオン
4. PR 作成時に PR 番号・リンクを記入
5. マージ後に `done/` へ移動し、完了日を記入

---

## チケットファイルを手動編集する場合の注意

Dev Team エージェントが起動する**前**であれば、チケットファイルを直接編集しても問題ありません。  
ただし、以下の点に注意してください。

| 操作 | 注意点 |
|---|---|
| ファイルを手動で `in-progress/` や `done/` に移動する | Dev Team は `backlog/` 固定でチケットを探すため、**見つけられなくなります**。移動はエージェントに任せてください。 |
| エージェント実行中に編集する | Phase 4 でチェックボックス更新の上書きが発生し、編集内容が失われる場合があります。 |
| Markdown のテーブル・セクション構造を変える | 必須セクション（概要・受け入れ条件・サブチケット）が欠けると、エージェントが内容を正しく解釈できなくなります。 |
| サブチケットを着手後に追加・削除する | Phase 4 のチェックボックス更新は実装済み項目を基準にするため、後から追加した項目は反映されません。 |

**推奨タイミング**: 内容の変更は **Dev Team に着手を依頼する前** に行ってください。

---

## カスタマイズのヒント

### 技術スタックを変える場合

`CLAUDE.md` の技術スタック表とディレクトリ構成・開発ガイドラインを書き換えます。  
Claude はこれを読んで技術選定の判断をするため、スタックと記述を一致させてください。

### エージェント定義を追加・変更する場合

`.claude/agents/` にマークダウンファイルを追加します。  
ファイルの先頭に以下のフロントマターを記述することで、エージェントの名前・説明・使用ツールを定義できます。

```yaml
---
name: agent-name
description: このエージェントを使う場面の説明
tools:
  - Read
  - Edit
  - Bash
---
```

既存の 4 つのエージェント定義を参考にしてください。

---

## ディレクトリ構成

```
.
├── .claude/
│   └── agents/                  # サブエージェント定義
│       ├── project-setup.md             # 初期セットアップ
│       ├── engineering-backend-architect.md
│       ├── engineering-code-reviewer.md
│       ├── engineering-frontend-developer.md
│       └── engineering-security-engineer.md
├── tickets/
│   ├── TICKET_TEMPLATE.md       # チケットテンプレート
│   ├── backlog/                 # 未着手チケット
│   ├── in-progress/             # 作業中チケット
│   └── done/                   # 完了チケット
└── CLAUDE.md                   # Claude への作業指示
```
