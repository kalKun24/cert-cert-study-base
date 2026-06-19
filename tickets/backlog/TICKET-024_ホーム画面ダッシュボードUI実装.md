# TICKET-024 ホーム画面ダッシュボードUI実装

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-024 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-024` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

現在ほぼ空の状態（見出しとキャッチコピーのみ）のホーム画面（HomePage）を、ウェルカムメッセージとクイックアクションカードを含むダッシュボード的な画面に改善する。

---

## 背景・目的

ログイン後に最初に表示されるホーム画面が「ホーム」という見出しとタグラインのみで、ユーザーが次に何をすればよいかの動線がまったくない。問題一覧・問題作成・チーム一覧への導線をクイックアクションカードとして提供し、アプリの価値を初期画面から伝える。

---

## 受け入れ条件

- [ ] ウェルカムメッセージとして `{display_name} さん、こんにちは` を表示する（AuthContext からユーザー情報を取得）
- [ ] クイックアクションカードを 3 つ表示する
  - 「問題一覧を見る」→ `/questions`
  - 「問題を作成する」→ `/questions/new`
  - 「チームを見る」→ `/teams`
- [ ] カードはデスクトップで 3 カラムグリッド、モバイルで 1 カラム縦積みになる
- [ ] 各カードにタイトル・説明文が含まれる
- [ ] カードはホバー時に視覚的なフィードバック（shadow 強調・translateY）がある
- [ ] カードは `<Link>` または `<a>` でラップされており、キーボードで操作可能
- [ ] 画面全体のスタイルが TICKET-022 のデザイントークンに準拠している

---

## サブチケット（コミット単位）

- [ ] `feat(home): ウェルカムメッセージとクイックアクションカードを実装`
- [ ] `feat(home): クイックアクションカードのスタイルとレスポンシブ対応を追加`

---

## 関連情報

### コンポーネント設計

```tsx
// HomePage.tsx 改修

// QuickActionCard (ローカルコンポーネント or 新規ファイル)
interface QuickActionCardProps {
  title: string;
  description: string;
  to: string;
}

function QuickActionCard({ title, description, to }: QuickActionCardProps) {
  return (
    <Link to={to} className="quick-action-card">
      <h2 className="quick-action-title">{title}</h2>
      <p className="quick-action-description">{description}</p>
    </Link>
  );
}
```

### レイアウト仕様

```css
.home-page { max-width: 900px; }

.home-welcome {
  margin-bottom: var(--space-8);
}
.home-welcome-title {
  font-size: var(--font-size-3xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  margin: 0 0 var(--space-2);
}
.home-tagline {
  font-size: var(--font-size-base);
  color: var(--color-text-secondary);
  margin: 0;
}

.quick-actions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-4);
}

.quick-action-card {
  display: block;
  padding: var(--space-6);
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  text-decoration: none;
  color: inherit;
  transition:
    box-shadow var(--transition-fast),
    border-color var(--transition-fast),
    transform var(--transition-fast);
}
.quick-action-card:hover {
  border-color: var(--color-primary-300);
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

/* モバイル */
@media (max-width: 767px) {
  .quick-actions-grid { grid-template-columns: 1fr; }
}
```

### i18n キー（ja.json に追加）

```json
{
  "home": {
    "welcome": "{name}さん、こんにちは",
    "action.questions.title": "問題一覧を見る",
    "action.questions.description": "過去の問題を検索・フィルタリングして学習する",
    "action.create.title": "問題を作成する",
    "action.create.description": "新しい問題・解答・解説を作成して共有する",
    "action.teams.title": "チームを見る",
    "action.teams.description": "勉強会チームのメンバーを確認・管理する"
  }
}
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 備考: 将来的に「最近の問題」「統計サマリー」などをカード追加できる拡張性を持たせること
