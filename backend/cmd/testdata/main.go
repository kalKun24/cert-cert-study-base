// テストデータ投入スクリプト。
// FIRESTORE_EMULATOR_HOST と GCP_PROJECT_ID を環境変数で渡して実行する。
//
// 投入内容:
//   - ユーザー: admin 1名 + teamowner 2名 + 一般ユーザー 6名 (計9名)
//   - チーム: 3チーム（CISSP勉強会 / 情報処理安全確保支援士勉強会 / 共通勉強会）
//   - タグ: 各チームに 5〜8 件
//   - 問題: 各チームに 10 件（ステータス混在）
//   - 問題コメント: 各問題に 2〜3 件
//   - ノート: 各チームに 5 件
//   - ノートコメント: 各ノートに 2 件
//
// 注意: 以下に定義する各 *Record 型は
// backend/internal/infrastructure/firestore/ 配下の同名 struct と
// フィールドタグを一致させる必要があります。
// どちらかを変更した場合は必ずもう一方も更新してください。
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "local-project"
	}
	emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if emulatorHost == "" {
		log.Fatal("FIRESTORE_EMULATOR_HOST が設定されていません。エミュレータ起動後に実行してください。")
	}

	ctx := context.Background()
	client, err := fs.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	defer client.Close()

	log.Printf("Firestoreエミュレータ (%s) にテストデータを投入します...", emulatorHost)

	s := &seeder{client: client, ctx: ctx}
	if err := s.run(); err != nil {
		log.Fatalf("テストデータ投入に失敗しました: %v", err)
	}
	log.Println("完了しました。")
}

type seeder struct {
	client *fs.Client
	ctx    context.Context
}

func (s *seeder) run() error {
	// ユーザー作成
	users, err := s.seedUsers()
	if err != nil {
		return err
	}

	// チーム作成
	teams, err := s.seedTeams(users)
	if err != nil {
		return err
	}

	// タグ作成
	tagsByTeam, err := s.seedTags(teams)
	if err != nil {
		return err
	}

	// 問題 + コメント作成
	if err := s.seedQuestions(users, teams, tagsByTeam); err != nil {
		return err
	}

	// ノート + コメント作成
	if err := s.seedNotes(users, teams, tagsByTeam); err != nil {
		return err
	}

	return nil
}

// ---------- ユーザー ----------

// userRecord は infrastructure/firestore/user_repository.go の userRecord と同一定義です。
type userRecord struct {
	ID           string     `firestore:"id"`
	Username     string     `firestore:"username"`
	DisplayName  string     `firestore:"display_name"`
	Email        string     `firestore:"email"`
	PasswordHash string     `firestore:"password_hash"`
	Role         string     `firestore:"role"`
	IsActive     bool       `firestore:"is_active"`
	IsTeamOwner  bool       `firestore:"is_team_owner"`
	MaxTeams     int        `firestore:"max_teams"`
	CreatedAt    time.Time  `firestore:"created_at"`
	UpdatedAt    time.Time  `firestore:"updated_at"`
	LastLoginAt  *time.Time `firestore:"last_login_at"`
}

type userInfo struct {
	id          string
	username    string
	displayName string
	role        string
	isTeamOwner bool
}

