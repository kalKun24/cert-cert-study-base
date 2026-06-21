#!/usr/bin/env bash
# テストデータ投入スクリプト
# 使い方: bash scripts/seed.sh
# 前提: make up でアプリが起動済みであること

set -euo pipefail

BASE_URL="http://localhost:8080"

# --- ヘルパー ---

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

patch() {
  local token="$1" path="$2" data="$3"
  curl -sf -X PATCH "$BASE_URL$path" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $token" \
    -d "$data"
}

login() {
  local username="$1" password="$2"
  post "" "/api/v1/auth/login" \
    "$(jq -n --arg u "$username" --arg p "$password" '{username:$u,password:$p}')" \
    | jq -r '.data.token'
}

create_user() {
  local token="$1" username="$2" display="$3" email="$4" pass="$5" role="${6:-user}"
  post "$token" "/api/v1/users" \
    "$(jq -n --arg u "$username" --arg d "$display" --arg e "$email" \
           --arg p "$pass" --arg r "$role" \
           '{username:$u,display_name:$d,email:$e,password:$p,role:$r}')" \
    | jq -r '.data.id'
}

create_team() {
  local token="$1" name="$2" desc="$3"
  post "$token" "/api/v1/teams" \
    "$(jq -n --arg n "$name" --arg d "$desc" '{name:$n,description:$d}')" \
    | jq -r '.data.id'
}

add_member() {
  local token="$1" team_id="$2" user_id="$3"
  post "$token" "/api/v1/teams/$team_id/members" \
    "$(jq -n --arg uid "$user_id" '{user_id:$uid}')" \
    | jq -r '.data.role'
}

create_tag() {
  local token="$1" team_id="$2" name="$3"
  post "$token" "/api/v1/teams/$team_id/tags" \
    "$(jq -n --arg n "$name" '{name:$n}')" \
    | jq -r '.data.name'
}

create_question() {
  local token="$1" team_id="$2"
  local title="$3" body="$4" answer="$5" explanation="$6" memo="$7"
  local tags_json="$8" status="${9:-published}"
  post "$token" "/api/v1/teams/$team_id/questions" \
    "$(jq -n \
        --arg t "$title" --arg b "$body" --arg a "$answer" \
        --arg e "$explanation" --arg m "$memo" --arg s "$status" \
        --argjson tg "$tags_json" \
        '{title:$t,body:$b,answer:$a,explanation:$e,memo:$m,tags:$tg,status:$s}')" \
    | jq -r '.data.id'
}

post_comment() {
  local token="$1" team_id="$2" question_id="$3" body="$4"
  post "$token" "/api/v1/teams/$team_id/questions/$question_id/comments" \
    "$(jq -n --arg b "$body" '{body:$b}')" | jq -r '.data.id'
}

echo "=========================================="
echo "  cert-study-base テストデータ投入"
echo "=========================================="

# ===========================================================
# 1. admin ログイン
# ===========================================================
echo ""
echo "【1】admin ログイン..."
ADMIN_TOKEN=$(login "admin" "Admin1234!")
echo "    OK"

# ===========================================================
# 2. ユーザー作成
# ===========================================================
echo ""
echo "【2】ユーザー作成..."

TANAKA_ID=$(create_user "$ADMIN_TOKEN" "tanaka"   "田中 優"   "tanaka@example.com"   "Password1!")
echo "    田中 優 (tanaka): $TANAKA_ID"

SATO_ID=$(create_user "$ADMIN_TOKEN" "sato"     "佐藤 健"   "sato@example.com"     "Password1!")
echo "    佐藤 健 (sato): $SATO_ID"

YAMADA_ID=$(create_user "$ADMIN_TOKEN" "yamada"   "山田 花子" "yamada@example.com"   "Password1!")
echo "    山田 花子 (yamada): $YAMADA_ID"

SUZUKI_ID=$(create_user "$ADMIN_TOKEN" "suzuki"   "鈴木 一郎" "suzuki@example.com"   "Password1!")
echo "    鈴木 一郎 (suzuki): $SUZUKI_ID"

NAKAMURA_ID=$(create_user "$ADMIN_TOKEN" "nakamura" "中村 さくら" "nakamura@example.com" "Password1!")
echo "    中村 さくら (nakamura): $NAKAMURA_ID"

INOUE_ID=$(create_user "$ADMIN_TOKEN" "inoue"    "井上 大輔" "inoue@example.com"    "Password1!")
echo "    井上 大輔 (inoue): $INOUE_ID"

KATO_ID=$(create_user "$ADMIN_TOKEN" "kato"     "加藤 美咲" "kato@example.com"     "Password1!")
echo "    加藤 美咲 (kato): $KATO_ID"

# ===========================================================
# 3. チームオーナー権限付与
# ===========================================================
echo ""
echo "【3】チームオーナー権限付与..."
patch "$ADMIN_TOKEN" "/api/v1/admin/users/$TANAKA_ID/team-owner" \
  '{"is_team_owner":true,"max_teams":3}' | jq -r '"    tanaka → is_team_owner=" + (.data.is_team_owner|tostring)'
patch "$ADMIN_TOKEN" "/api/v1/admin/users/$YAMADA_ID/team-owner" \
  '{"is_team_owner":true,"max_teams":3}' | jq -r '"    yamada → is_team_owner=" + (.data.is_team_owner|tostring)'
patch "$ADMIN_TOKEN" "/api/v1/admin/users/$SUZUKI_ID/team-owner" \
  '{"is_team_owner":true,"max_teams":2}' | jq -r '"    suzuki → is_team_owner=" + (.data.is_team_owner|tostring)'

# ===========================================================
# 4. チーム作成
# ===========================================================
echo ""
echo "【4】チーム作成..."

TANAKA_TOKEN=$(login "tanaka" "Password1!")
YAMADA_TOKEN=$(login "yamada" "Password1!")

TEAM_A_ID=$(create_team "$TANAKA_TOKEN" \
  "CISSPチーム" \
  "CISSP（Certified Information Systems Security Professional）取得を目指す勉強会です。ドメイン1〜8を順番に攻略し、毎週オンラインで問題を持ち寄って議論します。")
echo "    CISSPチーム (A): $TEAM_A_ID"

