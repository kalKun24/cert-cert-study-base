package firestore_test

import (
	"context"
	"os"
	"testing"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	firestoreRepo "github.com/kalKun24/cert-study-base/backend/internal/infrastructure/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// setupCascadeTest は Firestore エミュレータに接続してテスト用リポジトリ群を返します。
func setupCascadeTest(t *testing.T) (*fs.Client, *firestoreRepo.FirestoreQuestionRepository, *firestoreRepo.FirestoreNoteRepository, *firestoreRepo.FirestoreTeamRepository) {
	t.Helper()

	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST が設定されていないためスキップします")
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "test-project"
	}

	ctx := context.Background()
	client, err := fs.NewClient(ctx, projectID)
	if err != nil {
		t.Fatalf("Firestoreクライアントの初期化に失敗しました: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	questionRepo := firestoreRepo.NewFirestoreQuestionRepository(client)
	noteRepo := firestoreRepo.NewFirestoreNoteRepository(client)
	teamRepo := firestoreRepo.NewFirestoreTeamRepository(client)
	return client, questionRepo, noteRepo, teamRepo
}

// countDocs は指定コレクションのドキュメント数を返します。
func countDocs(ctx context.Context, t *testing.T, col *fs.CollectionRef) int {
	t.Helper()
	iter := col.Documents(ctx)
	defer iter.Stop()
	n := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			t.Fatalf("ドキュメント数の取得に失敗しました: %v", err)
		}
		n++
	}
	return n
}

// TestQuestionDelete_CascadesComments は問題削除時にコメントが孤立しないことを確認します。
func TestQuestionDelete_CascadesComments(t *testing.T) {
	ctx := context.Background()
	client, questionRepo, _, _ := setupCascadeTest(t)

	teamID := "test-team-qd-" + time.Now().Format("150405")
	questionID := "test-question-001"

	// 問題を作成
	q := &domain.Question{
		ID: questionID, TeamID: teamID, Title: "テスト問題",
		Body: "本文", Answer: "回答", CreatedBy: "user-1",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := questionRepo.Save(ctx, q); err != nil {
		t.Fatalf("問題の保存に失敗: %v", err)
	}

	// コメントを直接 Firestore に作成
	commentsCol := client.Collection("teams").Doc(teamID).Collection("questions").Doc(questionID).Collection("comments")
	for i := 0; i < 3; i++ {
		if _, err := commentsCol.NewDoc().Set(ctx, map[string]any{"body": "comment", "index": i}); err != nil {
			t.Fatalf("コメントの作成に失敗: %v", err)
		}
	}
	if got := countDocs(ctx, t, commentsCol); got != 3 {
		t.Fatalf("セットアップ失敗: コメント数 got %d, want 3", got)
	}

	// 問題を削除
	if err := questionRepo.Delete(ctx, teamID, questionID); err != nil {
		t.Fatalf("問題の削除に失敗: %v", err)
	}

	// コメントが孤立していないことを確認
	if got := countDocs(ctx, t, commentsCol); got != 0 {
		t.Errorf("孤立コメントが残存しています: got %d, want 0", got)
	}

	// 問題自体が消えていることを確認
	_, err := client.Collection("teams").Doc(teamID).Collection("questions").Doc(questionID).Get(ctx)
	if status.Code(err) != codes.NotFound {
		t.Errorf("問題ドキュメントが残存しています")
	}
}

// TestNoteDelete_CascadesComments はノート削除時にコメントが孤立しないことを確認します。
func TestNoteDelete_CascadesComments(t *testing.T) {
	ctx := context.Background()
	client, _, noteRepo, _ := setupCascadeTest(t)

	teamID := "test-team-nd-" + time.Now().Format("150405")
	noteID := "test-note-001"

	n := &domain.Note{
		ID: noteID, TeamID: teamID, Title: "テストノート",
		Body: "本文", CreatedBy: "user-1",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := noteRepo.Save(ctx, n); err != nil {
		t.Fatalf("ノートの保存に失敗: %v", err)
	}

	commentsCol := client.Collection("teams").Doc(teamID).Collection("notes").Doc(noteID).Collection("comments")
	for i := 0; i < 2; i++ {
		if _, err := commentsCol.NewDoc().Set(ctx, map[string]any{"body": "comment", "index": i}); err != nil {
			t.Fatalf("コメントの作成に失敗: %v", err)
		}
	}

	if err := noteRepo.Delete(ctx, teamID, noteID); err != nil {
		t.Fatalf("ノートの削除に失敗: %v", err)
	}

	if got := countDocs(ctx, t, commentsCol); got != 0 {
		t.Errorf("孤立コメントが残存しています: got %d, want 0", got)
	}
}

// TestTeamDelete_CascadesAll はチーム削除時に全サブコレクションが孤立しないことを確認します。
func TestTeamDelete_CascadesAll(t *testing.T) {
	ctx := context.Background()
	client, questionRepo, noteRepo, teamRepo := setupCascadeTest(t)

	teamID := "test-team-td-" + time.Now().Format("150405")

	// チームを作成
	team := &domain.Team{
		ID: teamID, Name: "テストチーム", OwnerID: "owner-1",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := teamRepo.Save(ctx, team); err != nil {
		t.Fatalf("チームの保存に失敗: %v", err)
	}

	// 問題 + コメントを作成
	q := &domain.Question{
		ID: "q-001", TeamID: teamID, Title: "問題", Body: "本文", Answer: "回答",
		CreatedBy: "user-1", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := questionRepo.Save(ctx, q); err != nil {
		t.Fatalf("問題の保存に失敗: %v", err)
	}
	qCommentsCol := client.Collection("teams").Doc(teamID).Collection("questions").Doc("q-001").Collection("comments")
	if _, err := qCommentsCol.NewDoc().Set(ctx, map[string]any{"body": "qcomment"}); err != nil {
		t.Fatalf("問題コメントの作成に失敗: %v", err)
	}

	// ノート + コメントを作成
	n := &domain.Note{
		ID: "n-001", TeamID: teamID, Title: "ノート", Body: "本文",
		CreatedBy: "user-1", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := noteRepo.Save(ctx, n); err != nil {
		t.Fatalf("ノートの保存に失敗: %v", err)
	}
	nCommentsCol := client.Collection("teams").Doc(teamID).Collection("notes").Doc("n-001").Collection("comments")
	if _, err := nCommentsCol.NewDoc().Set(ctx, map[string]any{"body": "ncomment"}); err != nil {
		t.Fatalf("ノートコメントの作成に失敗: %v", err)
	}

	// タグを作成
	tagsCol := client.Collection("teams").Doc(teamID).Collection("tags")
	if _, err := tagsCol.NewDoc().Set(ctx, map[string]any{"name": "tag1"}); err != nil {
		t.Fatalf("タグの作成に失敗: %v", err)
	}

	// チームを削除
	if err := teamRepo.Delete(ctx, teamID); err != nil {
		t.Fatalf("チームの削除に失敗: %v", err)
	}

	// 全サブコレクションが空になっていることを確認
	checks := map[string]*fs.CollectionRef{
		"questions":        client.Collection("teams").Doc(teamID).Collection("questions"),
		"question comments": qCommentsCol,
		"notes":            client.Collection("teams").Doc(teamID).Collection("notes"),
		"note comments":    nCommentsCol,
		"tags":             tagsCol,
	}
	for name, col := range checks {
		if got := countDocs(ctx, t, col); got != 0 {
			t.Errorf("孤立ドキュメントが残存 [%s]: got %d, want 0", name, got)
		}
	}
}