func hashPassword(pw string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (s *seeder) seedUsers() ([]userInfo, error) {
	now := time.Now().UTC()

	defs := []struct {
		username    string
		displayName string
		email       string
		role        string
		isTeamOwner bool
	}{
		{"admin", "管理者", "admin@example.com", "admin", false},
		{"owner_alice", "Alice（チームオーナー）", "alice@example.com", "user", true},
		{"owner_bob", "Bob（チームオーナー）", "bob@example.com", "user", true},
		{"user_carol", "Carol", "carol@example.com", "user", false},
		{"user_dave", "Dave", "dave@example.com", "user", false},
		{"user_eve", "Eve", "eve@example.com", "user", false},
		{"user_frank", "Frank", "frank@example.com", "user", false},
		{"user_grace", "Grace", "grace@example.com", "user", false},
		{"user_heidi", "Heidi", "heidi@example.com", "user", false},
	}

	hash := hashPassword("password123")
	var infos []userInfo

	for _, d := range defs {
		id := uuid.NewString()
		rec := userRecord{
			ID:           id,
			Username:     d.username,
			DisplayName:  d.displayName,
			Email:        d.email,
			PasswordHash: hash,
			Role:         d.role,
			IsActive:     true,
			IsTeamOwner:  d.isTeamOwner,
			MaxTeams:     3,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if _, err := s.client.Collection("users").Doc(id).Set(s.ctx, rec); err != nil {
			return nil, fmt.Errorf("ユーザー保存失敗 (%s): %w", d.username, err)
		}
		infos = append(infos, userInfo{id: id, username: d.username, displayName: d.displayName, role: d.role, isTeamOwner: d.isTeamOwner})
		log.Printf("  ユーザー作成: %s (%s)", d.displayName, d.username)
	}
	return infos, nil
}

// ---------- チーム ----------

// teamRecord は infrastructure/firestore/team_repository.go の teamRecord と同一定義です。
type teamRecord struct {
	ID          string    `firestore:"id"`
	Name        string    `firestore:"name"`
	Description string    `firestore:"description"`
	OwnerID     string    `firestore:"owner_id"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

// memberRecord は infrastructure/firestore/team_repository.go の teamMemberRecord と同一定義です。
type memberRecord struct {
	TeamID   string    `firestore:"team_id"`
	UserID   string    `firestore:"user_id"`
	Role     string    `firestore:"role"`
	JoinedAt time.Time `firestore:"joined_at"`
}

type teamInfo struct {
	id      string
	name    string
	members []string // userID list
}

func (s *seeder) seedTeams(users []userInfo) ([]teamInfo, error) {
	now := time.Now().UTC()

	// users index helpers
	byUsername := func(name string) userInfo {
		for _, u := range users {
			if u.username == name {
				return u
			}
		}
		panic("unknown user: " + name)
	}

	type teamDef struct {
		name        string
		description string
		ownerName   string
		memberNames []string
	}

	defs := []teamDef{
		{
			name:        "CISSP勉強会",
			description: "CISSP取得を目指す勉強会チームです。セキュリティドメインを中心に問題を作成・共有します。",
			ownerName:   "owner_alice",
			memberNames: []string{"user_carol", "user_dave", "user_eve"},
		},
		{
			name:        "情報処理安全確保支援士勉強会",
			description: "情報処理安全確保支援士試験（SC試験）の合格を目指す勉強会です。",
			ownerName:   "owner_bob",
			memberNames: []string{"user_frank", "user_grace", "user_heidi"},
		},
		{
			name:        "共通セキュリティ勉強会",
			description: "資格を問わず、セキュリティ全般について学ぶ共通勉強会です。",
			ownerName:   "owner_alice",
			memberNames: []string{"user_carol", "user_frank", "user_grace"},
		},
	}

	var infos []teamInfo

	for _, d := range defs {
		id := uuid.NewString()
		owner := byUsername(d.ownerName)

		rec := teamRecord{
			ID:          id,
			Name:        d.name,
			Description: d.description,
			OwnerID:     owner.id,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if _, err := s.client.Collection("teams").Doc(id).Set(s.ctx, rec); err != nil {
			return nil, fmt.Errorf("チーム保存失敗 (%s): %w", d.name, err)
		}

		// オーナーをメンバーとして追加
		allMemberIDs := []string{owner.id}
		ownerMem := memberRecord{TeamID: id, UserID: owner.id, Role: "owner", JoinedAt: now}
		if _, err := s.client.Collection("teams").Doc(id).Collection("members").Doc(owner.id).Set(s.ctx, ownerMem); err != nil {
			return nil, fmt.Errorf("オーナーメンバー追加失敗: %w", err)
		}

		// 一般メンバーを追加
		for _, mn := range d.memberNames {
			u := byUsername(mn)
			mem := memberRecord{TeamID: id, UserID: u.id, Role: "member", JoinedAt: now}
			if _, err := s.client.Collection("teams").Doc(id).Collection("members").Doc(u.id).Set(s.ctx, mem); err != nil {
				return nil, fmt.Errorf("メンバー追加失敗 (%s): %w", mn, err)
			}
			allMemberIDs = append(allMemberIDs, u.id)
		}

		infos = append(infos, teamInfo{id: id, name: d.name, members: allMemberIDs})
		log.Printf("  チーム作成: %s (%d名)", d.name, len(allMemberIDs))
	}
	return infos, nil
}

// ---------- タグ ----------

// tagRecord は infrastructure/firestore/tag_repository.go の tagRecord と同一定義です。
type tagRecord struct {
	ID        string    `firestore:"id"`
	TeamID    string    `firestore:"team_id"`
	Name      string    `firestore:"name"`
	CreatedAt time.Time `firestore:"created_at"`
}

func (s *seeder) seedTags(teams []teamInfo) (map[string][]string, error) {
	now := time.Now().UTC()

	tagNamesByTeam := map[string][]string{
		teams[0].name: {"アクセス制御", "暗号化", "ネットワークセキュリティ", "リスク管理", "法令・規制", "インシデント対応", "物理セキュリティ", "セキュリティ設計"},
		teams[1].name: {"脅威と脆弱性", "セキュアコーディング", "認証・認可", "ログ管理", "マルウェア対策", "クラウドセキュリティ"},
		teams[2].name: {"OWASP Top10", "ペネトレーションテスト", "フォレンジック", "セキュリティポリシー", "ゼロトラスト"},
	}

	// teamID → []tagID のマップを返す
	result := map[string][]string{}

	for _, team := range teams {
		names := tagNamesByTeam[team.name]
		for _, name := range names {
			id := uuid.NewString()
			rec := tagRecord{ID: id, TeamID: team.id, Name: name, CreatedAt: now}
			docPath := fmt.Sprintf("teams/%s/tags/%s", team.id, id)
			if _, err := s.client.Doc(docPath).Set(s.ctx, rec); err != nil {
				return nil, fmt.Errorf("タグ保存失敗 (%s / %s): %w", team.name, name, err)
			}
			result[team.id] = append(result[team.id], id)
		}
		log.Printf("  タグ作成: %s に %d件", team.name, len(names))
	}
	return result, nil
}

// ---------- 問題 ----------

// questionRecord は infrastructure/firestore/question_repository.go の questionRecord と同一定義です。
type questionRecord struct {
	ID          string    `firestore:"id"`
	TeamID      string    `firestore:"team_id"`
	Title       string    `firestore:"title"`
	Body        string    `firestore:"body"`
	Answer      string    `firestore:"answer"`
	Explanation string    `firestore:"explanation"`
	Memo        string    `firestore:"memo"`
	Tags        []string  `firestore:"tags"`
	Status      string    `firestore:"status"`
	CreatedBy   string    `firestore:"created_by"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

// commentRecord は infrastructure/firestore/comment_repository.go の commentRecord と同一定義です。
type commentRecord struct {
	ID         string    `firestore:"id"`
	QuestionID string    `firestore:"question_id"`
	Body       string    `firestore:"body"`
	CreatedBy  string    `firestore:"created_by"`
	CreatedAt  time.Time `firestore:"created_at"`
	UpdatedAt  time.Time `firestore:"updated_at"`
}

var questionTemplates = []struct {
	title       string
	body        string
	answer      string
	explanation string
	memo        string
}{
	{
		"機密性・完全性・可用性（CIA）トライアドとは何か",
		"情報セキュリティの3大要素について説明しなさい。それぞれの要素が侵害された場合の具体例も挙げること。",
		"**機密性（Confidentiality）**: 権限を持つ者のみが情報にアクセスできること。\n**完全性（Integrity）**: 情報が正確かつ完全であること。\n**可用性（Availability）**: 権限を持つ者が必要なときに情報やシステムにアクセスできること。",
		"CIA トライアドは情報セキュリティの根幹をなす概念です。\n\n- **機密性侵害の例**: 顧客データベースへの不正アクセスによる個人情報漏洩\n- **完全性侵害の例**: SQLインジェクションによるデータ改ざん\n- **可用性侵害の例**: DDoS攻撃によるWebサービスの停止",
		"CISSP試験では CIA の各要素に対する攻撃と対策を問う問題が頻出。AAA (認証・認可・アカウンタビリティ) と組み合わせて覚える。",
	},
	{
		"対称鍵暗号と非対称鍵暗号の違いを説明しなさい",
		"対称鍵暗号（秘密鍵暗号）と非対称鍵暗号（公開鍵暗号）の仕組み・特徴・用途の違いを比較して説明しなさい。",
		"**対称鍵暗号**: 暗号化と復号に同一の鍵を使用。高速だが鍵配送問題がある。例: AES, DES, 3DES\n\n**非対称鍵暗号**: 公開鍵と秘密鍵のペアを使用。鍵配送問題を解決するが遅い。例: RSA, ECC, DSA",
		"実際のシステムでは両者を組み合わせたハイブリッド暗号が一般的。TLSでは、非対称鍵暗号で対称鍵を安全に交換し、データ転送には対称鍵暗号を使用する。",
		"鍵長と安全性の関係も重要。RSA 2048bit ≈ AES 112bit の安全性レベル。ECCは同等安全性をより短い鍵長で実現できる。",
	},
	{
		"アクセス制御モデル（MAC / DAC / RBAC）の特徴を比較しなさい",
		"強制アクセス制御（MAC）、任意アクセス制御（DAC）、ロールベースアクセス制御（RBAC）それぞれの特徴と適用場面を説明しなさい。",
		"**MAC**: セキュリティラベルに基づいてOSが強制的に制御。軍や政府系で使用。\n**DAC**: 資源の所有者がアクセス権を設定。一般的なOSのファイルシステム。\n**RBAC**: ロール（役割）に権限を付与し、ユーザーにロールを割り当て。企業での利用が多い。",
		"MAC は Bell-LaPadula モデル（機密性重視）と Biba モデル（完全性重視）が代表的。DAC はトロイの木馬攻撃に弱い弱点がある。RBACは最小権限の原則を実装しやすい。",
		"ABAC（属性ベースアクセス制御）も近年注目されている。ゼロトラストアーキテクチャとの親和性が高い。",
	},
	{
		"PKIにおける証明書失効の仕組みを説明しなさい",
		"公開鍵基盤（PKI）における証明書の失効確認方法について、CRL と OCSP の違いを含めて説明しなさい。",
		"**CRL（証明書失効リスト）**: 失効した証明書のシリアル番号一覧。CAが定期的に発行するが、リストが大きくなりやすくリアルタイム性に欠ける。\n**OCSP（Online Certificate Status Protocol）**: リアルタイムで個別の証明書失効状態を問い合わせるプロトコル。CRLより軽量でリアルタイム性が高い。",
		"OCSP Stapling を使うと、サーバーが事前にOCSPレスポンスを取得してTLSハンドシェイク時に提示できるため、クライアントが個別にOCSPサーバーへ問い合わせる必要がなくなり、プライバシーとパフォーマンスが向上する。",
		"CAの信頼チェーン（Root CA → Intermediate CA → End Entity Certificate）の理解も重要。ルートCAの秘密鍵保護が全体の安全性を支える。",
	},
	{
		"ソーシャルエンジニアリング攻撃の種類と対策",
		"ソーシャルエンジニアリング攻撃の主な種類（フィッシング・スピアフィッシング・ビッシング・スミッシング・プリテキスティングなど）と、組織としての対策を説明しなさい。",
		"**フィッシング**: 偽のメールでユーザーを騙し認証情報を盗む攻撃\n**スピアフィッシング**: 特定個人を狙った標的型フィッシング\n**ビッシング**: 電話を使ったソーシャルエンジニアリング\n**スミッシング**: SMSを使ったフィッシング\n**プリテキスティング**: 偽の身元・状況を設定して情報を引き出す攻撃",
		"技術的対策だけでは防ぎにくいため、セキュリティ教育・訓練が重要。定期的なフィッシングシミュレーションを実施し、従業員の意識向上を図る。多要素認証の導入も有効。",
		"権限の低いユーザーが標的になりやすい。最小権限の原則と組み合わせてリスクを最小化する。",
	},
	{
		"ファイアウォールの種類と動作原理を説明しなさい",
		"パケットフィルタリング型・ステートフルインスペクション型・アプリケーション層（プロキシ）型ファイアウォールの動作原理と特徴を比較しなさい。",
		"**パケットフィルタリング**: IPアドレス・ポート番号・プロトコルでフィルタリング。高速だが状態管理なし。\n**ステートフルインスペクション**: セッションの状態を追跡し、不正な状態のパケットをブロック。\n**アプリケーション層（プロキシ）**: アプリケーションプロトコルを理解してフィルタリング。最も高機能だが処理負荷が高い。",
		"次世代ファイアウォール（NGFW）はステートフル + アプリケーション識別 + IPS/IDSを統合。UTM（統合脅威管理）とも呼ばれる。",
		"DMZ（非武装地帯）の設計でファイアウォールを2枚使いすることでセキュリティを高める手法も重要。",
	},
	{
		"脆弱性診断とペネトレーションテストの違いを説明しなさい",
		"脆弱性診断（VA）とペネトレーションテスト（ペンテスト）の目的・手法・成果物の違いを説明し、それぞれどのような場面で活用するか述べなさい。",
		"**脆弱性診断**: システムやアプリケーションに存在する脆弱性を網羅的に洗い出す。自動スキャンツールを中心に使用。脆弱性一覧とリスク評価が成果物。\n**ペネトレーションテスト**: 実際の攻撃者の視点でシステムへの侵入を試みる。手動による技術的な攻撃シミュレーション。侵入経路と影響範囲が成果物。",
		"ペンテストにはブラックボックス・グレーボックス・ホワイトボックステストがある。スコープ・ルール・免責事項を事前に明確化することが重要（適切な認可なしに実施すると不法行為となる）。",
		"CVSS（共通脆弱性評価システム）でリスクスコアを定量化することで、対策の優先順位付けが可能。",
	},
	{
		"インシデント対応の手順（IRプロセス）を説明しなさい",
		"情報セキュリティインシデント対応（IR）の一般的な手順・フェーズと、各フェーズで実施すべき活動を説明しなさい。",
		"NIST SP 800-61 に基づく一般的なIRプロセス:\n1. **準備（Preparation）**: インシデント対応計画・体制の整備\n2. **検知・分析（Detection & Analysis）**: インシデントの特定と影響範囲の把握\n3. **封じ込め（Containment）**: 被害の拡大防止\n4. **根絶（Eradication）**: 攻撃の痕跡と原因の除去\n5. **復旧（Recovery）**: サービスの安全な再開\n6. **事後活動（Post-Incident Activity）**: 教訓の文書化と再発防止",
		"インシデント対応はフォレンジックの観点も重要。証拠保全（デジタルフォレンジック）を適切に行うことで、法的手続きに対応できる。フォレンジックの観点で最初に行うべきは証拠のコピー（イメージング）。",
		"BCP（事業継続計画）と IRプランの連携も重要。インシデント時のエスカレーションフローを事前に決めておく。",
	},
	{
		"クラウドセキュリティの責任共有モデルを説明しなさい",
		"IaaS・PaaS・SaaSそれぞれのサービスモデルにおける、クラウドプロバイダーとユーザー（テナント）間のセキュリティ責任の分担について説明しなさい。",
		"**IaaS**: インフラ（ハードウェア・仮想化・ネットワーク）はプロバイダーの責任。OS・ミドルウェア・データ・アプリケーションはユーザーの責任。\n**PaaS**: OS・ランタイムまでプロバイダーの責任。アプリケーション・データはユーザーの責任。\n**SaaS**: ほぼすべてプロバイダーの責任。データ管理とアクセス権設定はユーザーの責任。",
		"責任共有モデルの誤解がクラウドセキュリティインシデントの主要原因の一つ。特にSaaSでも「データの保護責任はユーザーにある」という点が見落とされやすい。CSPMツールを使った継続的なクラウド設定のセキュリティ確認が重要。",
		"AWS・Azure・GCPそれぞれで責任共有モデルのドキュメントが公開されている。CCSP（クラウドセキュリティ専門資格）の観点でも重要概念。",
	},
	{
		"ゼロトラストアーキテクチャの概念と実装アプローチ",
		"ゼロトラストアーキテクチャ（ZTA）の基本原則と、従来の境界防御モデルとの違い、実装における主要コンポーネントを説明しなさい。",
		"**基本原則**: 「Never Trust, Always Verify（決して信頼せず、常に検証する）」\n**従来モデルとの違い**: 境界防御はネットワーク境界を信頼の根拠とするが、ZTAはネットワーク内外を問わず全ての通信を継続的に検証する。\n**主要コンポーネント**: ID管理・マイクロセグメンテーション・最小権限アクセス・継続的な監視・多要素認証",
		"NIST SP 800-207がゼロトラストアーキテクチャの標準的なガイドライン。SDP（Software Defined Perimeter）やマイクロセグメンテーションが実装技術として注目されている。",
		"リモートワーク普及でZTAの重要性が急増。VPNに依存した従来モデルの限界から、SASE（Secure Access Service Edge）との組み合わせが増えている。",
	},
}

var statuses = []string{"draft", "private", "published", "published", "published"} // publishedを多めに

func (s *seeder) seedQuestions(users []userInfo, teams []teamInfo, tagsByTeam map[string][]string) error {
	now := time.Now().UTC()

	// チームメンバーIDを取得するヘルパー
	memberIDs := func(team teamInfo) []string { return team.members }

	for ti, team := range teams {
		tagIDs := tagsByTeam[team.id]
		members := memberIDs(team)

		for i, tmpl := range questionTemplates {
			id := uuid.NewString()
			// 担当タグを1〜2件付ける（ローテーション）
			var qTags []string
			if len(tagIDs) > 0 {
				qTags = append(qTags, tagIDs[i%len(tagIDs)])
				if len(tagIDs) > 1 {
					qTags = append(qTags, tagIDs[(i+1)%len(tagIDs)])
				}
			}

			rec := questionRecord{
				ID:          id,
				TeamID:      team.id,
				Title:       fmt.Sprintf("[%s] %s", string([]rune(team.name)[:4]), tmpl.title),
				Body:        tmpl.body,
				Answer:      tmpl.answer,
				Explanation: tmpl.explanation,
				Memo:        tmpl.memo,
				Tags:        qTags,
				Status:      statuses[i%len(statuses)],
				CreatedBy:   members[i%len(members)],
				CreatedAt:   now.Add(-time.Duration(ti*10+i) * time.Hour),
				UpdatedAt:   now.Add(-time.Duration(ti*10+i) * time.Hour),
			}

			docPath := fmt.Sprintf("teams/%s/questions/%s", team.id, id)
			if _, err := s.client.Doc(docPath).Set(s.ctx, rec); err != nil {
				return fmt.Errorf("問題保存失敗: %w", err)
			}

			// コメントを2件追加
			for c := 0; c < 2; c++ {
				cid := uuid.NewString()
				crec := commentRecord{
					ID:         cid,
					QuestionID: id,
					CreatedBy:  members[(i+c+1)%len(members)],
					Body:       fmt.Sprintf("コメント%d: この問題について補足します。%s の観点からも重要なポイントです。", c+1, tmpl.title),
					CreatedAt:  now.Add(-time.Duration(ti*10+i)*time.Hour + time.Duration(c)*time.Minute),
					UpdatedAt:  now.Add(-time.Duration(ti*10+i)*time.Hour + time.Duration(c)*time.Minute),
				}
				cPath := fmt.Sprintf("teams/%s/questions/%s/comments/%s", team.id, id, cid)
				if _, err := s.client.Doc(cPath).Set(s.ctx, crec); err != nil {
					return fmt.Errorf("コメント保存失敗: %w", err)
				}
			}
		}
		log.Printf("  問題作成: %s に %d件（コメント %d件）", team.name, len(questionTemplates), len(questionTemplates)*2)
	}
	return nil
}

// ---------- ノート ----------

// noteRecord は infrastructure/firestore/note_repository.go の noteRecord と同一定義です。
type noteRecord struct {
	ID               string    `firestore:"id"`
	TeamID           string    `firestore:"team_id"`
	Title            string    `firestore:"title"`
	Body             string    `firestore:"body"`
	DiscussionPoints string    `firestore:"discussion_points"`
	Memo             string    `firestore:"memo"`
	Tags             []string  `firestore:"tags"`
	Status           string    `firestore:"status"`
	CreatedBy        string    `firestore:"created_by"`
	CreatedAt        time.Time `firestore:"created_at"`
	UpdatedAt        time.Time `firestore:"updated_at"`
}

// noteCommentRecord は infrastructure/firestore/note_comment_repository.go の noteCommentRecord と同一定義です。
type noteCommentRecord struct {
	ID        string    `firestore:"id"`
	NoteID    string    `firestore:"note_id"`
	Body      string    `firestore:"body"`
	CreatedBy string    `firestore:"created_by"`
	CreatedAt time.Time `firestore:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at"`
}

var noteTemplates = []struct {
	title            string
	body             string
	discussionPoints string
	memo             string
}{
	{
		"セキュリティフレームワーク比較メモ（NIST CSF / ISO 27001 / CIS Controls）",
		"## NIST Cybersecurity Framework (CSF)\n5つのコア機能: **識別 → 保護 → 検知 → 対応 → 復旧**\n\n## ISO/IEC 27001\nISMS（情報セキュリティ管理システム）の国際標準。PDCAサイクルに基づくリスク管理。\n\n## CIS Controls\n18の重要セキュリティコントロール。優先順位付きで実装可能。",
		"- NIST CSF と ISO 27001 はどのような組織に向いているか？\n- CIS Controls の Implementation Group (IG) 1〜3 の使い分けは？\n- 日本企業が ISO 27001 認証を取るメリット・デメリットは？",
		"勉強会で各フレームワークをマッピングする演習をやると理解が深まりそう。",
	},
	{
		"TLS 1.3 の変更点まとめ",
		"## TLS 1.3 の主な変更点（対 TLS 1.2）\n\n### ハンドシェイクの簡略化\n- 1-RTT ハンドシェイク（TLS 1.2は2-RTT）\n- 0-RTT リザンプション（再接続時の高速化）\n\n### 廃止された脆弱なアルゴリズム\n- RSA 鍵交換（静的鍵交換）→ 前方秘匿性のない方式を全廃\n- RC4, DES, 3DES, MD5, SHA-1 を廃止\n\n### 使用可能な暗号スイート（5種のみ）\n- TLS_AES_128_GCM_SHA256\n- TLS_AES_256_GCM_SHA384\n- TLS_CHACHA20_POLY1305_SHA256 など",
		"- 0-RTT のリプレイ攻撃リスクと対策は？\n- TLS 1.2 からの移行スケジュールをどう計画するか？",
		"RFC 8446 を一度読んでおきたい。",
	},
	{
		"OAuth 2.0 と OpenID Connect の関係整理",
		"## OAuth 2.0\n**認可**のためのプロトコル。アクセストークンを発行してリソースへのアクセスを委譲する。\n\n## OpenID Connect (OIDC)\nOAuth 2.0 の上に**認証**レイヤーを追加したプロトコル。IDトークン（JWT）を発行してユーザーのアイデンティティを確認する。\n\n| | OAuth 2.0 | OIDC |\n|---|---|---|\n| 目的 | 認可 | 認証 + 認可 |\n| トークン | アクセストークン | IDトークン + アクセストークン |",
		"- OAuth 2.0 の Authorization Code Flow と PKCE の必要性は？\n- JWTのペイロードに機密情報を入れてはいけない理由は？\n- Implicit Flow が非推奨になった背景は？",
		"OAuth 2.0 を「認証」プロトコルと誤解しているケースが多い。CISSPでも混同しやすいポイント。",
	},
	{
		"OWASP Top 10 2021 変更点と対策まとめ",
		"## OWASP Top 10 2021\n\n1. **A01: アクセス制御の不備** (2017年5位から1位に浮上)\n2. **A02: 暗号化の失敗** (旧: 機密データの露出)\n3. **A03: インジェクション** (SQLi, XSS, コマンドインジェクション)\n4. **A04: 安全でない設計** (**新規** - 設計段階のセキュリティ)\n5. **A05: セキュリティの設定ミス**\n6. **A06: 脆弱で古いコンポーネント**\n7. **A07: 識別と認証の失敗** (旧: 認証の不備)\n8. **A08: ソフトウェアとデータの整合性の不具合** (**新規**)\n9. **A09: セキュリティログとモニタリングの失敗**\n10. **A10: サーバーサイドリクエストフォージェリ (SSRF)** (**新規**)",
		"- A04「安全でない設計」が新規追加された背景は？\n- A08「ソフトウェアとデータの整合性の不具合」に含まれる CI/CD パイプラインへの攻撃（サプライチェーン攻撃）の具体例は？",
		"2021年版では設計・アーキテクチャ段階のセキュリティがより重視されている。",
	},
	{
		"パスワードレス認証技術の動向（FIDO2 / WebAuthn / パスキー）",
		"## FIDO2 アーキテクチャ\n- **WebAuthn**: ブラウザ-サーバー間のAPI仕様\n- **CTAP**: 認証器-クライアント間のプロトコル\n\n## 認証の流れ\n1. サーバーがチャレンジを送信\n2. 認証器がチャレンジに秘密鍵で署名\n3. サーバーが公開鍵で署名を検証\n\n## パスキー（Passkeys）\nFIDO2 クレデンシャルのデバイス間同期を実現。AppleのiCloud Keychain・GoogleのPassword Managerが対応。フィッシング耐性が高い。",
		"- パスキーのセキュリティモデルにおける信頼の起点は？\n- 生体認証データはデバイス内にのみ保存され、サーバーには送られないが、デバイス紛失時のリカバリーは？",
		"パスワードレス認証は2024年以降急速に普及。SaaS製品への実装が増えているので、ユースケースと制限事項を理解しておきたい。",
	},
}

func (s *seeder) seedNotes(users []userInfo, teams []teamInfo, tagsByTeam map[string][]string) error {
	now := time.Now().UTC()

	for ti, team := range teams {
		tagIDs := tagsByTeam[team.id]
		members := team.members

		for i, tmpl := range noteTemplates {
			id := uuid.NewString()
			var nTags []string
			if len(tagIDs) > 0 {
				nTags = append(nTags, tagIDs[i%len(tagIDs)])
			}

			rec := noteRecord{
				ID:               id,
				TeamID:           team.id,
				Title:            tmpl.title,
				Body:             tmpl.body,
				DiscussionPoints: tmpl.discussionPoints,
				Memo:             tmpl.memo,
				Tags:             nTags,
				Status:           statuses[i%len(statuses)],
				CreatedBy:        members[i%len(members)],
				CreatedAt:        now.Add(-time.Duration(ti*5+i) * time.Hour),
				UpdatedAt:        now.Add(-time.Duration(ti*5+i) * time.Hour),
			}

			docPath := fmt.Sprintf("teams/%s/notes/%s", team.id, id)
			if _, err := s.client.Doc(docPath).Set(s.ctx, rec); err != nil {
				return fmt.Errorf("ノート保存失敗: %w", err)
			}

			// ノートコメントを2件追加
			for c := 0; c < 2; c++ {
				cid := uuid.NewString()
				crec := noteCommentRecord{
					ID:        cid,
					NoteID:    id,
					CreatedBy: members[(i+c+1)%len(members)],
					Body:      fmt.Sprintf("ノートコメント%d: 「%s」について追記します。実際の試験でも重要なトピックです。", c+1, tmpl.title),
					CreatedAt: now.Add(-time.Duration(ti*5+i)*time.Hour + time.Duration(c+10)*time.Minute),
					UpdatedAt: now.Add(-time.Duration(ti*5+i)*time.Hour + time.Duration(c+10)*time.Minute),
				}
				cPath := fmt.Sprintf("teams/%s/notes/%s/comments/%s", team.id, id, cid)
				if _, err := s.client.Doc(cPath).Set(s.ctx, crec); err != nil {
					return fmt.Errorf("ノートコメント保存失敗: %w", err)
				}
			}
		}
		log.Printf("  ノート作成: %s に %d件（コメント %d件）", team.name, len(noteTemplates), len(noteTemplates)*2)
	}
	return nil
}
