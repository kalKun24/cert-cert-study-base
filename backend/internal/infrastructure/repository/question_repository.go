// Package repository はドメインリポジトリインターフェースの具体実装を提供します。
package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/kalKun24/cert-study-base/backend/internal/domain"
	"github.com/kalKun24/cert-study-base/backend/internal/infrastructure/storage"
)

// questionsObjectName はGCSバケット内で問題データを保存するオブジェクト名です。
const questionsObjectName = "questions.json"

// questionRecord はGCS上のJSONファイルに保存する問題レコードです。
// domain.Question と対応しており、JSON直列化のための構造体です。
type questionRecord struct {
	ID          string    `json:"id"`
	TeamID      string    `json:"team_id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Answer      string    `json:"answer"`
	Explanation string    `json:"explanation"`
	Memo        string    `json:"memo"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// toQuestionRecord はドメインエンティティをJSONレコードに変換します。
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

// toQuestion はJSONレコードをドメインエンティティに変換します。
// 後方互換性のため、status が空の場合はデフォルト値を設定します。
// （既存の questions.json にフィールドがない場合でも安全に動作します）
func toQuestion(r questionRecord) *domain.Question {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}

	// 後方互換: status フィールドが空の場合はデフォルト値を設定
	status := domain.QuestionStatus(r.Status)
	if status == "" {
		status = domain.QuestionStatusDraft
	}

	return &domain.Question{
		ID:          r.ID,
		TeamID:      r.TeamID, // 既存データでは空文字のまま（後方互換）
		Title:       r.Title,
		Body:        r.Body,
		Answer:      r.Answer,
		Explanation: r.Explanation,
		Memo:        r.Memo,
		Tags:        tags,
		Status:      status,
		CreatedBy:   r.CreatedBy,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// GCSQuestionRepository はGCS上のJSONファイルに問題データを永続化するリポジトリです。
// domain.QuestionRepository インターフェースを実装します。
// 問題データは GCS バケット内の questions.json に保存します。
type GCSQuestionRepository struct {
	mu      sync.RWMutex
	storage storage.StorageClient
	bucket  string
}

// NewGCSQuestionRepository は GCSQuestionRepository を生成します。
func NewGCSQuestionRepository(sc storage.StorageClient, bucket string) *GCSQuestionRepository {
	return &GCSQuestionRepository{
		storage: sc,
		bucket:  bucket,
	}
}

// loadQuestions はGCSから問題データを読み込みます。
// オブジェクトが存在しない場合は空のスライスを返します。
func (r *GCSQuestionRepository) loadQuestions(ctx context.Context) ([]questionRecord, error) {
	exists, err := r.storage.Exists(ctx, r.bucket, questionsObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS オブジェクト存在確認に失敗しました: %w", err)
	}

	if !exists {
		return []questionRecord{}, nil
	}

	rc, err := r.storage.Read(ctx, r.bucket, questionsObjectName)
	if err != nil {
		return nil, fmt.Errorf("GCS からの読み込みに失敗しました: %w", err)
	}
	defer func() {
		if cerr := rc.Close(); cerr != nil {
			slog.Warn("GCS ReadCloser のクローズに失敗しました", "error", cerr)
		}
	}()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("GCS データの読み取りに失敗しました: %w", err)
	}

	var records []questionRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("問題データのJSONデコードに失敗しました: %w", err)
	}

	return records, nil
}

// saveQuestions は問題データをGCSに書き込みます。
func (r *GCSQuestionRepository) saveQuestions(ctx context.Context, records []questionRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("問題データのJSONエンコードに失敗しました: %w", err)
	}

	if err := r.storage.Write(ctx, r.bucket, questionsObjectName, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("GCS への書き込みに失敗しました: %w", err)
	}

	return nil
}

// FindByID はIDで問題を検索します。
func (r *GCSQuestionRepository) FindByID(ctx context.Context, id string) (*domain.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadQuestions(ctx)
	if err != nil {
		return nil, fmt.Errorf("問題データ読み込みに失敗しました: %w", err)
	}

	for _, rec := range records {
		if rec.ID == id {
			return toQuestion(rec), nil
		}
	}

	return nil, domain.ErrQuestionNotFound
}

