#!/usr/bin/env python3
"""
テストデータ投入スクリプト
使い方: python3 scripts/seed.py
前提: make up でアプリが起動済みであること
"""

import json
import urllib.request
import urllib.error
import sys

BASE_URL = "http://localhost:8080"


def req(method, path, data=None, token=None):
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    body = json.dumps(data).encode() if data else None
    r = urllib.request.Request(f"{BASE_URL}{path}", data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(r) as resp:
            return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        print(f"  ERROR {e.code} {method} {path}: {body}", file=sys.stderr)
        raise


def login(username, password):
    resp = req("POST", "/api/v1/auth/login", {"username": username, "password": password})
    return resp["data"]["token"]


def create_user(token, username, display_name, email, password, role="user"):
    resp = req("POST", "/api/v1/users",
               {"username": username, "display_name": display_name,
                "email": email, "password": password, "role": role}, token)
    return resp["data"]["id"]


def grant_team_owner(token, user_id, max_teams=3):
    resp = req("PATCH", f"/api/v1/admin/users/{user_id}/team-owner",
               {"is_team_owner": True, "max_teams": max_teams}, token)
    return resp["data"]["is_team_owner"]


def create_team(token, name, description):
    resp = req("POST", "/api/v1/teams", {"name": name, "description": description}, token)
    return resp["data"]["id"]


def add_member(token, team_id, user_id):
    resp = req("POST", f"/api/v1/teams/{team_id}/members", {"user_id": user_id}, token)
    return resp["data"]["role"]


def create_tag(token, team_id, name):
    resp = req("POST", f"/api/v1/teams/{team_id}/tags", {"name": name}, token)
    return resp["data"]["id"]


def create_question(token, team_id, title, body, answer, explanation, memo, tags, status="published"):
    resp = req("POST", f"/api/v1/teams/{team_id}/questions", {
        "title": title, "body": body, "answer": answer,
        "explanation": explanation, "memo": memo,
        "tags": tags, "status": status,
    }, token)
    return resp["data"]["id"]


def post_comment(token, team_id, question_id, body):
    resp = req("POST", f"/api/v1/teams/{team_id}/questions/{question_id}/comments",
               {"body": body}, token)
    return resp["data"]["id"]


print("==========================================")
print("  cert-study-base テストデータ投入")
print("==========================================")

# ============================================================
# 1. admin ログイン
# ============================================================
print("\n【1】admin ログイン...")
ADMIN = login("admin", "Admin1234!")
print("    OK")

# ============================================================
# 2. ユーザー作成
# ============================================================
print("\n【2】ユーザー作成...")
TANAKA_ID   = create_user(ADMIN, "tanaka",   "田中 優",     "tanaka@example.com",   "Password1!")
print(f"    田中 優 (tanaka): {TANAKA_ID}")
SATO_ID     = create_user(ADMIN, "sato",     "佐藤 健",     "sato@example.com",     "Password1!")
print(f"    佐藤 健 (sato): {SATO_ID}")
YAMADA_ID   = create_user(ADMIN, "yamada",   "山田 花子",   "yamada@example.com",   "Password1!")
print(f"    山田 花子 (yamada): {YAMADA_ID}")
SUZUKI_ID   = create_user(ADMIN, "suzuki",   "鈴木 一郎",   "suzuki@example.com",   "Password1!")
print(f"    鈴木 一郎 (suzuki): {SUZUKI_ID}")
NAKAMURA_ID = create_user(ADMIN, "nakamura", "中村 さくら", "nakamura@example.com", "Password1!")
print(f"    中村 さくら (nakamura): {NAKAMURA_ID}")
INOUE_ID    = create_user(ADMIN, "inoue",    "井上 大輔",   "inoue@example.com",    "Password1!")
print(f"    井上 大輔 (inoue): {INOUE_ID}")
KATO_ID     = create_user(ADMIN, "kato",     "加藤 美咲",   "kato@example.com",     "Password1!")
print(f"    加藤 美咲 (kato): {KATO_ID}")

# ============================================================
# 3. チームオーナー権限付与
# ============================================================
print("\n【3】チームオーナー権限付与...")
grant_team_owner(ADMIN, TANAKA_ID,  max_teams=3)
print("    tanaka → is_team_owner=True")
grant_team_owner(ADMIN, YAMADA_ID,  max_teams=3)
print("    yamada → is_team_owner=True")
grant_team_owner(ADMIN, SUZUKI_ID,  max_teams=2)
print("    suzuki → is_team_owner=True")

# ============================================================
# 4. チーム作成
# ============================================================
print("\n【4】チーム作成...")
TANAKA_T = login("tanaka", "Password1!")
YAMADA_T = login("yamada", "Password1!")

TEAM_A = create_team(TANAKA_T,
    "CISSPチーム",
    "CISSP（Certified Information Systems Security Professional）取得を目指す勉強会です。"
    "ドメイン1〜8を順番に攻略し、毎週オンラインで問題を持ち寄って議論します。")
print(f"    CISSPチーム (A): {TEAM_A}")

TEAM_B = create_team(YAMADA_T,
    "SC試験チーム",
    "情報処理安全確保支援士（SC）合格を目標にしたグループです。"
    "午前Ⅱ・午後Ⅰ・午後Ⅱ問題を週次で取り組み、解説と議論を重ねます。")
print(f"    SC試験チーム (B): {TEAM_B}")

# ============================================================
# 5. メンバー追加
# ============================================================
print("\n【5】メンバー追加...")
SATO_T     = login("sato",     "Password1!")
SUZUKI_T   = login("suzuki",   "Password1!")
NAKAMURA_T = login("nakamura", "Password1!")
INOUE_T    = login("inoue",    "Password1!")
KATO_T     = login("kato",     "Password1!")

# チームA: 田中(オーナー), 佐藤, 山田, 鈴木, 加藤
for uid, name in [(SATO_ID, "佐藤 健"), (YAMADA_ID, "山田 花子"),
                  (SUZUKI_ID, "鈴木 一郎"), (KATO_ID, "加藤 美咲")]:
    add_member(TANAKA_T, TEAM_A, uid)
    print(f"    チームA ← {name}")

# チームB: 山田(オーナー), 中村, 井上, 鈴木
for uid, name in [(NAKAMURA_ID, "中村 さくら"), (INOUE_ID, "井上 大輔"), (SUZUKI_ID, "鈴木 一郎")]:
    add_member(YAMADA_T, TEAM_B, uid)
    print(f"    チームB ← {name}")

# ============================================================
# 6. タグ作成
# ============================================================
print("\n【6】タグ作成...")
# タグIDをname→IDの辞書で保持し、問題作成時にIDを渡す
TAGS_A = {}
print("    --- チームA (CISSP) ---")
for tag in ["暗号化", "アクセス制御", "リスク管理", "セキュリティアーキテクチャ",
            "物理セキュリティ", "インシデント対応", "ソフトウェアセキュリティ", "認証"]:
    TAGS_A[tag] = create_tag(ADMIN, TEAM_A, tag)
    print(f"    {tag}")

TAGS_B = {}
print("    --- チームB (SC) ---")
for tag in ["不正アクセス", "マルウェア", "Webセキュリティ", "ネットワークセキュリティ",
            "暗号プロトコル", "脆弱性管理", "ログ・監視", "インシデント対応"]:
    TAGS_B[tag] = create_tag(ADMIN, TEAM_B, tag)
    print(f"    {tag}")

# ============================================================
# 7. 問題作成 — チームA (CISSP)
# ============================================================
print("\n【7】問題作成 — チームA (CISSPチーム)...")

Q_A1 = create_question(TANAKA_T, TEAM_A,
    "AES の暗号化モードの違いを説明せよ",
    """\
## 問題

AES（Advanced Encryption Standard）を使った対称暗号において、以下の動作モードをそれぞれ説明し、適切なユースケースを述べよ。

1. ECB（Electronic Codebook）モード
2. CBC（Cipher Block Chaining）モード
3. GCM（Galois/Counter Mode）モード""",
    """\
## 解答

**1. ECB モード**
- 各ブロックを独立して暗号化する
- 同じ平文ブロックは同じ暗号文ブロックになる（パターンが露出する）
- **原則として使用禁止**

**2. CBC モード**
- 前のブロックの暗号文と XOR してから暗号化する
- IV（初期化ベクター）が必要
- ランダムな IV を使えば同じ平文でも毎回異なる暗号文になる

**3. GCM モード**
- CTR モード + GHASH による認証タグ生成
- 認証付き暗号（AEAD）のため機密性と完全性を同時に提供
- TLS 1.3 / IPsec で標準的に使用される""",
    """\
## 解説

ECB は教科書的な「使ってはいけない暗号」の代表例。Linux のペンギンロゴを ECB で暗号化するとロゴの輪郭がそのまま残る、という有名なデモがある。

CBC は長年使われてきたが、BEAST 攻撃・POODLE 攻撃など IV の取り扱いに起因する攻撃が発見された。

現代では **AES-GCM** が推奨。TLS 1.3 で CBC ベースの暗号スイートは廃止された。""",
    """\
## 議論点・知らなかった知識

- GCM の認証タグは 128 bit が標準。96 bit の nonce（IV）を使い回すと壊滅的に危険（nonce 再利用攻撃）
- AES-GCM-SIV は nonce 再利用耐性を持つ改良版として RFC 8452 で標準化されている
- TLS 1.3 では AES-GCM が必須暗号スイートに含まれる""",
    [TAGS_A["暗号化"]], "published")
print(f"    Q-A1: AES暗号化モード → {Q_A1}")

Q_A2 = create_question(SATO_T, TEAM_A,
    "アクセス制御モデル（MAC / DAC / RBAC）の比較",
    """\
## 問題

以下の 3 つのアクセス制御モデルを比較し、それぞれの特徴・長所・短所・適用例を説明せよ。

- **MAC**（Mandatory Access Control）
- **DAC**（Discretionary Access Control）
- **RBAC**（Role-Based Access Control）""",
    """\
## 解答

| モデル | 制御主体 | 特徴 | 適用例 |
|-------|---------|------|-------|
| MAC | システム | ラベル（機密レベル）で強制制御 | 政府・軍事系システム |
| DAC | リソース所有者 | 所有者が任意に権限付与 | Unix/Windows の標準ファイル権限 |
| RBAC | ロール定義 | ロールに権限をまとめて管理 | 企業情報システム・SaaS |

**MAC** の例：Bell-LaPadula モデル（機密性重視）、Biba モデル（完全性重視）

**DAC** の問題点：トロイの木馬攻撃に弱い（ユーザーが不正プログラムに自分の権限を渡してしまう）

**RBAC** の利点：権限管理が役職単位で直感的。最小権限の原則を組織的に実現しやすい""",
    """\
## 解説

CISSP ではこの 3 モデルは頻出。重要なのは「誰が権限を決めるか」という視点。

- MAC → 情報のラベルを OS/システムが強制管理（ユーザーは変更不可）
- DAC → ファイルの所有者が chmod などで自由に設定できる → 柔軟だが危険
- RBAC → 役割（Role）ベース → 現代の企業システムで最も一般的

さらに発展として **ABAC**（Attribute-Based AC）もある。ゼロトラストアーキテクチャでは ABAC の考え方が多く使われる。""",
    """\
## 議論点

- Bell-LaPadula と Biba の「No Read Up / No Write Down」ルールの暗記法
- 実際のシステムでは複数のモデルを組み合わせることが多い（例：RBAC + MAC）
- 「最小権限の原則（Least Privilege）」はどのモデルでも重要な設計指針""",
    [TAGS_A["アクセス制御"]], "published")
print(f"    Q-A2: アクセス制御モデル → {Q_A2}")

Q_A3 = create_question(YAMADA_T, TEAM_A,
    "定性的リスクアセスメントと定量的リスクアセスメントの違い",
    """\
## 問題

情報セキュリティにおけるリスクアセスメントには大きく 2 つのアプローチがある。

1. **定性的リスクアセスメント**（Qualitative Risk Assessment）
2. **定量的リスクアセスメント**（Quantitative Risk Assessment）

それぞれの手法、使用する指標、メリット・デメリットを説明し、代表的な定量的手法の計算式を示せ。""",
    """\
## 解答

### 定性的リスクアセスメント

- 脅威・脆弱性・影響度を「高/中/低」などの段階で評価
- 専門家の判断・経験に依存
- **メリット**: 実施が速い、データが少なくても実施可能
- **デメリット**: 主観的、異なる評価者で結果がぶれる

### 定量的リスクアセスメント

| 指標 | 意味 | 計算 |
|-----|------|------|
| AV（Asset Value）| 資産の価値 | — |
| EF（Exposure Factor）| 一事象あたりの損失割合 | 0〜100% |
| SLE（Single Loss Expectancy）| 1 回の事象による期待損失額 | AV × EF |
| ARO（Annual Rate of Occurrence）| 年間発生頻度 | — |
| ALE（Annual Loss Expectancy）| 年間期待損失額 | SLE × ARO |

**例**: サーバー（AV = 1,000 万円）が火災で 40% 損失し、年 0.1 回発生する場合
- SLE = 1,000 万円 × 0.4 = 400 万円
- ALE = 400 万円 × 0.1 = 40 万円""",
    """\
## 解説

定量的手法の ALE を使うと「対策コストと比較して投資対効果を説明できる」点が重要。

セーフガード選択の判断式: **ALE（before）- ALE（after）- コスト > 0** なら投資価値あり

実務では完全な定量化は難しいため、定性的手法で優先度をつけてから、重要なリスクだけ定量的に掘り下げるハイブリッド手法が多い。""",
    """\
## 議論点・知らなかった知識

- ALE の計算は CISSP 本試験で出題実績あり。SLE = AV × EF を忘れないこと
- 定量化が難しい損失（レピュテーション損害、法的リスク）は定性的評価で補完
- FAIR（Factor Analysis of Information Risk）は定量的リスク分析のフレームワーク""",
    [TAGS_A["リスク管理"]], "published")
print(f"    Q-A3: リスクアセスメント → {Q_A3}")

Q_A4 = create_question(SUZUKI_T, TEAM_A,
    "Kerberos 認証プロトコルの動作フローを説明せよ",
    """\
## 問題

Kerberos v5 の認証フローを、以下のコンポーネントを使って順を追って説明せよ。

- クライアント（C）
- 認証サーバー（AS: Authentication Server）
- チケット付与サーバー（TGS: Ticket Granting Server）
- サービスサーバー（SS: Service Server）

また、Kerberos の前提条件と主な弱点も述べよ。""",
    """\
## 解答

### 認証フロー（6 ステップ）

1. **C → AS**: ユーザー名（平文）で認証要求
2. **AS → C**: TGT（Ticket Granting Ticket）と TGS 用セッションキーを返す
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
- **Golden Ticket 攻撃**: KRBTGT アカウントのハッシュ取得で任意の TGT を偽造""",
    """\
## 解説

Kerberos は Windows Active Directory の認証基盤であり、企業ネットワークで広く使われている。

「チケットを持ち歩く」モデルのため、パスワードをネットワーク上に流さない点が優れている。

Golden Ticket 攻撃は KRBTGT パスワードをリセットして対応するが、2 回リセットが必要（レプリケーションの都合）。""",
    """\
## 議論点

- Kerberoasting の対策: サービスアカウントに長いランダムパスワードを設定、グループ管理サービスアカウント（gMSA）を使う
- 時刻同期が崩れると認証が通らなくなる → NTP の重要性
- Kerberos と NTLM の使い分け（Kerberos が優先、失敗時に NTLM へフォールバック）""",
    [TAGS_A["認証"], TAGS_A["セキュリティアーキテクチャ"]], "published")
print(f"    Q-A4: Kerberos → {Q_A4}")

Q_A5 = create_question(KATO_T, TEAM_A,
    "インシデント対応の 6 フェーズと各フェーズでの行動",
    """\
## 問題

NIST SP 800-61 に基づくインシデント対応の 6 フェーズを列挙し、各フェーズで行うべき具体的な行動を説明せよ。また、封じ込め（Containment）フェーズで短期と長期を分ける理由も述べよ。""",
    """\
## 解答

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
**長期**: 証拠を保全しながら恒久的な対策を施す → フォレンジック調査との並行作業が必要""",
    """\
## 解説

「根絶の前に必ず証拠を保全する」点が重要。先にシステムをクリーンアップしてしまうと攻撃者の痕跡が失われる。

封じ込めフェーズでは **ネットワーク分離・ACL 追加・アカウント無効化** などの手段を組み合わせる。

事後対応（Lessons Learned）をスキップする組織が多いが、ここが再発防止の要。""",
    """\
## 議論点・知らなかった知識

- IoC（Indicators of Compromise）の収集と共有（ISACs / STIX / TAXII）
- フォレンジック調査の原則: 揮発性の高いデータ（メモリ）から先に取得
- CSIRT と SOC の役割分担（CSIRT = 意思決定と調整、SOC = 24/7 監視・検知）""",
    [TAGS_A["インシデント対応"]], "published")
print(f"    Q-A5: インシデント対応 → {Q_A5}")

Q_A6 = create_question(TANAKA_T, TEAM_A,
    "Bell-LaPadula モデルと Biba モデルの比較",
    """\
## 問題

以下の 2 つのセキュリティモデルについて、基本ルールと目的を比較せよ。

1. **Bell-LaPadula モデル**
2. **Biba モデル**

また、なぜこれらが互いに相反する設計になっているかを説明せよ。""",
    """\
## 解答

### Bell-LaPadula モデル（機密性重視）

- **No Read Up**: 自分より高い機密レベルのデータを読めない
- **No Write Down**: 自分より低い機密レベルへデータを書けない
- **目的**: 機密情報の漏洩防止（上位から下位への情報流出をブロック）

### Biba モデル（完全性重視）

- **No Write Up**: 自分より高い完全性レベルへ書き込めない
- **No Read Down**: 自分より低い完全性レベルのデータを読めない
- **目的**: データの改ざん防止（低信頼データが高信頼データを汚染しないよう）

### なぜ相反するか

Bell-LaPadula では「上位から下位への書き込みは許可」されているが、Biba では「上位への書き込みは禁止」。両方同時に適用するとほぼ全ての操作が制限される。""",
    """\
## 解説

覚え方：
- Bell-LaPadula = **BLP = 秘密保持（読み上げ禁止・書き下げ禁止）**
- Biba = **完全性（書き上げ禁止・読み下げ禁止）**

軍事システムは BLP を使う（機密漏洩を最優先で防ぐ）。金融システムは Biba を優先することが多い（データの正確性が命）。""",
    """\
## 議論点

- Clark-Wilson モデル: 商用環境向けの完全性モデル。トランザクションの整合性をルールで担保
- Chinese Wall（Brewer-Nash）モデル: 利益相反防止が目的（コンサルタント等が競合企業の情報にアクセスできないよう）
- Graham-Denning モデル: オブジェクトとサブジェクトの生成・削除も含めたアクセス制御""",
    [TAGS_A["セキュリティアーキテクチャ"], TAGS_A["アクセス制御"]], "published")
print(f"    Q-A6: BLP/Biba → {Q_A6}")

Q_A7 = create_question(SATO_T, TEAM_A,
    "物理セキュリティ：多重防護（Defense in Depth）の設計",
    """\
## 問題

データセンターの物理セキュリティを「多重防護（Defense in Depth）」の観点から設計せよ。外周から内部に向けて、各層でどのような対策を講じるべきか。""",
    """\
## 解答

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
- 動作中の筐体への耐タンパー対策""",
    """\
## 解説

「多重防護」の本質は、1 つの対策が破られても次の層が機能すること。

マンタップは「1 人ずつしか通れない二重ドア」でテールゲーティング（ピギーバック）を防ぐ。

物理セキュリティを軽視するとソーシャルエンジニアリング（なりすまし入館）で全ての論理的なセキュリティが無力化される。""",
    """\
## 議論点

- 防犯カメラの映像保存期間（最低 90 日推奨）と証拠としての法的有効性
- 電磁波漏洩対策（TEMPEST / ファラデーケージ）: 軍事・政府施設で適用
- 環境制御（温度・湿度・消火設備）も物理セキュリティの重要要素""",
    [TAGS_A["物理セキュリティ"]], "draft")
print(f"    Q-A7: 物理セキュリティ（下書き）→ {Q_A7}")

# ============================================================
# 8. 問題作成 — チームB (SC試験)
# ============================================================
print("\n【8】問題作成 — チームB (SC試験チーム)...")

Q_B1 = create_question(YAMADA_T, TEAM_B,
    "SQL インジェクションの仕組みと対策",
    """\
## 問題

以下のコードに対する SQL インジェクション攻撃の仕組みを説明し、適切な対策を 3 つ以上挙げよ。

```sql
SELECT * FROM users WHERE username = '入力値' AND password = '入力値';
```

攻撃者が username に `' OR '1'='1` を入力した場合、クエリはどう変化するか。""",
    """\
## 解答

### 攻撃の仕組み

入力値 `' OR '1'='1` を代入すると：

```sql
SELECT * FROM users WHERE username = '' OR '1'='1' AND password = '';
```

さらに `' OR '1'='1' --` とすることで AND 以降がコメントアウトされ全件ヒットする。

### 対策

1. **プリペアドステートメント（バインド変数）** — 最も効果的
2. **ストアドプロシージャ** — パラメータが適切にバインドされる場合のみ有効
3. **入力値のバリデーション** — ホワイトリスト方式で許可文字を限定
4. **エスケープ処理** — シングルクォートをエスケープ（プリペアドの代替だが脆弱性が残りやすい）
5. **WAF（Web Application Firewall）** — 既知の攻撃パターンをブロック（多層防御として）""",
    """\
## 解説

SQL インジェクションは OWASP Top 10 の常連で、SC 試験でも毎年出題される。

プリペアドステートメントが最善策だが、「なぜ安全か」を説明できることが重要：SQL 文の構造（テンプレート）がデータベースに先に解析されるため、後からバインドされる値は SQL 構文として解釈されない。

UNION インジェクション・エラーベース・ブラインドインジェクションなど、手法のバリエーションも把握しておくこと。""",
    """\
## 議論点・知らなかった知識

- 2 次 SQL インジェクション: データを一度 DB に格納し、後で別の処理で取り出す際に注入が発動する
- ORM を使えば自動的に安全というわけではない。生クエリ（raw query）を使う箇所は要注意
- 最小権限: DB ユーザーに SELECT しか権限を与えなければ INSERT/DROP のリスクを下げられる""",
    [TAGS_B["Webセキュリティ"], TAGS_B["脆弱性管理"]], "published")
print(f"    Q-B1: SQLインジェクション → {Q_B1}")

Q_B2 = create_question(NAKAMURA_T, TEAM_B,
    "XSS（クロスサイトスクリプティング）の種類と対策",
    """\
## 問題

XSS には大きく 3 種類ある。それぞれの仕組みと特徴を説明し、開発者として実施すべき対策をまとめよ。

1. 反射型 XSS（Reflected XSS）
2. 蓄積型 XSS（Stored XSS）
3. DOM ベース XSS（DOM-based XSS）""",
    """\
## 解答

### 1. 反射型 XSS
- リクエストに含まれたスクリプトがレスポンスに反射して実行される
- 罠リンクを踏ませる必要がある

### 2. 蓄積型 XSS
- 攻撃スクリプトが DB などに保存され、閲覧したユーザー全員に実行される
- **最も危険**（持続的に被害が発生する）

### 3. DOM ベース XSS
- サーバーは関与せず、ブラウザ上の JS が DOM を操作する際に脆弱性が生じる
- WAF では検知が困難

### 対策

| 対策 | 効果 |
|-----|------|
| 出力エスケープ（HTML エンティティ変換）| 反射型・蓄積型に有効 |
| Content Security Policy（CSP）ヘッダー | インラインスクリプト実行をブロック |
| HttpOnly Cookie | スクリプトから Cookie を読めなくする |
| `innerHTML` の代わりに `textContent` 使用 | DOM ベース XSS に有効 |""",
    """\
## 解説

「エスケープは出力時に行う」が鉄則。入力時に除去しようとするアプローチは二重エスケープや迂回の温床になる。

CSP は XSS の緩和策として強力だが、設定が複雑で実運用で `unsafe-inline` を許可してしまうケースが多い。

DOM ベース XSS はサーバー側のコードレビューだけでは見つからない → フロントエンドのコードレビューが必要。""",
    """\
## 議論点

- Trusted Types API: DOM 操作に型付けされた値のみ許可する仕組み（Chrome で実装済み）
- mXSS（mutation XSS）: サニタイズ後に DOM が変化してスクリプトが復活するケース
- Self-XSS: ユーザー自身が騙されてコンソールにスクリプトを入力させられる攻撃""",
    [TAGS_B["Webセキュリティ"], TAGS_B["脆弱性管理"]], "published")
print(f"    Q-B2: XSS → {Q_B2}")

Q_B3 = create_question(INOUE_T, TEAM_B,
    "TLS ハンドシェイクの流れ（TLS 1.3）",
    """\
## 問題

TLS 1.3 のハンドシェイクフローを説明し、TLS 1.2 と比べた際の主な改善点を 3 つ挙げよ。また **前方秘匿性（PFS: Perfect Forward Secrecy）** とは何か説明せよ。""",
    """\
## 解答

### TLS 1.3 ハンドシェイクフロー（1-RTT）

```
Client                          Server
  |-- ClientHello ─────────────→  |  (対応暗号スイート・鍵共有パラメータ)
  |← ServerHello ─────────────── |  (選択した暗号スイート・鍵共有)
  |← {EncryptedExtensions} ───── |
  |← {Certificate} ────────────── |
  |← {CertificateVerify} ──────── |
  |← {Finished} ───────────────── |
  |-- {Finished} ──────────────→  |
  |== アプリケーションデータ通信 ==|
```

### TLS 1.2 との比較

| 項目 | TLS 1.2 | TLS 1.3 |
|-----|---------|---------|
| ハンドシェイク | 2-RTT | 1-RTT |
| 前方秘匿性 | オプション | **必須**（ECDHE/DHE のみ） |
| 廃止された機能 | RSA 鍵交換, RC4, DES, MD5, SHA-1 | 廃止 |

### 前方秘匿性（PFS）

セッションごとに一時的な鍵を生成し、後から長期秘密鍵が漏洩しても過去のセッションを復号できない性質。""",
    """\
## 解説

TLS 1.3 は 2018 年に RFC 8446 として標準化。1.2 の様々な既知攻撃（POODLE, BEAST, CRIME 等）への対策が設計に組み込まれている。

0-RTT（Early Data）は前のセッション情報を使って即座にデータ送信できるが、リプレイ攻撃の危険性があるため冪等な操作にのみ使用すべき。""",
    """\
## 議論点

- Certificate Transparency（CT）: 証明書の透明性ログで不正な証明書の発行を検知
- OCSP Stapling: 証明書失効確認をリアルタイムに行う仕組み
- 量子コンピュータ時代に備えた PQC（Post-Quantum Cryptography）: NIST が ML-KEM(Kyber) 等を標準化""",
    [TAGS_B["暗号プロトコル"], TAGS_B["ネットワークセキュリティ"]], "published")
print(f"    Q-B3: TLSハンドシェイク → {Q_B3}")

Q_B4 = create_question(SUZUKI_T, TEAM_B,
    "CSRF（クロスサイトリクエストフォージェリ）攻撃と防御",
    """\
## 問題

CSRF 攻撃の仕組みを具体的なシナリオで説明し、効果的な防御策をそれぞれの原理とともに述べよ。また XSS との違いも明確にせよ。""",
    """\
## 解答

### CSRF 攻撃のシナリオ

1. 被害者がオンラインバンクにログイン中（Cookie でセッション維持）
2. 攻撃者が罠サイトに以下を仕込む：
   `<img src="https://bank.example.com/transfer?to=attacker&amount=100000">`
3. 被害者が罠サイトを閲覧 → ブラウザが自動的に Cookie 付きリクエストを送信
4. バンクは正規ユーザーからのリクエストと判断し送金を実行

### 防御策

| 対策 | 原理 |
|-----|------|
| **CSRF トークン** | フォームに予測不能なトークンを埋め込み、サーバーで検証 |
| **SameSite Cookie 属性** | Strict: 完全にクロスサイトリクエストで Cookie を送らない |
| **カスタムリクエストヘッダー** | `X-Requested-With: XMLHttpRequest` をチェック |
| **Origin / Referer ヘッダー検証** | リクエスト元ドメインを確認（補助的対策）|

### XSS との違い

| | CSRF | XSS |
|-|------|-----|
| 攻撃対象 | サーバーに対するリクエスト | ブラウザ上でのスクリプト実行 |
| 認証の悪用 | 被害者の認証情報を利用 | 被害者のブラウザを乗っ取る |""",
    """\
## 解説

SameSite=Strict が最も強力だが、「別ドメインからのリンクでも Cookie が送られない」ため UX への影響が大きい場合がある。

多くのモダンブラウザは `SameSite=Lax` をデフォルトで適用するようになっているが、明示的に設定することを推奨。""",
    """\
## 議論点

- JSON API は CSRF に比較的安全（Content-Type: application/json はプリフライトが発生するため）だが、text/plain で JSON を受け付けるなら危険
- CORS の設定ミス（Access-Control-Allow-Origin: *）は CSRF 対策を無効化しうる
- ログアウト CSRF: 被害者を強制ログアウトさせる攻撃""",
    [TAGS_B["Webセキュリティ"], TAGS_B["不正アクセス"]], "published")
print(f"    Q-B4: CSRF → {Q_B4}")

Q_B5 = create_question(YAMADA_T, TEAM_B,
    "マルウェアの分類と各種対策技術",
    """\
## 問題

以下のマルウェアの種類について、それぞれの特徴・感染経路・代表的な対策を説明せよ。

1. ランサムウェア
2. ルートキット
3. ワーム
4. RAT（Remote Access Trojan）""",
    """\
## 解答

### 1. ランサムウェア
- **特徴**: ファイルを暗号化して身代金を要求
- **感染経路**: フィッシングメール添付ファイル、RDP 脆弱性
- **対策**: 定期バックアップ（3-2-1 ルール）、EDR 導入、ネットワーク分離

### 2. ルートキット
- **特徴**: OS のカーネルレベルに潜伏し、自身や他のマルウェアを隠蔽
- **感染経路**: ブートローダー改ざん、カーネルモジュール挿入
- **対策**: セキュアブート（UEFI）、定期的なハッシュ検証、再インストール

### 3. ワーム
- **特徴**: ホストを必要とせず自己複製して拡散する
- **感染経路**: ネットワーク脆弱性（MS17-010 等）、USB
- **対策**: パッチ管理（迅速適用）、ネットワーク分離

### 4. RAT（Remote Access Trojan）
- **特徴**: 攻撃者にリモートコントロール手段を提供
- **感染経路**: フィッシング、水飲み場攻撃
- **対策**: EDR/XDR による振る舞い検知、通信の監視""",
    """\
## 解説

3-2-1 バックアップルール：3 つのコピー、2 種類のメディア、1 つはオフサイト（ネットワーク切り離し）。ランサムウェアはバックアップも暗号化しようとするため、オフライン・エアギャップバックアップが重要。

ルートキットの検出は極めて困難。感染が確認されたら「クリーンアップ」ではなく「OS 再インストール」が原則。""",
    """\
## 議論点

- APT（Advanced Persistent Threat）: 国家関与のマルウェアは Living-off-the-Land（LOLBins）で正規ツールを悪用するため EDR でも検知困難
- PolyMorphic/Metamorphic Malware: コードが変形するためシグネチャベース検知が効かない
- サンドボックス解析: 動的解析でマルウェアの振る舞いを把握""",
    [TAGS_B["マルウェア"]], "published")
print(f"    Q-B5: マルウェア → {Q_B5}")

Q_B6 = create_question(NAKAMURA_T, TEAM_B,
    "ファイアウォールと IDS/IPS の違いと組み合わせ方",
    """\
## 問題

ファイアウォール、IDS（侵入検知システム）、IPS（侵入防止システム）の違いを説明し、企業ネットワークでどのように配置・組み合わせるべきかを述べよ。""",
    """\
## 解答

### 各機器の役割

| 機器 | 動作 | トラフィックへの影響 |
|-----|------|-----------------|
| ファイアウォール | パケットフィルタリング（IP/Port ベース） | インライン、遮断可能 |
| IDS | トラフィックを監視・アラート通知 | **コピーを監視（受動的）** |
| IPS | IDS + 自動遮断 | **インラインに配置（能動的）** |

### 配置パターン（企業ネットワーク）

```
インターネット → [外部 FW] → [DMZ] → [内部 FW + IPS] → [内部ネットワーク]
                                              ↑
                                    [IDS（SPAN ポートでコピー監視）]
```

IPS はインラインで誤検知（False Positive）があるとサービス影響が出るため、初期は **IDS モード**で運用 → チューニング後に **IPS モード**に切り替えるのが一般的。""",
    """\
## 解説

次世代 FW（NGFW）は L7 までの検査ができ、IPS 機能を内蔵したものが多い。

SIEM と組み合わせると IDS/IPS・FW のログを集約して相関分析ができる。単体のアラートでは気づきにくい攻撃パターンを検出できる。""",
    """\
## 議論点

- UTM vs NGFW: UTM はオールインワン（SMB 向け）、NGFW はエンタープライズ向けで高性能
- ゼロデイ攻撃: シグネチャベースの IDS/IPS では防げない → サンドボックス・振る舞い検知が必要
- 暗号化トラフィック（HTTPS）の検査: SSL インスペクション（中間で復号）はプライバシーとのバランスが必要""",
    [TAGS_B["ネットワークセキュリティ"], TAGS_B["不正アクセス"]], "published")
print(f"    Q-B6: FW/IDS/IPS → {Q_B6}")

Q_B7 = create_question(INOUE_T, TEAM_B,
    "ログ分析：不正アクセスの痕跡を特定せよ",
    """\
## 問題

以下の Web サーバーアクセスログ（抜粋）から、不正アクセスの痕跡を特定し、攻撃者が実施した攻撃の種類と対策を述べよ。

```
203.0.113.5 - - [21/Jun/2026:03:12:01 +0900] "GET /login HTTP/1.1" 200 1234
203.0.113.5 - - [21/Jun/2026:03:12:03 +0900] "POST /login HTTP/1.1" 401 256
(同じIPから03:12〜03:13の1分間にPOST /loginが248回)
203.0.113.5 - - [21/Jun/2026:03:13:58 +0900] "POST /login HTTP/1.1" 200 4521
203.0.113.5 - - [21/Jun/2026:03:14:02 +0900] "GET /admin/users HTTP/1.1" 200 9873
203.0.113.5 - - [21/Jun/2026:03:14:15 +0900] "GET /admin/export?format=csv HTTP/1.1" 200 589234
```""",
    """\
## 解答

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
| ログ監視 | SIEM でリアルタイムアラート設定 |""",
    """\
## 解説

このログパターンは試験でも実務でも頻出。「短時間に大量の 401 → 突然の 200」というパターンは即座にアラートが上がるべき。

ブルートフォース対策のアカウントロックアウトは DoS にも使われる（存在するアカウントを意図的にロックさせる）。CAPTCHA や指数バックオフのほうが現実的なケースも多い。""",
    """\
## 議論点

- ログの完全性保護: 攻撃者がログを改ざんした場合に備えて、リモートの syslog サーバーや WORM ストレージへ転送
- タイムゾーンの統一: ログ相関分析では UTC 統一が必須
- 589KB の CSV: 個人情報 DB ならば GDPR・個人情報保護法のインシデント報告義務が発生する可能性""",
    [TAGS_B["ログ・監視"], TAGS_B["不正アクセス"], TAGS_B["インシデント対応"]], "published")
print(f"    Q-B7: ログ分析 → {Q_B7}")

# ============================================================
# 9. コメント投稿
# ============================================================
print("\n【9】コメント投稿...")

comments = [
    # チームA
    (SATO_T, TEAM_A, Q_A1, "ECB の「ペンギン問題」は視覚的で分かりやすいですね。GCM の nonce 再利用攻撃は実際に 2016 年の TLS 実装で問題になったケースがありました。"),
    (YAMADA_T, TEAM_A, Q_A1, "AES-GCM-SIV はまだ普及途上ですが、LibreSSL などで実装が進んでいます。試験には出なくても知っておく価値がありそうです。"),
    (TANAKA_T, TEAM_A, Q_A1, "TLS 1.3 では AES-GCM が必須になっているので、このあたりの理解は実務でも直結しますね。"),
    (TANAKA_T, TEAM_A, Q_A2, "「DAC は所有者が権限を管理する」という点が混乱しやすいですが、Unix の chmod がまさに DAC の実装ですね。"),
    (KATO_T, TEAM_A, Q_A2, "ABAC をゼロトラストと結びつけて理解できました。属性（ユーザー・リソース・環境・時刻など）で動的に判断するイメージです。"),
    (SATO_T, TEAM_A, Q_A3, "ALE の計算式は本番でも使えるよう手を動かして練習しておきます。SLE = AV × EF をまず頭に入れます。"),
    (SUZUKI_T, TEAM_A, Q_A3, "「ALE(before) - ALE(after) - コスト > 0」の式、経営層への投資判断の説明に使えそうです。"),
    (YAMADA_T, TEAM_A, Q_A4, "Golden Ticket 攻撃はドメインコントローラーが侵害された場合のシナリオですね。対策が KRBTGT パスワードリセット × 2 というのも覚えておきます。"),
    (KATO_T, TEAM_A, Q_A4, "Kerberoasting は攻撃ツール（Rubeus 等）で簡単に実行できるため、現実のペネトレテストでも頻繁に使われる手法です。"),
    (TANAKA_T, TEAM_A, Q_A5, "フォレンジック調査の「揮発性の高いデータから先に取得」というのは実務でも重要ですね。メモリダンプを先に取る理由が理解できました。"),
    (SATO_T, TEAM_A, Q_A6, "Clark-Wilson と Chinese Wall は CISSP のドメイン 3 でよく出てきますね。Clark-Wilson の CDP/UDPの概念もまとめておきたいです。"),
    # チームB
    (NAKAMURA_T, TEAM_B, Q_B1, "ORM でも raw クエリの部分は要注意というのは盲点でした。ActiveRecord の where に文字列補間を使うと危険という話と同じですね。"),
    (INOUE_T, TEAM_B, Q_B1, "2 次 SQL インジェクションは初めて知りました。「一度保存した値が後で SQL として解釈される」というシナリオは気をつけます。"),
    (SUZUKI_T, TEAM_B, Q_B1, "WAF も完全ではないので、アプリ側の対策（プリペアドステートメント）が必須ですね。WAF は多層防御の一つとして。"),
    (INOUE_T, TEAM_B, Q_B2, "DOM ベース XSS は innerHTML 以外にも document.write や eval も危険ですね。フロントエンドのレビューが重要だと改めて思いました。"),
    (YAMADA_T, TEAM_B, Q_B2, "CSP ヘッダーの設定は実際に試してみると難しいですね。report-only モードで段階的に導入するアプローチが実用的です。"),
    (SUZUKI_T, TEAM_B, Q_B3, "0-RTT のリプレイ攻撃リスクは盲点でした。GET でも冪等でない操作には使わない方が安全ですね。"),
    (NAKAMURA_T, TEAM_B, Q_B3, "PQC のあたりは SC 試験にも出始めているようです。CRYSTALS-Kyber（ML-KEM）が NIST 標準になったのを覚えておきます。"),
    (YAMADA_T, TEAM_B, Q_B4, "SameSite=Lax のデフォルト化で CSRF 攻撃は減りましたが、明示的な設定とトークン検証の組み合わせが確実ですね。"),
    (NAKAMURA_T, TEAM_B, Q_B5, "3-2-1 バックアップはランサムウェア対策の基本ですが、オフサイト（クラウド含む）への転送が漏れているケースが多いですね。"),
    (INOUE_T, TEAM_B, Q_B5, "LOLBins（Living-off-the-Land）は正規ツール（PowerShell, WMI 等）を悪用するため検知が難しい。振る舞い検知が重要です。"),
    (SUZUKI_T, TEAM_B, Q_B7, "このログ問題、実際の試験形式に近くて良い練習になりました。タイムスタンプの読み方や 401/200 の意味を素早く判断できるようにしたいです。"),
    (YAMADA_T, TEAM_B, Q_B7, "589KB の CSV ダウンロードは確かに GDPR 違反になりそうです。インシデント対応後の報告義務（72 時間以内）も押さえておきます。"),
]

for token, team_id, q_id, body in comments:
    post_comment(token, team_id, q_id, body)

print(f"    {len(comments)} 件のコメントを投稿しました")

# ============================================================
# 完了
# ============================================================
print("""
==========================================
  テストデータ投入完了！
==========================================

ユーザー一覧:
  admin        / Admin1234!    (admin)
  tanaka       / Password1!    (CISSPチーム オーナー・is_team_owner)
  sato         / Password1!    (CISSPチームメンバー)
  yamada       / Password1!    (SC試験チームオーナー・CISSPチームメンバー・is_team_owner)
  suzuki       / Password1!    (両チームメンバー・is_team_owner)
  nakamura     / Password1!    (SC試験チームメンバー)
  inoue        / Password1!    (SC試験チームメンバー)
  kato         / Password1!    (CISSPチームメンバー)

チーム:
  CISSPチーム  … 問題 7 件（公開 6 / 下書き 1）、タグ 8 種
  SC試験チーム … 問題 7 件（公開 7）、タグ 8 種

アクセス: http://localhost:3000
""")
