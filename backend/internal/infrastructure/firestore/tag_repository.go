package firestore

import (
	"context"
	"fmt"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"google.golang.org/api/iterator"
)

// tagRecord はFirestoreに保存するタグレコードです。
// コレクション: teams/{teamID}/tags/{tagID}
type tagRecord struct {
	ID        string    `firestore:"id"`
	TeamID    string    `firestore:"team_id"`
	Name      string    `firestore:"name"`
	CreatedAt time.Time `firestore:"created_at"`
}

func toTagRecord(t *domain.Tag) tagRecord {
	return tagRecord{
		ID:        t.ID,
		TeamID:    t.TeamID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}
}

func toTag(r tagRecord) *domain.Tag {
	return &domain.Tag{
		ID:        r.ID,
		TeamID:    r.TeamID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
	}
}

// FirestoreTagRepository はCloud Firestoreにタグデータを永続化するリポジトリです。
// domain.TagRepository インターフェースを実装します。
// questionRepo と noteRepo を保持することで、Delete 時の使用中チェックを実施します。
type FirestoreTagRepository struct {
	client       *fs.Client
	questionRepo domain.QuestionRepository
	noteRepo     domain.NoteRepository
}

// NewFirestoreTagRepository は FirestoreTagRepository を生成します。
// questionRepo と noteRepo はタグ削除時の使用中チェックに使用します。
func NewFirestoreTagRepository(client *fs.Client, questionRepo domain.QuestionRepository, noteRepo domain.NoteRepository) *FirestoreTagRepository {
	return &FirestoreTagRepository{
		client:       client,
		questionRepo: questionRepo,
		noteRepo:     noteRepo,
	}
}

func (r *FirestoreTagRepository) tagsCol(teamID string) *fs.CollectionRef {
	return r.client.Collection("teams").Doc(teamID).Collection("tags")
}

// FindByID はIDでタグを検索します。
// タグはチームサブコレクションに保存されているため、全チームを横断する collectionGroup
// クエリを使用します。
func (r *FirestoreTagRepository) FindByID(ctx context.Context, id string) (*domain.Tag, error) {
	iter := r.client.CollectionGroup("tags").Where("id", "==", id).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, domain.ErrTagNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("タグの検索に失敗しました: %w", err)
	}

	var rec tagRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("タグデータのデコードに失敗しました: %w", err)
	}
	return toTag(rec), nil
}

// FindByName はチームIDとタグ名でタグを検索します。
func (r *FirestoreTagRepository) FindByName(ctx context.Context, teamID string, name string) (*domain.Tag, error) {
	iter := r.tagsCol(teamID).Where("name", "==", name).Limit(1).Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, domain.ErrTagNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("タグの検索に失敗しました: %w", err)
	}

	var rec tagRecord
	if err := doc.DataTo(&rec); err != nil {
		return nil, fmt.Errorf("タグデータのデコードに失敗しました: %w", err)
	}
	return toTag(rec), nil
}

// ListByTeam は指定チームのタグ一覧を返します。
func (r *FirestoreTagRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Tag, error) {
	iter := r.tagsCol(teamID).Documents(ctx)
	defer iter.Stop()

	var tags []*domain.Tag
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("タグ一覧の取得に失敗しました: %w", err)
		}

		var rec tagRecord
		if err := doc.DataTo(&rec); err != nil {
			return nil, fmt.Errorf("タグデータのデコードに失敗しました: %w", err)
		}
		tags = append(tags, toTag(rec))
	}

	if tags == nil {
		tags = []*domain.Tag{}
	}
	return tags, nil
}

// Save はタグを新規作成または更新します（Upsert）。
func (r *FirestoreTagRepository) Save(ctx context.Context, tag *domain.Tag) error {
	rec := toTagRecord(tag)
	_, err := r.tagsCol(tag.TeamID).Doc(tag.ID).Set(ctx, rec)
	if err != nil {
		return fmt.Errorf("タグの保存に失敗しました: %w", err)
	}
	return nil
}

// Delete はIDで指定したタグを削除します。
// 削除前に使用中チェックを行います。
// タグが問題に使用されている場合は domain.ErrTagInUse を返します。
func (r *FirestoreTagRepository) Delete(ctx context.Context, id string) error {
	// まずタグを取得してチームIDを確認
	tag, err := r.FindByID(ctx, id)
	if err != nil {
		return err // ErrTagNotFound を含む
	}

	// teamID の整合性チェック
	if tag.TeamID == "" {
		return fmt.Errorf("タグレコードのteamIDが空です（データ不整合）: %w", domain.ErrTagDataInconsistent)
	}

	// 使用中チェック: 問題またはノートが参照するタグは削除できない
	questions, err := r.questionRepo.FindByTagID(ctx, tag.TeamID, id)
	if err != nil {
		return fmt.Errorf("使用中チェック（問題）に失敗しました: %w", err)
	}
	if len(questions) > 0 {
		return domain.ErrTagInUse
	}

	notes, err := r.noteRepo.FindByTagID(ctx, tag.TeamID, id)
	if err != nil {
		return fmt.Errorf("使用中チェック（ノート）に失敗しました: %w", err)
	}
	if len(notes) > 0 {
		return domain.ErrTagInUse
	}

	// FindByID で存在確認済みのため再読み取り不要
	ref := r.tagsCol(tag.TeamID).Doc(id)
	_, err = ref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("タグの削除に失敗しました: %w", err)
	}
	return nil
}

// FirestoreTagRepository が domain.TagRepository を実装していることをコンパイル時に保証します。
var _ domain.TagRepository = (*FirestoreTagRepository)(nil)