TEAM_B_ID=$(create_team "$YAMADA_TOKEN" \
  "SC試験チーム" \
  "情報処理安全確保支援士（SC）合格を目標にしたグループです。午前Ⅱ・午後Ⅰ・午後Ⅱ問題を週次で取り組み、解説と議論を重ねます。")
echo "    SC試験チーム (B): $TEAM_B_ID"

# ===========================================================
# 5. メンバー追加
# ===========================================================
echo ""
echo "【5】メンバー追加..."

# チームA: 田中（オーナー）, 佐藤, 山田, 鈴木, 加藤
add_member "$TANAKA_TOKEN" "$TEAM_A_ID" "$SATO_ID"     > /dev/null && echo "    チームA ← 佐藤 健"
add_member "$TANAKA_TOKEN" "$TEAM_A_ID" "$YAMADA_ID"   > /dev/null && echo "    チームA ← 山田 花子"
add_member "$TANAKA_TOKEN" "$TEAM_A_ID" "$SUZUKI_ID"   > /dev/null && echo "    チームA ← 鈴木 一郎"
add_member "$TANAKA_TOKEN" "$TEAM_A_ID" "$KATO_ID"     > /dev/null && echo "    チームA ← 加藤 美咲"

# チームB: 山田（オーナー）, 中村, 井上, 鈴木
add_member "$YAMADA_TOKEN" "$TEAM_B_ID" "$NAKAMURA_ID" > /dev/null && echo "    チームB ← 中村 さくら"
add_member "$YAMADA_TOKEN" "$TEAM_B_ID" "$INOUE_ID"   > /dev/null && echo "    チームB ← 井上 大輔"
add_member "$YAMADA_TOKEN" "$TEAM_B_ID" "$SUZUKI_ID"  > /dev/null && echo "    チームB ← 鈴木 一郎"

# ===========================================================
# 6. タグ作成
# ===========================================================
echo ""
echo "【6】タグ作成..."

echo "    --- チームA (CISSP) ---"
create_tag "$ADMIN_TOKEN" "$TEAM_A_ID" "暗号化"               | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_A_ID" "アクセス制御"         | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_A_ID" "リスク管理"           | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_A_ID" "セキュリティアーキテクチャ" | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_A_ID" "物理セキュリティ"     | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_A_ID" "インシデント対応"     | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_A_ID" "ソフトウェアセキュリティ" | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_A_ID" "認証"                 | xargs -I{} echo "    {}"

echo "    --- チームB (SC) ---"
create_tag "$ADMIN_TOKEN" "$TEAM_B_ID" "不正アクセス"         | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_B_ID" "マルウェア"           | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_B_ID" "Webセキュリティ"      | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_B_ID" "ネットワークセキュリティ" | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_B_ID" "暗号プロトコル"       | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_B_ID" "脆弱性管理"           | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_B_ID" "ログ・監視"           | xargs -I{} echo "    {}"
create_tag "$ADMIN_TOKEN" "$TEAM_B_ID" "インシデント対応"     | xargs -I{} echo "    {}"

# ===========================================================
# 7. 各ユーザーのトークン取得
# ===========================================================
echo ""
echo "【7】メンバーのログイン..."
SATO_TOKEN=$(login "sato" "Password1!")
SUZUKI_TOKEN=$(login "suzuki" "Password1!")
NAKAMURA_TOKEN=$(login "nakamura" "Password1!")
INOUE_TOKEN=$(login "inoue" "Password1!")
KATO_TOKEN=$(login "kato" "Password1!")
echo "    OK"

# ===========================================================
# 8. 問題作成 — チームA (CISSP)
# ===========================================================
echo ""
echo "【8】問題作成 — チームA (CISSPチーム)..."

