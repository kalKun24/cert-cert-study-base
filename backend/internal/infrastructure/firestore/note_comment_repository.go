package firestore

import (
	"context"
	"fmt"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// noteCommentRecord はFirestoreに保存するノートコメントレコードです。
// コレクション: teams/{teamID}/notes/{noteID}/comments/{commentID}
type noteCommentRecord struct {
	ID        string    `firestore:"id"`
	NoteID    string    `firestore:"note_id"`
	Body      string    `firestore:"body"`
	CreatedBy string    `firestore:"created_by"`
	CreatedAt time.Time `firestore:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at"`
}

func toNoteCommentRecord(c *domain.NoteComment) noteCommentRecord {
	return noteCommentRecord{
		ID:        c.ID,
		NoteID:    c.NoteID,
		Body:      c.Body,
		CreatedBy: c.CreatedBy,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toNoteComment(r noteCommentRecord) *domain.NoteComment {
	return &domain.NoteComment{
		ID:        r.ID,
		NoteID:    r.NoteID,
		Body:      r.Body,
		CreatedBy: r.CreatedBy,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// FirestoreNoteCommentRepository はCloud Firestoreにノートコメントデータを永続化するリポジトリです。
// domain.NoteCommentRepository インターフェースを実装します。
type FirestoreNoteCommentRepository struct {
	client *fs.Client
}

// NewFirestoreNoteCommentRepository は FirestoreNoteCommentRepository を生成します。
func NewFirestoreNoteCommentRepository(client *fs.Client) *FirestoreNoteCommentRepository {
	return &FirestoreNoteCommentRepository{client: client}
}

func (r *FirestoreNoteCommentRepository) commentsCol(teamID, noteID string) *fs.CollectionRef {
	return r.client.Collection("teams").Doc(teamID).Collection("notes").Doc(noteID).Collection("comments")
}

// FindByID はIDでノートコメントを検索します。
func (r *FirestoreNoteCommentRepository) FindByID(ctx context.Context, teamID, noteID, commentID string) (*domain.NoteComment, error) {
	doc, err := r.commentsCol(teamID, noteID).Doc(commentID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrNoteCommentNotFound
		}
		return nil, fmt.Errorf("ノートコメントの取得に失敗しました: %w", err)
	}

	var rec noteCommentRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("ノートコメントデータのデコードに失敗しました: %w", err)
	}
	return toNoteComment(rec), nil
}

// ListByNoteID は指定したチームIDとノートIDのコメント一覧を返します。
// ソートはユースケース層で行うため、ここでは順序を保証しません。
func (r *FirestoreNoteCommentRepository) ListByNoteID(ctx context.Context, teamID, noteID string) ([]*domain.NoteComment, error) {
	iter := r.commentsCol(teamID, noteID).Documents(ctx)
	defer iter.Stop()

	var comments []*domain.NoteComment
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ノートコメント一覧の取得に失敗しました: %w", err)
		}

		var rec noteCommentRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("ノートコメントデータのデコードに失敗しました: %w", err)
		}
		comments = append(comments, toNoteComment(rec))
	}

	if comments == nil {
		comments = []*domain.NoteComment{}
	}
	return comments, nil
}

// Save はノートコメントを新規作成または更新します（Upsert）。
func (r *FirestoreNoteCommentRepository) Save(ctx context.Context, teamID string, comment *domain.NoteComment) error {
	rec := toNoteCommentRecord(comment)
	_, err := r.commentsCol(teamID, comment.NoteID).Doc(comment.ID).Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("ノートコメントの保存に失敗しました: %w", err)
	}
	return nil
}

// Delete はIDで指定したノートコメントを削除します。
func (r *FirestoreNoteCommentRepository) Delete(ctx context.Context, teamID, noteID, commentID string) error {
	ref := r.commentsCol(teamID, noteID).Doc(commentID)

	_, err := ref.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrNoteCommentNotFound
		}
		return fmt.Errorf("ノートコメントの存在確認に失敗しました: %w", err)
	}

	_, err = ref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("ノートコメントの削除に失敗しました: %w", err)
	}
	return nil
}

// FirestoreNoteCommentRepository が domain.NoteCommentRepository を実装していることをコンパイル時に保証します。
var _ domain.NoteCommentRepository = (*FirestoreNoteCommentRepository)(nil)
