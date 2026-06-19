# TICKET-029 チーム一覧画面UI改善

## 基本情報

| 項目 | 内容 |
|---|---|
| チケットID | TICKET-029 |
| ステータス | 🔴 未着手 |
| 作成日 | 2026-06-19 |
| 着手日 | - |
| 完了日 | - |
| ブランチ名 | `feature/TICKET-029` |
| PR番号 | - |
| PRリンク | - |

---

## 概要

チーム一覧画面（TeamListPage）にスタイルを適用し、チームのカードデザイン・説明文の折り返し処理・ローディング/エラー/空状態のスタイルを改善する。

---

## 背景・目的

チーム一覧画面は現状スタイルがなく、チームリストが単純なリンクのリストとして表示されている。チーム名・説明文をカードデザインで表示し、説明文が長い場合の行数クランプ、hover エフェクトを追加して問題一覧と統一感のある UI にする。

---

## 受け入れ条件

- [ ] チーム一覧の各アイテムがカードデザイン（border・shadow・border-radius）で表示される
  - チーム名（セミボールド・大きめ）
  - 説明文（通常テキスト・最大2行でクランプ）
  - hover 時に border-color が primary 色に変化し shadow が強調される
- [ ] ページヘッダー（タイトル + 「チームを作成」ボタン）が横並びで表示される
  - 「チームを作成」ボタンは admin/teamowner にのみ表示（現状実装済み）
- [ ] チームが 0 件のときの空状態メッセージがスタイルされている
- [ ] ローディング状態・エラー状態がスタイルされている
- [ ] デスクトップ・モバイル両方で正しくレイアウトされる（max-width: 800px）

---

## サブチケット（コミット単位）

- [ ] `feat(team-list): チームカードリストのスタイルを適用`
- [ ] `feat(team-list): ページヘッダーと空状態・ローディング・エラー状態のスタイルを適用`

---

## 関連情報

### スタイル仕様

```css
.team-list-page { max-width: 800px; }

.team-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-6);
}

.team-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

/* チームカード（<li> に直接スタイル） */
.team-list-item {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-xs);
  transition:
    box-shadow var(--transition-fast),
    border-color var(--transition-fast),
    transform var(--transition-fast);
}
.team-list-item:hover {
  border-color: var(--color-primary-300);
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

/* カードリンク（<li> > <Link>） */
.team-list-link {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-4) var(--space-5);
  text-decoration: none;
  color: inherit;
}

.team-list-name {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.team-list-description {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  /* 2行でクランプ */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.team-list-empty {
  padding: var(--space-16) 0;
  text-align: center;
  color: var(--color-text-secondary);
}
```

### 関連チケット

- 関連チケット: TICKET-022（デザインシステム構築）が完了していること
- 関連チケット: TICKET-030（チーム詳細）と合わせてチーム関連 UI を統一する