Q_A1=$(create_question "$TANAKA_TOKEN" "$TEAM_A_ID" \
  "AES の暗号化モードの違いを説明せよ" \
  "## 問題

AES（Advanced Encryption Standard）を使った対称暗号において、以下の動作モードをそれぞれ説明し、適切なユースケースを述べよ。

1. ECB（Electronic Codebook）モード
2. CBC（Cipher Block Chaining）モード
3. GCM（Galois/Counter Mode）モード" \
  "## 解答

**1. ECB モード**
- 各ブロックを独立して暗号化する
- 同じ平文ブロックは同じ暗号文ブロックになる（パターンが露出する）
- **原則として使用禁止**

**2. CBC モード**
- 前のブロックの暗号文と XOR してから暗号化する
- IV（初期化ベクター）が必要
- ランダムな IV を使えば同じ平文でも毎回異なる暗号文になる
- 完全性保護がないため認証付き暗号との組み合わせが必要

**3. GCM モード**
- CTR モード + GHASH による認証タグ生成
- 認証付き暗号（AEAD）のため機密性と完全性を同時に提供
- TLS 1.3 / IPsec で標準的に使用される" \
  "## 解説

ECB は教科書的な「使ってはいけない暗号」の代表例。Linux のペンギンロゴを ECB で暗号化するとロゴの輪郭がそのまま残る、という有名なデモがある。

CBC は長年使われてきたが、BEAST 攻撃・POODLE 攻撃など IV の取り扱いに起因する攻撃が発見された。

現代では **AES-GCM** が推奨。TLS 1.3 で CBC ベースの暗号スイートは廃止された。

認証付き暗号（AEAD）の重要性：暗号化だけでは中間者によるビット反転攻撃（ビットフリッピング攻撃）に対して無力。" \
  "## 議論点・知らなかった知識

- ECB モードがなぜ危険なのか、ビジュアルで説明できると説得力が増す
- GCM の認証タグは 128 bit が標準。96 bit の nonce（IV）を使い回すと壊滅的に危険（nonce 再利用攻撃）
- AES-GCM-SIV は nonce 再利用耐性を持つ改良版として RFC 8452 で標準化されている" \
  '["暗号化"]' "published")
echo "    Q-A1 作成: $Q_A1"

Q_A2=$(create_question "$SATO_TOKEN" "$TEAM_A_ID" \
  "アクセス制御モデル（MAC / DAC / RBAC）の比較" \
  "## 問題

以下の 3 つのアクセス制御モデルを比較し、それぞれの特徴・長所・短所・適用例を説明せよ。

- **MAC**（Mandatory Access Control）
- **DAC**（Discretionary Access Control）
- **RBAC**（Role-Based Access Control）" \
  "## 解答

| モデル | 制御主体 | 特徴 | 適用例 |
|-------|---------|------|-------|
| MAC | システム | ラベル（機密レベル）で強制制御 | 政府・軍事系システム |
| DAC | リソース所有者 | 所有者が任意に権限付与 | Unix/Windows の標準ファイル権限 |
| RBAC | ロール定義 | ロールに権限をまとめて管理 | 企業情報システム・SaaS |

**MAC** の例：Bell-LaPadula モデル（機密性重視）、Biba モデル（完全性重視）

**DAC** の問題点：トロイの木馬攻撃に弱い（ユーザーが不正プログラムに自分の権限を渡してしまう）

**RBAC** の利点：権限管理が役職単位で直感的。最小権限の原則を組織的に実現しやすい" \
  "## 解説

CISSP ではこの 3 モデルは頻出。重要なのは「誰が権限を決めるか」という視点。

- MAC → 情報のラベルを OS/システムが強制管理（ユーザーは変更不可）
- DAC → ファイルの所有者が chmod などで自由に設定できる → 柔軟だが危険
- RBAC → 役割（Role）ベース → 現代の企業システムで最も一般的

さらに発展として **ABAC**（Attribute-Based AC）もある。ゼロトラストアーキテクチャでは ABAC の考え方が多く使われる。" \
  "## 議論点

- Bell-LaPadula と Biba の「No Read Up / No Write Down」「No Write Up / No Read Down」ルールの暗記法
- 実際のシステムでは複数のモデルを組み合わせることが多い（例：RBAC + MAC）
- 「最小権限の原則（Least Privilege）」はどのモデルでも重要な設計指針" \
  '["アクセス制御"]' "published")
echo "    Q-A2 作成: $Q_A2"

Q_A3=$(create_question "$YAMADA_TOKEN" "$TEAM_A_ID" \
  "定性的リスクアセスメントと定量的リスクアセスメントの違い" \
  "## 問題

情報セキュリティにおけるリスクアセスメントには大きく 2 つのアプローチがある。

1. **定性的リスクアセスメント**（Qualitative Risk Assessment）
2. **定量的リスクアセスメント**（Quantitative Risk Assessment）

それぞれの手法、使用する指標、メリット・デメリットを説明し、代表的な定量的手法の計算式を示せ。" \
  "## 解答

### 定性的リスクアセスメント

- 脅威・脆弱性・影響度を「高/中/低」などの段階で評価
- 専門家の判断・経験に依存
- **メリット**: 実施が速い、データが少なくても実施可能
- **デメリット**: 主観的、異なる評価者で結果がぶれる

### 定量的リスクアセスメント

数値で表現する。主な指標：

| 指標 | 意味 | 計算 |
|-----|------|------|
| AV（Asset Value）| 資産の価値 | — |
| EF（Exposure Factor）| 一事象あたりの損失割合 | 0〜100% |
| SLE（Single Loss Expectancy）| 1 回の事象による期待損失額 | AV × EF |
| ARO（Annual Rate of Occurrence）| 年間発生頻度 | — |
| ALE（Annual Loss Expectancy）| 年間期待損失額 | SLE × ARO |

**例**: サーバー（AV = 1,000 万円）が火災で 40% 損失し、年 0.1 回発生する場合
- SLE = 1,000 万円 × 0.4 = 400 万円
- ALE = 400 万円 × 0.1 = 40 万円" \
  "## 解説

定量的手法の ALE を使うと「対策コストと比較して投資対効果を説明できる」点が重要。

例：ALE = 40 万円のリスクに対して 20 万円の対策を打てば、残余リスクの ALE が下がり投資対効果が成立する。

実務では完全な定量化は難しいため、定性的手法で優先度をつけてから、重要なリスクだけ定量的に掘り下げるハイブリッド手法が多い。" \
  "## 議論点・知らなかった知識

- ALE の計算は CISSP 本試験で出題実績あり。SLE = AV × EF を忘れないこと
- セーフガード選択の判断式: **ALE（before）- ALE（after）- コスト > 0** なら投資価値あり
- 定量化が難しい損失（レピュテーション損害、法的リスク）は定性的評価で補完" \
  '["リスク管理"]' "published")
echo "    Q-A3 作成: $Q_A3"

Q_A4=$(create_question "$SUZUKI_TOKEN" "$TEAM_A_ID" \
  "Kerberos 認証プロトコルの動作フローを説明せよ" \
  "## 問題

Kerberos v5 の認証フローを、以下のコンポーネントを使って順を追って説明せよ。

- クライアント（C）
- 認証サーバー（AS: Authentication Server）
- チケット付与サーバー（TGS: Ticket Granting Server）
- サービスサーバー（SS: Service Server）

また、Kerberos の前提条件と主な弱点も述べよ。" \
  "## 解答

### 認証フロー（6 ステップ）

1. **C → AS**: ユーザー名（平文）で認証要求
2. **AS → C**: TGT（Ticket Granting Ticket）と TGS 用セッションキーを返す（クライアントの秘密鍵で暗号化）
3. **C → TGS**: TGT と利用したいサービス名を提示
4. **TGS → C**: サービスチケット（ST）と SS 用セッションキーを返す
5. **C → SS**: ST を提示してサービス要求
6. **SS → C**: 認証完了、サービス提供

### 前提条件
- 全コンポーネントで時刻同期（±5 分以内）が必要
- KDC（AS + TGS）が単一障害点になりうる

### 主な弱点
- **Pass-the-Ticket 攻撃**: TGT を盗んでリプレイ
- **Kerberoasting**: サービスアカウントのハッシュをオフラインクラック
- **Golden Ticket 攻撃**: KRBTGT アカウントのハッシュ取得で任意の TGT を偽造" \
  "## 解説

Kerberos は Windows Active Directory の認証基盤であり、企業ネットワークで広く使われている。

「チケットを持ち歩く」モデルのため、パスワードをネットワーク上に流さない点が優れている。

Golden Ticket 攻撃は KRBTGT パスワードをリセットして対応するが、実は 2 回リセットが必要（レプリケーションの都合）。" \
  "## 議論点

- Kerberoasting の対策: サービスアカウントに長いランダムパスワードを設定、グループ管理サービスアカウント（gMSA）を使う
- 時刻同期が崩れると認証が通らなくなる → NTP の重要性
- Kerberos と NTLM の使い分け（Kerberos が優先、失敗時に NTLM へフォールバック）" \
  '["認証","セキュリティアーキテクチャ"]' "published")
echo "    Q-A4 作成: $Q_A4"

Q_A5=$(create_question "$KATO_TOKEN" "$TEAM_A_ID" \
  "インシデント対応の 6 フェーズと各フェーズでの行動" \
  "## 問題

NIST SP 800-61 に基づくインシデント対応の 6 フェーズを列挙し、各フェーズで行うべき具体的な行動を説明せよ。また、封じ込め（Containment）フェーズで短期と長期を分ける理由も述べよ。" \
  "## 解答

### インシデント対応の 6 フェーズ

| フェーズ | 内容 |
|---------|------|
| 1. 準備（Preparation）| ポリシー策定、ツール整備、訓練 |
| 2. 検知・分析（Detection & Analysis）| アラート検知、証拠収集、影響範囲特定 |
| 3. 封じ込め（Containment）| 被害拡大を防ぐ（短期・長期） |
| 4. 根絶（Eradication）| マルウェア除去、脆弱性修正 |
| 5. 復旧（Recovery）| サービス再開、監視強化 |
| 6. 事後対応（Post-Incident Activity）| 教訓のまとめ、対策の改善 |

### 短期封じ込めと長期封じ込めを分ける理由

**短期**: 即時に被害を止める（ネットワーク遮断など）→ サービス影響大でも優先
**長期**: 証拠を保全しながら恒久的な対策を施す → フォレンジック調査との並行作業が必要" \
  "## 解説

「根絶の前に必ず証拠を保全する」点が重要。先にシステムをクリーンアップしてしまうと攻撃者の痕跡が失われる。

封じ込めフェーズでは **ネットワーク分離・ACL 追加・アカウント無効化** などの手段を組み合わせる。

事後対応（Lessons Learned）をスキップする組織が多いが、ここが再発防止の要。" \
  "## 議論点・知らなかった知識

- IoC（Indicators of Compromise）の収集と共有（ISACs / STIX / TAXII）
- フォレンジック調査の原則: 揮発性の高いデータ（メモリ）から先に取得
- CSIRT と SOC の役割分担（CSIRT = 意思決定と調整、SOC = 24/7 監視・検知）" \
  '["インシデント対応"]' "published")
echo "    Q-A5 作成: $Q_A5"

Q_A6=$(create_question "$TANAKA_TOKEN" "$TEAM_A_ID" \
  "Bell-LaPadula モデルと Biba モデルの比較" \
  "## 問題

以下の 2 つのセキュリティモデルについて、基本ルールと目的を比較せよ。

1. **Bell-LaPadula モデル**
2. **Biba モデル**

また、なぜこれらが互いに相反する設計になっているかを説明せよ。" \
  "## 解答

### Bell-LaPadula モデル（機密性重視）

- **No Read Up**: 自分より高い機密レベルのデータを読めない
- **No Write Down**: 自分より低い機密レベルへデータを書けない
- **目的**: 機密情報の漏洩防止（上位から下位への情報流出をブロック）

### Biba モデル（完全性重視）

- **No Write Up**: 自分より高い完全性レベルへ書き込めない
- **No Read Down**: 自分より低い完全性レベルのデータを読めない
- **目的**: データの改ざん防止（低信頼データが高信頼データを汚染しないよう）

### なぜ相反するか

Bell-LaPadula では「上位から下位への書き込みは許可」されているが、Biba では「上位への書き込みは禁止」。両方同時に適用するとほぼ全ての操作が制限される。現実には用途に応じてどちらかを選択するか、補完的に組み合わせる。" \
  "## 解説

覚え方：
- Bell-LaPadula = **BLP = 秘密保持（読み上げ禁止・書き下げ禁止）**
- Biba = **完全性（書き上げ禁止・読み下げ禁止）**

軍事システムは BLP を使う（機密漏洩を最優先で防ぐ）。金融システムは Biba を優先することが多い（データの正確性が命）。" \
  "## 議論点

- Clark-Wilson モデル: 商用環境向けの完全性モデル。トランザクションの整合性をルールで担保
- Chinese Wall（Brewer-Nash）モデル: 利益相反防止が目的（コンサルタント等が競合企業の情報にアクセスできないよう）" \
  '["セキュリティアーキテクチャ","アクセス制御"]' "published")
echo "    Q-A6 作成: $Q_A6"

Q_A7=$(create_question "$SATO_TOKEN" "$TEAM_A_ID" \
  "物理セキュリティ：多重防護（Defense in Depth）の設計" \
  "## 問題

データセンターの物理セキュリティを「多重防護（Defense in Depth）」の観点から設計せよ。外周から内部に向けて、各層でどのような対策を講じるべきか。" \
  "## 解答

### 多重防護の各層

**層 1: 外周（Perimeter）**
- フェンス・有刺鉄線・ボラード（車両突入防止）
- 警備員・CCTV

**層 2: 建物外部**
- 施錠されたエントランス
- マンタップ（入退室管理エアロック）
- バッジリーダー（IC カード認証）

**層 3: 建物内部・共用エリア**
- 受付での身分証確認
- 来訪者への入館証発行・エスコート義務

**層 4: サーバールーム**
- 生体認証（指紋・虹彩）+ PIN の多要素認証
- ケージ・ラックの施錠

**層 5: 個別ラック**
- ラック鍵・ケーブルロック
- 動作中の筐体への耐タンパー対策" \
  "## 解説

「多重防護」の本質は、1 つの対策が破られても次の層が機能すること。

マンタップは「1 人ずつしか通れない二重ドア」でテールゲーティング（ピギーバック）を防ぐ。

物理セキュリティを軽視するとソーシャルエンジニアリング（なりすまし入館）で全ての論理的なセキュリティが無力化される。" \
  "## 議論点

- 防犯カメラの映像保存期間（最低 90 日推奨）と証拠としての法的有効性
- 電磁波漏洩対策（TEMPEST / ファラデーケージ）: 軍事・政府施設で適用
- 環境制御（温度・湿度・消火設備）も物理セキュリティの重要要素" \
  '["物理セキュリティ"]' "draft")
echo "    Q-A7 作成: $Q_A7 (下書き)"

# ===========================================================
# 9. 問題作成 — チームB (SC試験)
# ===========================================================
echo ""
echo "【9】問題作成 — チームB (SC試験チーム)..."

Q_B1=$(create_question "$YAMADA_TOKEN" "$TEAM_B_ID" \
  "SQL インジェクションの仕組みと対策" \
  "## 問題

以下のコードに対する SQL インジェクション攻撃の仕組みを説明し、適切な対策を 3 つ以上挙げよ。

\`\`\`sql
SELECT * FROM users WHERE username = '入力値' AND password = '入力値';
\`\`\`

攻撃者が username に \`' OR '1'='1\` を入力した場合、クエリはどう変化するか。" \
  "## 解答

### 攻撃の仕組み

入力値 \`' OR '1'='1\` を代入すると：

\`\`\`sql
SELECT * FROM users WHERE username = '' OR '1'='1' AND password = '';
\`\`\`

演算子の優先順位（AND > OR）に注意が必要だが、さらに \`' OR '1'='1' --\` とすることで AND 以降がコメントアウトされ全件ヒットする。

### 対策

1. **プリペアドステートメント（バインド変数）** — 最も効果的
   \`\`\`python
   cursor.execute('SELECT * FROM users WHERE username = ? AND password = ?', (username, password))
   \`\`\`

2. **ストアドプロシージャ** — パラメータが適切にバインドされる場合のみ有効

3. **入力値のバリデーション** — ホワイトリスト方式で許可文字を限定

4. **エスケープ処理** — シングルクォートを \\'に変換（プリペアドステートメントの代替だが脆弱性が残りやすい）

5. **WAF（Web Application Firewall）** — 既知の攻撃パターンをブロック（多層防御として）" \
  "## 解説

SQL インジェクションは OWASP Top 10 の常連で、SC 試験でも毎年出題される。

プリペアドステートメントが最善策だが、「なぜ安全か」を説明できることが重要：SQL 文の構造（テンプレート）がデータベースに先に解析されるため、後からバインドされる値は SQL 構文として解釈されない。

UNION インジェクション・エラーベース・ブラインドインジェクションなど、手法のバリエーションも把握しておくこと。" \
  "## 議論点・知らなかった知識

- 2 次 SQL インジェクション: データを一度 DB に格納し、後で別の処理で取り出す際に注入が発動する
- ORM を使えば自動的に安全というわけではない。生クエリ（raw query）を使う箇所は要注意
- 最小権限: DB ユーザーに SELECT しか権限を与えなければ INSERT/DROP のリスクを下げられる" \
  '["Webセキュリティ","脆弱性管理"]' "published")
echo "    Q-B1 作成: $Q_B1"

Q_B2=$(create_question "$NAKAMURA_TOKEN" "$TEAM_B_ID" \
  "XSS（クロスサイトスクリプティング）の種類と対策" \
  "## 問題

XSS には大きく 3 種類ある。それぞれの仕組みと特徴を説明し、開発者として実施すべき対策をまとめよ。

1. 反射型 XSS（Reflected XSS）
2. 蓄積型 XSS（Stored XSS）
3. DOM ベース XSS（DOM-based XSS）" \
  "## 解答

### 1. 反射型 XSS

- リクエストに含まれたスクリプトがレスポンスに反射して実行される
- 罠リンクを踏ませる必要がある
- **例**: \`https://example.com/search?q=<script>alert(1)</script>\`

### 2. 蓄積型 XSS

- 攻撃スクリプトが DB などに保存され、閲覧したユーザー全員に実行される
- **最も危険**（持続的に被害が発生する）
- **例**: 掲示板の投稿にスクリプトを仕込む

### 3. DOM ベース XSS

- サーバーは関与せず、ブラウザ上の JS が DOM を操作する際に脆弱性が生じる
- WAF では検知が困難
- **例**: \`document.innerHTML\` に URL のフラグメント（\`#\`以降）を直接挿入

### 対策

| 対策 | 効果 |
|-----|------|
| 出力エスケープ（HTML エンティティ変換）| 反射型・蓄積型に有効 |
| Content Security Policy（CSP）ヘッダー | インラインスクリプト実行をブロック |
| HttpOnly Cookie | スクリプトから Cookie を読めなくする |
| \`innerHTML\` の代わりに \`textContent\` 使用 | DOM ベース XSS に有効 |
| DOMPurify 等のサニタイズライブラリ | HTML を許可したい場合 |" \
  "## 解説

「エスケープは出力時に行う」が鉄則。入力時に除去しようとするアプローチは二重エスケープや迂回の温床になる。

CSP は XSS の緩和策として強力だが、設定が複雑で実運用で \`unsafe-inline\` を許可してしまうケースが多い。

DOM ベース XSS はサーバー側のコードレビューだけでは見つからない → フロントエンドのコードレビューが必要。" \
  "## 議論点

- Trusted Types API: DOM 操作に型付けされた値のみ許可する仕組み（Chrome で実装済み）
- mXSS（mutation XSS）: サニタイズ後に DOM が変化してスクリプトが復活するケース
- Self-XSS: ユーザー自身が騙されてコンソールにスクリプトを入力させられる攻撃" \
  '["Webセキュリティ","脆弱性管理"]' "published")
echo "    Q-B2 作成: $Q_B2"

Q_B3=$(create_question "$INOUE_TOKEN" "$TEAM_B_ID" \
  "TLS ハンドシェイクの流れ（TLS 1.3）" \
  "## 問題

TLS 1.3 のハンドシェイクフローを説明し、TLS 1.2 と比べた際の主な改善点を 3 つ挙げよ。また **前方秘匿性（PFS: Perfect Forward Secrecy）** とは何か説明せよ。" \
  "## 解答

### TLS 1.3 ハンドシェイクフロー（1-RTT）

\`\`\`
Client                          Server
  |                               |
  |-- ClientHello ─────────────→  |  (対応暗号スイート・鍵共有パラメータ)
  |← ServerHello ─────────────── |  (選択した暗号スイート・鍵共有)
  |← {EncryptedExtensions} ───── |
  |← {Certificate} ────────────── |
  |← {CertificateVerify} ──────── |
  |← {Finished} ───────────────── |
  |-- {Finished} ──────────────→  |
  |== アプリケーションデータ通信 ==|
\`\`\`

TLS 1.3 では **最初から暗号化**（ServerHello 以降は暗号化済み）。

### TLS 1.2 との比較

| 項目 | TLS 1.2 | TLS 1.3 |
|-----|---------|---------|
| ハンドシェイク | 2-RTT | 1-RTT（0-RTT も可能） |
| 前方秘匿性 | オプション | **必須**（ECDHE/DHE のみ） |
| 廃止された機能 | RSA 鍵交換, RC4, DES, MD5, SHA-1 | 廃止 |

### 前方秘匿性（PFS）

セッションごとに一時的な鍵を生成し、後から長期秘密鍵が漏洩しても過去のセッションを復号できない性質。

- TLS 1.2 の RSA 鍵交換: サーバーの秘密鍵が漏洩 → 過去の通信がすべて復号される
- TLS 1.3 の ECDHE: セッション鍵は通信後に破棄 → 過去の通信は保護される" \
  "## 解説

TLS 1.3 は 2018 年に RFC 8446 として標準化。1.2 の様々な既知攻撃（POODLE, BEAST, CRIME 等）への対策が設計に組み込まれている。

0-RTT（Early Data）は前のセッション情報を使って即座にデータ送信できるが、リプレイ攻撃の危険性があるため冪等な操作（GET リクエスト等）にのみ使用すべき。" \
  "## 議論点

- Certificate Transparency（CT）: 証明書の透明性ログで不正な証明書の発行を検知
- OCSP Stapling: 証明書失効確認をリアルタイムに行う仕組み（OCSP レスポンスをサーバーがキャッシュして付与）
- 量子コンピュータ時代に備えた PQC（Post-Quantum Cryptography）: NIST が ML-KEM(Kyber) 等を標準化" \
  '["暗号プロトコル","ネットワークセキュリティ"]' "published")
echo "    Q-B3 作成: $Q_B3"

Q_B4=$(create_question "$SUZUKI_TOKEN" "$TEAM_B_ID" \
  "CSRF（クロスサイトリクエストフォージェリ）攻撃と防御" \
  "## 問題

CSRF 攻撃の仕組みを具体的なシナリオで説明し、効果的な防御策をそれぞれの原理とともに述べよ。また XSS との違いも明確にせよ。" \
  "## 解答

### CSRF 攻撃のシナリオ

1. 被害者がオンラインバンクにログイン中（Cookie でセッション維持）
2. 攻撃者が罠サイトに以下を仕込む：
   \`\`\`html
   <img src=\"https://bank.example.com/transfer?to=attacker&amount=100000\">
   \`\`\`
3. 被害者が罠サイトを閲覧 → ブラウザが自動的に Cookie 付きリクエストを送信
4. バンクは正規ユーザーからのリクエストと判断し送金を実行

### 防御策

| 対策 | 原理 |
|-----|------|
| **CSRF トークン** | フォームに予測不能なトークンを埋め込み、サーバーで検証 |
| **SameSite Cookie 属性** | \`Strict\`: 完全にクロスサイトリクエストで Cookie を送らない<br>\`Lax\`: TOP-LEVEL な GET のみ許可 |
| **カスタムリクエストヘッダー** | \`X-Requested-With: XMLHttpRequest\` をチェック（CORS プリフライトが保護） |
| **二重送信 Cookie** | トークンを Cookie と hidden フィールド両方に含め一致確認 |
| **Origin / Referer ヘッダー検証** | リクエスト元ドメインを確認（補助的対策）|

### XSS との違い

| | CSRF | XSS |
|-|------|-----|
| 攻撃対象 | サーバーに対するリクエスト | ブラウザ上でのスクリプト実行 |
| 認証の悪用 | 被害者の認証情報を利用 | 被害者のブラウザを乗っ取る |
| Cookie | ブラウザが自動送信 | HttpOnly があれば直接読めない |" \
  "## 解説

SameSite=Strict が最も強力だが、「別ドメインからのリンクでも Cookie が送られない」ため、外部サイトからのリンクで認証維持が必要なサービスでは UX への影響が大きい。

多くのモダンブラウザは \`SameSite=Lax\` をデフォルトで適用するようになっているが、明示的に設定することを推奨。" \
  "## 議論点

- ログアウト CSRF: 被害者を強制ログアウトさせる攻撃（比較的軽微だが対策は同じ）
- JSON API は CSRF に比較的安全（Content-Type: application/json はプリフライトが発生するため）だが、text/plain で JSON を受け付けるなら危険
- CORS の設定ミス（Access-Control-Allow-Origin: *）は CSRF 対策を無効化しうる" \
  '["Webセキュリティ","不正アクセス"]' "published")
echo "    Q-B4 作成: $Q_B4"

Q_B5=$(create_question "$YAMADA_TOKEN" "$TEAM_B_ID" \
  "マルウェアの分類と各種対策技術" \
  "## 問題

以下のマルウェアの種類について、それぞれの特徴・感染経路・代表的な対策を説明せよ。

1. ランサムウェア
2. ルートキット
3. ワーム
4. RAT（Remote Access Trojan）" \
  "## 解答

### 1. ランサムウェア

- **特徴**: ファイルを暗号化して身代金を要求
- **感染経路**: フィッシングメール添付ファイル、RDP 脆弱性、サプライチェーン攻撃
- **対策**: 定期バックアップ（3-2-1 ルール）、EDR 導入、ネットワーク分離

### 2. ルートキット

- **特徴**: OS のカーネルレベルに潜伏し、自身や他のマルウェアを隠蔽
- **感染経路**: ブートローダー改ざん、カーネルモジュール挿入
- **対策**: セキュアブート（UEFI）、定期的なハッシュ検証、再インストール

### 3. ワーム

- **特徴**: ホストを必要とせず自己複製して拡散する
- **感染経路**: ネットワーク脆弱性（MS17-010 等）、USB
- **対策**: パッチ管理（迅速適用）、ネットワーク分離、ポートフィルタリング

### 4. RAT（Remote Access Trojan）

- **特徴**: 攻撃者にリモートコントロール手段を提供
- **感染経路**: フィッシング、水飲み場攻撃
- **対策**: EDR/XDR による振る舞い検知、通信の監視（外部への異常な通信）" \
  "## 解説

3-2-1 バックアップルール：3 つのコピー、2 種類のメディア、1 つはオフサイト（ネットワーク切り離し）。ランサムウェアはバックアップも暗号化しようとするため、オフライン・エアギャップバックアップが重要。

ルートキットの検出は極めて困難。感染が確認されたら「クリーンアップ」ではなく「OS 再インストール」が原則。" \
  "## 議論点

- APT（Advanced Persistent Threat）: 国家関与のマルウェアは Living-off-the-Land（LOLBins）で正規ツールを悪用するため EDR でも検知困難
- PolyMorphic/Metamorphic Malware: コードが変形するためシグネチャベース検知が効かない
- サンドボックス解析: 動的解析でマルウェアの振る舞いを把握（VM 検知を回避するサンプルも多い）" \
  '["マルウェア"]' "published")
echo "    Q-B5 作成: $Q_B5"

Q_B6=$(create_question "$NAKAMURA_TOKEN" "$TEAM_B_ID" \
  "ファイアウォールとIDS/IPSの違いと組み合わせ方" \
  "## 問題

ファイアウォール、IDS（侵入検知システム）、IPS（侵入防止システム）の違いを説明し、企業ネットワークでどのように配置・組み合わせるべきかを述べよ。" \
  "## 解答

### 各機器の役割

| 機器 | 動作 | トラフィックへの影響 |
|-----|------|-----------------|
| ファイアウォール | パケットフィルタリング（IP/Port ベース） | インラインに配置、遮断可能 |
| IDS | トラフィックを監視・アラート通知 | **インラインではなくコピーを監視（受動的）** |
| IPS | IDS + 自動遮断 | **インラインに配置（能動的）** |

### 配置パターン（企業ネットワーク）

\`\`\`
インターネット
     ↓
[外部 FW] ← L3/L4 フィルタリング
     ↓
   [DMZ] ← Web サーバー等を配置
     ↓
[内部 FW + IPS] ← 深い検査・自動遮断
     ↓
[内部ネットワーク]
[IDS（コピー監視）]← SPAN ポートでトラフィックをコピー
\`\`\`

IPS はインラインで誤検知（False Positive）があるとサービス影響が出るため、初期は **IDS モード**で運用 → チューニング後に **IPS モード**に切り替えるのが一般的。" \
  "## 解説

次世代 FW（NGFW）は L7 までの検査ができ、IPS 機能を内蔵したものが多い。

SIEM と組み合わせると IDS/IPS・FW のログを集約して相関分析ができる。単体のアラートでは気づきにくい攻撃パターンを検出できる。" \
  "## 議論点

- UTM vs NGFW: UTM はオールインワン（SMB 向け）、NGFW はエンタープライズ向けで高性能
- ゼロデイ攻撃: シグネチャベースの IDS/IPS では防げない → サンドボックス・振る舞い検知が必要
- 暗号化トラフィック（HTTPS）の検査: SSL インスペクション（中間で復号）はプライバシーとのバランスが必要" \
  '["ネットワークセキュリティ","不正アクセス"]' "published")
echo "    Q-B6 作成: $Q_B6"

Q_B7=$(create_question "$INOUE_TOKEN" "$TEAM_B_ID" \
  "ログ分析：不正アクセスの痕跡を特定せよ" \
  "## 問題

以下の Web サーバーアクセスログ（抜粋）から、不正アクセスの痕跡を特定し、攻撃者が実施した攻撃の種類と対策を述べよ。

\`\`\`
203.0.113.5 - - [21/Jun/2026:03:12:01 +0900] \"GET /login HTTP/1.1\" 200 1234
203.0.113.5 - - [21/Jun/2026:03:12:03 +0900] \"POST /login HTTP/1.1\" 401 256
203.0.113.5 - - [21/Jun/2026:03:12:05 +0900] \"POST /login HTTP/1.1\" 401 256
(同じIPから03:12〜03:13の1分間にPOST /loginが248回)
203.0.113.5 - - [21/Jun/2026:03:13:58 +0900] \"POST /login HTTP/1.1\" 200 4521
203.0.113.5 - - [21/Jun/2026:03:14:02 +0900] \"GET /admin/users HTTP/1.1\" 200 9873
203.0.113.5 - - [21/Jun/2026:03:14:15 +0900] \"GET /admin/export?format=csv HTTP/1.1\" 200 589234
\`\`\`" \
  "## 解答

### 攻撃の特定

**1. ブルートフォース攻撃（パスワード総当たり）**
- 1 分間に 248 回の POST /login （401 連続）
- 最終的に 200 OK → **認証突破成功**

**2. 不正アクセス後の情報窃取**
- 管理者ページ（/admin/users）の閲覧 → **権限昇格またはアカウント乗っ取り**
- /admin/export でCSV（589KB）をダウンロード → **データ漏洩の疑い**

### 推奨対策

| 問題 | 対策 |
|-----|------|
| ブルートフォース | アカウントロックアウト（5 回失敗で 30 分ロック）|
| 同一 IP の大量リクエスト | レート制限（Rate Limiting）、IP ブロック |
| 認証強化 | MFA（多要素認証）の必須化 |
| 管理画面保護 | IP アドレス制限、管理画面を別ドメイン/パスに隔離 |
| ログ監視 | SIEM でリアルタイムアラート設定（同一 IP から X 回失敗等）|" \
  "## 解説

このログパターンは試験でも実務でも頻出。「短時間に大量の 401 → 突然の 200」というパターンは即座にアラートが上がるべき。

ブルートフォース対策のアカウントロックアウトは DoS にも使われる（存在するアカウントを意図的にロックさせる）。CAPTCHA や指数バックオフのほうが現実的なケースも多い。" \
  "## 議論点

- ログの完全性保護: 攻撃者がログを改ざんした場合に備えて、リモートの syslog サーバーや WORM ストレージへ転送
- タイムゾーンの統一: ログ相関分析では UTC 統一が必須
- 589KB の CSV: 個人情報 DB ならば GDPR・個人情報保護法のインシデント報告義務が発生する可能性" \
  '["ログ・監視","不正アクセス","インシデント対応"]' "published")
echo "    Q-B7 作成: $Q_B7"

# ===========================================================
# 10. コメント投稿
# ===========================================================
echo ""
echo "【10】コメント投稿..."

# チームA の質問へのコメント
post_comment "$SATO_TOKEN"    "$TEAM_A_ID" "$Q_A1" "ECB の「ペンギン問題」は視覚的で分かりやすいですね。GCM の nonce 再利用攻撃は実際に 2016 年の TLS 実装で問題になったケースがありました。" > /dev/null
post_comment "$YAMADA_TOKEN"  "$TEAM_A_ID" "$Q_A1" "AES-GCM-SIV はまだ普及途上ですが、LibreSSL などで実装が進んでいます。試験には出なくても知っておく価値がありそうです。" > /dev/null
post_comment "$TANAKA_TOKEN"  "$TEAM_A_ID" "$Q_A1" "TLS 1.3 では AES-GCM が必須になっているので、このあたりの理解は実務でも直結しますね。" > /dev/null

post_comment "$TANAKA_TOKEN"  "$TEAM_A_ID" "$Q_A2" "「DAC は所有者が権限を管理する」という点が混乱しやすいですが、Unix の chmod がまさに DAC の実装ですね。" > /dev/null
post_comment "$KATO_TOKEN"    "$TEAM_A_ID" "$Q_A2" "ABAC をゼロトラストと結びつけて理解できました。属性（ユーザー・リソース・環境・時刻など）で動的に判断するイメージです。" > /dev/null

post_comment "$SATO_TOKEN"    "$TEAM_A_ID" "$Q_A3" "ALE の計算式は本番でも使えるよう手を動かして練習しておきます。SLE = AV × EF をまず頭に入れます。" > /dev/null
post_comment "$SUZUKI_TOKEN"  "$TEAM_A_ID" "$Q_A3" "「ALE(before) - ALE(after) - コスト > 0」の式、投資判断で使えて実用的ですね。経営層への説明に使えそうです。" > /dev/null

post_comment "$YAMADA_TOKEN"  "$TEAM_A_ID" "$Q_A4" "Golden Ticket 攻撃はドメインコントローラーが侵害された場合のシナリオですね。対策が KRBTGT パスワードリセット × 2 というのも覚えておきます。" > /dev/null
post_comment "$KATO_TOKEN"    "$TEAM_A_ID" "$Q_A4" "Kerberoasting は攻撃ツール（Rubeus 等）で簡単に実行できるため、現実のペネトレテストでも頻繁に使われる手法です。" > /dev/null

# チームB の質問へのコメント
post_comment "$NAKAMURA_TOKEN" "$TEAM_B_ID" "$Q_B1" "ORM でも raw クエリの部分は要注意というのは盲点でした。ActiveRecord の \`where\` に文字列補間を使うと危険という話と同じですね。" > /dev/null
post_comment "$INOUE_TOKEN"    "$TEAM_B_ID" "$Q_B1" "2 次 SQL インジェクションは初めて知りました。「一度保存した値が後で SQL として解釈される」というシナリオは気をつけます。" > /dev/null
post_comment "$SUZUKI_TOKEN"   "$TEAM_B_ID" "$Q_B1" "WAF も完全ではないので、アプリ側の対策（プリペアドステートメント）が必須ですね。WAF は多層防御の一つとして。" > /dev/null

post_comment "$INOUE_TOKEN"    "$TEAM_B_ID" "$Q_B2" "DOM ベース XSS は \`innerHTML\` 以外にも \`document.write\` や \`eval\` も危険ですね。フロントエンドのレビューが重要だと改めて思いました。" > /dev/null
post_comment "$YAMADA_TOKEN"   "$TEAM_B_ID" "$Q_B2" "CSP ヘッダーの設定は実際に試してみると難しいですね。\`report-only\` モードで段階的に導入するアプローチが実用的です。" > /dev/null

post_comment "$SUZUKI_TOKEN"   "$TEAM_B_ID" "$Q_B3" "0-RTT のリプレイ攻撃リスクは盲点でした。GET でも冪等でない操作（残高照会後の記録等）には使わない方が安全ですね。" > /dev/null
post_comment "$NAKAMURA_TOKEN" "$TEAM_B_ID" "$Q_B3" "PQC のあたりは SC 試験にも出始めているようです。CRYSTALS-Kyber（ML-KEM）が NIST 標準になったのを覚えておきます。" > /dev/null

post_comment "$YAMADA_TOKEN"   "$TEAM_B_ID" "$Q_B4" "SameSite=Lax のデフォルト化で CSRF 攻撃は減りましたが、明示的な設定とトークン検証の組み合わせが確実ですね。" > /dev/null

post_comment "$NAKAMURA_TOKEN" "$TEAM_B_ID" "$Q_B5" "3-2-1 バックアップはランサムウェア対策の基本ですが、オフサイト（クラウド含む）への転送が漏れているケースが多いですね。" > /dev/null
post_comment "$INOUE_TOKEN"    "$TEAM_B_ID" "$Q_B5" "LOLBins（Living-off-the-Land）は正規ツール（PowerShell, WMI 等）を悪用するため検知が難しい。振る舞い検知が重要です。" > /dev/null

post_comment "$SUZUKI_TOKEN"   "$TEAM_B_ID" "$Q_B7" "このログ問題、実際の試験形式に近くて良い練習になりました。タイムスタンプの読み方や 401/200 の意味を素早く判断できるようにしたいです。" > /dev/null
post_comment "$YAMADA_TOKEN"   "$TEAM_B_ID" "$Q_B7" "589KB の CSV ダウンロードは確かに GDPR 違反になりそうです。インシデント対応後の報告義務（72 時間以内）も押さえておきます。" > /dev/null

# ===========================================================
# 完了
# ===========================================================
echo ""
echo "=========================================="
echo "  テストデータ投入完了！"
echo "=========================================="
echo ""
echo "ユーザー一覧:"
echo "  admin        / Admin1234!    (admin)"
echo "  tanaka       / Password1!    (CISSPチーム オーナー)"
echo "  sato         / Password1!    (CISSPチーム メンバー)"
echo "  yamada       / Password1!    (SC試験チーム オーナー・CISSPチームメンバー)"
echo "  suzuki       / Password1!    (両チーム メンバー)"
echo "  nakamura     / Password1!    (SC試験チーム メンバー)"
echo "  inoue        / Password1!    (SC試験チーム メンバー)"
echo "  kato         / Password1!    (CISSPチーム メンバー)"
echo ""
echo "チーム:"
echo "  CISSPチーム  (ID: $TEAM_A_ID)"
echo "  SC試験チーム (ID: $TEAM_B_ID)"
echo ""
echo "アクセス: http://localhost:3000"
