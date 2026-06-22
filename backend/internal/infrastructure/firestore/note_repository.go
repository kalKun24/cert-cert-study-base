package firestore

import (
	"context"
	"fmt"
	"strings"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// noteRecord はFirestoreに保存するノートレコードです。
// コレクション: teams/{teamID}/notes/{noteID}
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

func toNoteRecord(n *domain.Note) noteRecord {
	tags := n.Tags
	if tags == nil {
		tags = []string{}
	}
	return noteRecord{
		ID:               n.ID,
		TeamID:           n.TeamID,
		Title:            n.Title,
		Body:             n.Body,
		DiscussionPoints: n.DiscussionPoints,
		Memo:             n.Memo,
		Tags:             tags,
		Status:           string(n.Status),
		CreatedBy:        n.CreatedBy,
		CreatedAt:        n.CreatedAt,
		UpdatedAt:        n.UpdatedAt,
	}
}

func toNote(r noteRecord) *domain.Note {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}

	// 後方互換: status フィールドが空の場合はデフォルト値を設定
	st := domain.NoteStatus(r.Status)
	if st == "" {
		st = domain.NoteStatusDraft
	}

	return &domain.Note{
		ID:               r.ID,
		TeamID:           r.TeamID,
		Title:            r.Title,
		Body:             r.Body,
		DiscussionPoints: r.DiscussionPoints,
		Memo:             r.Memo,
		Tags:             tags,
		Status:           st,
		CreatedBy:        r.CreatedBy,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

// FirestoreNoteRepository はCloud Firestoreにノートデータを永続化するリポジトリです。
// domain.NoteRepository インターフェースを実装します。
type FirestoreNoteRepository struct {
	client *fs.Client
}

// NewFirestoreNoteRepository は FirestoreNoteRepository を生成します。
func NewFirestoreNoteRepository(client *fs.Client) *FirestoreNoteRepository {
	return &FirestoreNoteRepository{client: client}
}

func (r *FirestoreNoteRepository) notesCol(teamID string) *fs.CollectionRef {
	return r.client.Collection("teams").Doc(teamID).Collection("notes")
}

// FindByID はチームIDとIDでノートを検索します。
func (r *FirestoreNoteRepository) FindByID(ctx context.Context, teamID, id string) (*domain.Note, error) {
	doc, err := r.notesCol(teamID).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrNoteNotFound
		}
		return nil, fmt.Errorf("ノートの取得に失敗しました: %w", err)
	}

	var rec noteRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("ノートデータのデコードに失敗しました: %w", err)
	}
	return toNote(rec), nil
}

// ListByTeam は指定チームのノート一覧を返します。
func (r *FirestoreNoteRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Note, error) {
	iter := r.notesCol(teamID).Documents(ctx)
	defer iter.Stop()

	return r.collectNotes(iter)
}

// SearchByTeam は指定チームのノートを検索・フィルタリングして返します。
// フィルタ条件が空の場合はチーム全件を返します。
//
// タグフィルタ最適化:
//   - TagIDs が空でない場合、最初の1タグを Firestore の array-contains クエリでサーバーサイド絞り込みする。
//     Firestore は1クエリにつき array-contains を1つしか使用できないため、残りのタグはメモリフィルタで AND 処理する。
//   - TagIDs が空の場合は全件取得してメモリフィルタのみ適用する。
func (r *FirestoreNoteRepository) SearchByTeam(ctx context.Context, teamID string, filter domain.NoteSearchFilter) ([]*domain.Note, error) {
	var (
		candidates []*domain.Note
		err        error
	)

	if len(filter.TagIDs) > 0 {
		// 最初の1タグをサーバーサイドで絞り込む（Firestore は array-contains を1クエリ1つのみ許可）
		iter := r.notesCol(teamID).Where("tags", "array-contains", filter.TagIDs[0]).Documents(ctx)
		defer iter.Stop()
		candidates, err = r.collectNotes(iter)
	} else {
		candidates, err = r.ListByTeam(ctx, teamID)
	}
	if err != nil {
		return nil, err
	}

	var result []*domain.Note
	for _, n := range candidates {
		// 残りのタグ（index 1 以降）をメモリフィルタで AND 処理
		if len(filter.TagIDs) > 1 {
			tagSet := make(map[string]struct{}, len(n.Tags))
			for _, t := range n.Tags {
				tagSet[t] = struct{}{}
			}
			match := true
			for _, tid := range filter.TagIDs[1:] {
				if _, ok := tagSet[tid]; !ok {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		// キーワード検索: Firestore はネイティブな全文検索をサポートしないため、
		// 取得した全ドキュメントをメモリ内でスキャンする。
		// スケール限界: チームあたり数百件以上になると応答時間が線形増加し、
		// メモリ使用量も増大する。将来的には Algolia や Typesense 等の
		// 外部全文検索エンジンへの移行を検討すること。
		if filter.Keyword != "" {
			kw := filter.Keyword
			if !strings.Contains(n.Title, kw) &&
				!strings.Contains(n.Body, kw) &&
				!strings.Contains(n.DiscussionPoints, kw) &&
				!strings.Contains(n.Memo, kw) {
				continue
			}
		}

		result = append(result, n)
	}

	if result == nil {
		result = []*domain.Note{}
	}
	return result, nil
}

// FindByTagID は指定チームの指定タグIDを持つノートの一覧を返します。
func (r *FirestoreNoteRepository) FindByTagID(ctx context.Context, teamID, tagID string) ([]*domain.Note, error) {
	iter := r.notesCol(teamID).Where("tags", "array-contains", tagID).Documents(ctx)
	defer iter.Stop()

	return r.collectNotes(iter)
}

// Save はノートを新規作成または更新します（Upsert）。
func (r *FirestoreNoteRepository) Save(ctx context.Context, note *domain.Note) error {
	rec := toNoteRecord(note)
	_, err := r.notesCol(note.TeamID).Doc(note.ID).Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("ノートの保存に失敗しました: %w", err)
	}
	return nil
}

// Delete はチームIDとIDで指定したノートとそのコメントサブコレクションを削除します。
// Firestoreはドキュメント削除時にサブコレクションを自動削除しないため、明示的に削除します。
func (r *FirestoreNoteRepository) Delete(ctx context.Context, teamID, id string) error {
	ref := r.notesCol(teamID).Doc(id)

	_, err := ref.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrNoteNotFound
		}
		return fmt.Errorf("ノートの存在確認に失敗しました: %w", err)
	}

	if err := deleteSubCollection(ctx, r.client, ref.Collection("comments")); err != nil {
		return fmt.Errorf("ノートコメントの削除に失敗しました: %w", err)
	}

	_, err = ref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("ノートの削除に失敗しました: %w", err)
	}
	return nil
}

func (r *FirestoreNoteRepository) collectNotes(iter *fs.DocumentIterator) ([]*domain.Note, error) {
	var notes []*domain.Note
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ノート一覧の取得に失敗しました: %w", err)
		}

		var rec noteRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("ノートデータのデコードに失敗しました: %w", err)
		}
		notes = append(notes, toNote(rec))
	}

	if notes == nil {
		notes = []*domain.Note{}
	}
	return notes, nil
}

// FirestoreNoteRepository が domain.NoteRepository を実装していることをコンパイル時に保証します。
var _ domain.NoteRepository = (*FirestoreNoteRepository)(nil)
