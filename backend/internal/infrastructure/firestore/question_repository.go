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

// questionRecord はFirestoreに保存する問題レコードです。
// コレクション: teams/{teamID}/questions/{questionID}
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

func toQuestionRecord(q *domain.Question) questionRecord {
	tags := q.Tags
	if tags == nil {
		tags = []string{}
	}
	return questionRecord{
		ID:          q.ID,
		TeamID:      q.TeamID,
		Title:       q.Title,
		Body:        q.Body,
		Answer:      q.Answer,
		Explanation: q.Explanation,
		Memo:        q.Memo,
		Tags:        tags,
		Status:      string(q.Status),
		CreatedBy:   q.CreatedBy,
		CreatedAt:   q.CreatedAt,
		UpdatedAt:   q.UpdatedAt,
	}
}

func toQuestion(r questionRecord) *domain.Question {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}

	// 後方互換: status フィールドが空の場合はデフォルト値を設定
	st := domain.QuestionStatus(r.Status)
	if st == "" {
		st = domain.QuestionStatusDraft
	}

	return &domain.Question{
		ID:          r.ID,
		TeamID:      r.TeamID,
		Title:       r.Title,
		Body:        r.Body,
		Answer:      r.Answer,
		Explanation: r.Explanation,
		Memo:        r.Memo,
		Tags:        tags,
		Status:      st,
		CreatedBy:   r.CreatedBy,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// FirestoreQuestionRepository はCloud Firestoreに問題データを永続化するリポジトリです。
// domain.QuestionRepository インターフェースを実装します。
type FirestoreQuestionRepository struct {
	client *fs.Client
}

// NewFirestoreQuestionRepository は FirestoreQuestionRepository を生成します。
func NewFirestoreQuestionRepository(client *fs.Client) *FirestoreQuestionRepository {
	return &FirestoreQuestionRepository{client: client}
}

func (r *FirestoreQuestionRepository) questionsCol(teamID string) *fs.CollectionRef {
	return r.client.Collection("teams").Doc(teamID).Collection("questions")
}

// FindByID はチームIDとIDで問題を検索します。
func (r *FirestoreQuestionRepository) FindByID(ctx context.Context, teamID, id string) (*domain.Question, error) {
	doc, err := r.questionsCol(teamID).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrQuestionNotFound
		}
		return nil, fmt.Errorf("問題の取得に失敗しました: %w", err)
	}

	var rec questionRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("問題データのデコードに失敗しました: %w", err)
	}
	return toQuestion(rec), nil
}

// ListByTeam は指定チームの問題一覧を返します。
func (r *FirestoreQuestionRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Question, error) {
	iter := r.questionsCol(teamID).Documents(ctx)
	defer iter.Stop()

	return r.collectQuestions(iter)
}

// SearchByTeam は指定チームの問題を検索・フィルタリングして返します。
// フィルタ条件が空の場合はチーム全件を返します。
func (r *FirestoreQuestionRepository) SearchByTeam(ctx context.Context, teamID string, filter domain.QuestionSearchFilter) ([]*domain.Question, error) {
	all, err := r.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	var result []*domain.Question
	for _, q := range all {
		// タグANDフィルタリング
		if len(filter.TagIDs) > 0 {
			tagSet := make(map[string]struct{}, len(q.Tags))
			for _, t := range q.Tags {
				tagSet[t] = struct{}{}
			}
			match := true
			for _, tid := range filter.TagIDs {
				if _, ok := tagSet[tid]; !ok {
					match = false
					break
				}
			}
			if !match {
				continue
			}
		}

		// キーワード検索
		if filter.Keyword != "" {
			kw := filter.Keyword
			if !strings.Contains(q.Title, kw) &&
				!strings.Contains(q.Body, kw) &&
				!strings.Contains(q.Explanation, kw) &&
				!strings.Contains(q.Memo, kw) {
				continue
			}
		}

		result = append(result, q)
	}

	if result == nil {
		result = []*domain.Question{}
	}
	return result, nil
}

// FindByTagID は指定チームの指定タグIDを持つ問題の一覧を返します。
func (r *FirestoreQuestionRepository) FindByTagID(ctx context.Context, teamID, tagID string) ([]*domain.Question, error) {
	// Firestore の array-contains クエリを使用
	iter := r.questionsCol(teamID).Where("tags", "array-contains", tagID).Documents(ctx)
	defer iter.Stop()

	return r.collectQuestions(iter)
}

// Save は問題を新規作成または更新します（Upsert）。
func (r *FirestoreQuestionRepository) Save(ctx context.Context, question *domain.Question) error {
	rec := toQuestionRecord(question)
	_, err := r.questionsCol(question.TeamID).Doc(question.ID).Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("問題の保存に失敗しました: %w", err)
	}
	return nil
}

// Delete はチームIDとIDで指定した問題を削除します。
func (r *FirestoreQuestionRepository) Delete(ctx context.Context, teamID, id string) error {
	ref := r.questionsCol(teamID).Doc(id)

	_, err := ref.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrQuestionNotFound
		}
		return fmt.Errorf("問題の存在確認に失敗しました: %w", err)
	}

	_, err = ref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("問題の削除に失敗しました: %w", err)
	}
	return nil
}

func (r *FirestoreQuestionRepository) collectQuestions(iter *fs.DocumentIterator) ([]*domain.Question, error) {
	var questions []*domain.Question
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("問題一覧の取得に失敗しました: %w", err)
		}

		var rec questionRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("問題データのデコードに失敗しました: %w", err)
		}
		questions = append(questions, toQuestion(rec))
	}

	if questions == nil {
		questions = []*domain.Question{}
	}
	return questions, nil
}

// FirestoreQuestionRepository が domain.QuestionRepository を実装していることをコンパイル時に保証します。
var _ domain.QuestionRepository = (*FirestoreQuestionRepository)(nil)
