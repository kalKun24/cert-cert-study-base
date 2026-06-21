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

// commentRecord はFirestoreに保存するコメントレコードです。
// コレクション: teams/{teamID}/questions/{questionID}/comments/{commentID}
type commentRecord struct {
	ID         string    `firestore:"id"`
	QuestionID string    `firestore:"question_id"`
	Body       string    `firestore:"body"`
	CreatedBy  string    `firestore:"created_by"`
	CreatedAt  time.Time `firestore:"created_at"`
	UpdatedAt  time.Time `firestore:"updated_at"`
}

func toCommentRecord(c *domain.Comment) commentRecord {
	return commentRecord{
		ID:         c.ID,
		QuestionID: c.QuestionID,
		Body:       c.Body,
		CreatedBy:  c.CreatedBy,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

func toComment(r commentRecord) *domain.Comment {
	return &domain.Comment{
		ID:         r.ID,
		QuestionID: r.QuestionID,
		Body:       r.Body,
		CreatedBy:  r.CreatedBy,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// FirestoreCommentRepository はCloud Firestoreにコメントデータを永続化するリポジトリです。
// domain.CommentRepository インターフェースを実装します。
type FirestoreCommentRepository struct {
	client *fs.Client
}

// NewFirestoreCommentRepository は FirestoreCommentRepository を生成します。
func NewFirestoreCommentRepository(client *fs.Client) *FirestoreCommentRepository {
	return &FirestoreCommentRepository{client: client}
}

func (r *FirestoreCommentRepository) commentsCol(teamID, questionID string) *fs.CollectionRef {
	return r.client.Collection("teams").Doc(teamID).Collection("questions").Doc(questionID).Collection("comments")
}

// FindByID はIDでコメントを検索します。
func (r *FirestoreCommentRepository) FindByID(ctx context.Context, teamID, questionID, commentID string) (*domain.Comment, error) {
	doc, err := r.commentsCol(teamID, questionID).Doc(commentID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrCommentNotFound
		}
		return nil, fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}

	var rec commentRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("コメントデータのデコードに失敗しました: %w", err)
	}
	return toComment(rec), nil
}

// ListByQuestionID は指定したチームIDと問題IDのコメント一覧を返します。
// ソートはユースケース層で行うため、ここでは順序を保証しません。
func (r *FirestoreCommentRepository) ListByQuestionID(ctx context.Context, teamID, questionID string) ([]*domain.Comment, error) {
	iter := r.commentsCol(teamID, questionID).Documents(ctx)
	defer iter.Stop()

	var comments []*domain.Comment
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("コメント一覧の取得に失敗しました: %w", err)
		}

		var rec commentRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("コメントデータのデコードに失敗しました: %w", err)
		}
		comments = append(comments, toComment(rec))
	}

	if comments == nil {
		comments = []*domain.Comment{}
	}
	return comments, nil
}

// Save はコメントを新規作成または更新します（Upsert）。
func (r *FirestoreCommentRepository) Save(ctx context.Context, teamID string, comment *domain.Comment) error {
	rec := toCommentRecord(comment)
	_, err := r.commentsCol(teamID, comment.QuestionID).Doc(comment.ID).Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("コメントの保存に失敗しました: %w", err)
	}
	return nil
}

// Delete はIDで指定したコメントを削除します。
func (r *FirestoreCommentRepository) Delete(ctx context.Context, teamID, questionID, commentID string) error {
	ref := r.commentsCol(teamID, questionID).Doc(commentID)

	_, err := ref.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrCommentNotFound
		}
		return fmt.Errorf("コメントの存在確認に失敗しました: %w", err)
	}

	_, err = ref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("コメントの削除に失敗しました: %w", err)
	}
	return nil
}

// FirestoreCommentRepository が domain.CommentRepository を実装していることをコンパイル時に保証します。
var _ domain.CommentRepository = (*FirestoreCommentRepository)(nil)