// ListByTeam は指定チームの問題一覧を返します。
func (r *GCSQuestionRepository) ListByTeam(ctx context.Context, teamID string) ([]*domain.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadQuestions(ctx)
	if err != nil {
		return nil, fmt.Errorf("問題データ読み込みに失敗しました: %w", err)
	}

	questions := make([]*domain.Question, 0)
	for _, rec := range records {
		if rec.TeamID == teamID {
			questions = append(questions, toQuestion(rec))
		}
	}

	return questions, nil
}

// SearchByTeam は指定チームの問題を検索・フィルタリングして返します。
// GCSから全件読み込み後、teamID フィルタ → タグ・キーワードフィルタリングを行います。
func (r *GCSQuestionRepository) SearchByTeam(ctx context.Context, teamID string, filter domain.QuestionSearchFilter) ([]*domain.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadQuestions(ctx)
	if err != nil {
		return nil, fmt.Errorf("問題データ読み込みに失敗しました: %w", err)
	}

	questions := make([]*domain.Question, 0)
	for _, rec := range records {
		// チームIDフィルタ
		if rec.TeamID != teamID {
			continue
		}

		q := toQuestion(rec)

		// タグANDフィルタリング: 指定されたタグIDをすべて持つ問題のみ返す
		if len(filter.TagIDs) > 0 {
			tagSet := make(map[string]struct{}, len(rec.Tags))
			for _, t := range rec.Tags {
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

		// キーワード検索: title / body / explanation / memo を対象とした部分一致
		if filter.Keyword != "" {
			kw := filter.Keyword
			if !strings.Contains(rec.Title, kw) &&
				!strings.Contains(rec.Body, kw) &&
				!strings.Contains(rec.Explanation, kw) &&
				!strings.Contains(rec.Memo, kw) {
				continue
			}
		}

		questions = append(questions, q)
	}

	return questions, nil
}

// FindByTagID は指定されたタグIDを持つ問題の一覧を返します。
func (r *GCSQuestionRepository) FindByTagID(ctx context.Context, tagID string) ([]*domain.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records, err := r.loadQuestions(ctx)
	if err != nil {
		return nil, fmt.Errorf("問題データ読み込みに失敗しました: %w", err)
	}

	var questions []*domain.Question
	for _, rec := range records {
		for _, tid := range rec.Tags {
			if tid == tagID {
				questions = append(questions, toQuestion(rec))
				break
			}
		}
	}

	return questions, nil
}

// Save は問題を新規作成または更新します。
// IDが一致するレコードが存在する場合は更新、存在しない場合は追加します。
func (r *GCSQuestionRepository) Save(ctx context.Context, question *domain.Question) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadQuestions(ctx)
	if err != nil {
		return fmt.Errorf("問題データ読み込みに失敗しました: %w", err)
	}

	rec := toQuestionRecord(question)
	updated := false
	for i, existing := range records {
		if existing.ID == question.ID {
			records[i] = rec
			updated = true
			break
		}
	}

	if !updated {
		records = append(records, rec)
	}

	if err := r.saveQuestions(ctx, records); err != nil {
		return fmt.Errorf("問題データ保存に失敗しました: %w", err)
	}

	return nil
}

// Delete はIDで指定した問題を削除します。
func (r *GCSQuestionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	records, err := r.loadQuestions(ctx)
	if err != nil {
		return fmt.Errorf("問題データ読み込みに失敗しました: %w", err)
	}

	newRecords := make([]questionRecord, 0, len(records))
	found := false
	for _, rec := range records {
		if rec.ID == id {
			found = true
			continue
		}
		newRecords = append(newRecords, rec)
	}

	if !found {
		return domain.ErrQuestionNotFound
	}

	if err := r.saveQuestions(ctx, newRecords); err != nil {
		return fmt.Errorf("問題データ保存に失敗しました: %w", err)
	}

	return nil
}

// GCSQuestionRepository が domain.QuestionRepository を実装していることをコンパイル時に保証します。
var _ domain.QuestionRepository = (*GCSQuestionRepository)(nil)
