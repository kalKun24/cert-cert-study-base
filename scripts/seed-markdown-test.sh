#!/usr/bin/env bash
# Markdown プレビューテスト用データ投入スクリプト
# 使い方: bash scripts/seed-markdown-test.sh
# 前提: make up でアプリが起動済みであること（seed.sh 実行後でなくても動作）

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

post() {
  local token="$1" path="$2" data="$3"
  if [ -n "$token" ]; then
    curl -sf -X POST "$BASE_URL$path" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $token" \
      -d "$data"
  else
    curl -sf -X POST "$BASE_URL$path" \
      -H "Content-Type: application/json" \
      -d "$data"
  fi
}

echo "=========================================="
echo "  Markdown プレビューテストデータ投入"
echo "=========================================="

# --- 1. ログイン ---
echo ""
echo "【1】ログイン..."
ADMIN_TOKEN=$(post "" "/api/v1/auth/login" \
  "$(jq -n '{username:"admin",password:"Admin1234!"}')" \
  | jq -r '.data.token')
echo "    OK"

# --- 2. チームIDを取得（最初のチームを使用） ---
echo ""
echo "【2】チーム取得..."
TEAM_ID=$(curl -sf "$BASE_URL/api/v1/teams" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  | jq -r '.data[0].id // empty')

if [ -z "$TEAM_ID" ]; then
  echo "    チームが存在しないため新規作成..."
  TEAM_ID=$(post "$ADMIN_TOKEN" "/api/v1/teams" \
    "$(jq -n '{name:"Markdownテストチーム",description:"プレビューテスト用"}')" \
    | jq -r '.data.id')
fi
echo "    チームID: $TEAM_ID"

# --- 3. Markdown全要素テスト問題を作成 ---
echo ""
echo "【3】Markdownテスト問題を投入..."

BODY='# 見出し1（H1）— Markdown全要素テスト

## 見出し2（H2）— 強調・インライン

### 見出し3（H3）

#### 見出し4（H4）

##### 見出し5（H5）

###### 見出し6（H6）

---

## 段落と改行

これは1行目です。
これはEnterで改行した2行目です（remark-breaksが有効なら表示される）。

空行を入れると新しい段落になります。

---

## 強調・装飾

**太字テキスト** / *斜体テキスト* / ***太字＋斜体*** / ~~取り消し線~~

インラインコード: `const x = 42;`

---

## リスト

### 順序なしリスト

- 項目 A
- 項目 B
  - ネスト B-1
  - ネスト B-2
    - さらにネスト
- 項目 C

### 順序ありリスト

1. 手順 1
2. 手順 2
   1. サブ手順 2-1
   2. サブ手順 2-2
3. 手順 3

### チェックリスト

- [x] 完了タスク
- [x] これも完了
- [ ] 未完了タスク
- [ ] まだやっていない

---

## リンク

[Google](https://www.google.com)

[タイトル付き](https://www.google.com "Googleのホーム")

---

## 引用

> これは引用ブロックです。
> 複数行の引用も書けます。
>
> > ネストした引用です。

---

## コードブロック

```python
def hello(name: str) -> str:
    return f"Hello, {name}!"

print(hello("World"))
```

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
```

```bash
for i in {1..5}; do
  echo "ループ $i 回目"
done
```

---

## テーブル

| 左寄せ | 中央揃え | 右寄せ |
| :--- | :---: | ---: |
| セルA | セルB | セルC |
| 100 | 200 | 300 |

---

## 水平線

上のコンテンツ

---

下のコンテンツ

---

## エスケープ

\*エスケープしたアスタリスク\* / \`エスケープしたバッククォート\`

---

## 数式（KaTeX）

### インライン数式

シャノンのエントロピー: $H = -\sum_{i} p_i \log_2 p_i$

AES の鍵長と安全性: $2^{128}$ 通りの鍵空間

RSA の暗号化: $c = m^e \mod n$

### ブロック数式（中央揃え）

ベイズの定理:

$$
P(A \mid B) = \frac{P(B \mid A)\, P(A)}{P(B)}
$$

ALE（年間期待損失額）の計算式:

$$
\text{ALE} = \text{SLE} \times \text{ARO} = (AV \times EF) \times ARO
$$

オイラーの等式（番外）:

$$
e^{i\pi} + 1 = 0
$$'

ANSWER='## 解答

このテスト問題に解答はありません。

プレビューで以下が正しく表示されることを確認してください。

- [ ] H1〜H6 の見出しスタイル（H2 に下線が入るか）
- [ ] Enter 1回の改行がプレビューに反映されるか
- [ ] **太字** / *斜体* / ~~取り消し線~~ が正しく表示されるか
- [ ] 順序なし・順序あり・チェックリストのリスト
- [ ] コードブロックのシンタックスハイライト
- [ ] テーブルの整列
- [ ] 引用ブロックのスタイル
- [ ] 水平線の表示'

EXPLANATION='## 解説

このデータは Markdown プレビューの動作確認専用です。

```
remark-breaks プラグインが有効であれば、
段落中の単一改行がプレビューに <br> として反映されます。
```'

MEMO='## 確認ポイント

- H2 に `border-bottom` が表示されているか（TICKET-078 の確認）
- Enter 1回改行が機能しているか（TICKET-077 の確認）
- シンタックスハイライトが機能しているか'

post "$ADMIN_TOKEN" "/api/v1/teams/$TEAM_ID/questions" \
  "$(jq -n \
    --arg title "【Markdownプレビューテスト】全要素確認用" \
    --arg body "$BODY" \
    --arg answer "$ANSWER" \
    --arg explanation "$EXPLANATION" \
    --arg memo "$MEMO" \
    '{title:$title,body:$body,answer:$answer,explanation:$explanation,memo:$memo,tags:[],status:"published"}')" \
  | jq -r '"    問題ID: " + .data.id'

echo ""
echo "=========================================="
echo "  投入完了！"
echo "=========================================="
echo ""
echo "  http://localhost:3000 にアクセスして"
echo "  投入した問題のプレビューを確認してください。"
echo ""
