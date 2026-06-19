import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { fetchQuestions } from '../utils/questionApi';
import { fetchTags } from '../utils/tagApi';
import { Question } from '../types/question';
import { Tag } from '../types/tag';
import TagChip from '../components/TagChip';

const PER_PAGE = 20;

export default function QuestionListPage() {
  const { t } = useTranslation();

  const [questions, setQuestions] = useState<Question[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [keyword, setKeyword] = useState('');
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  const loadQuestions = useCallback(
    (currentPage: number, currentKeyword: string, currentTagIds: string[]) => {
      let isMounted = true;
      setIsLoading(true);
      setLoadError('');

      const params: { page: number; per_page: number; keyword?: string; tag_ids?: string } = {
        page: currentPage,
        per_page: PER_PAGE,
      };
      if (currentKeyword.trim()) {
        params.keyword = currentKeyword.trim();
      }
      if (currentTagIds.length > 0) {
        params.tag_ids = currentTagIds.join(',');
      }

      fetchQuestions(params)
        .then((data) => {
          if (isMounted) {
            setQuestions(data.items);
            setTotalPages(data.total_pages);
          }
        })
        .catch(() => {
          if (isMounted) setLoadError(t('question.error.fetchFailed'));
        })
        .finally(() => {
          if (isMounted) setIsLoading(false);
        });

      return () => {
        isMounted = false;
      };
    },
    [t]
  );

  useEffect(() => {
    fetchTags()
      .then(setTags)
      .catch(() => {
        // タグ取得エラーは非致命的。無視して続行。
      });
  }, []);

  useEffect(() => {
    const cleanup = loadQuestions(page, keyword, selectedTagIds);
    return cleanup;
  }, [page, keyword, selectedTagIds, loadQuestions]);

  const handleTagToggle = (tagId: string) => {
    setSelectedTagIds((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
    setPage(1);
  };

  const handleKeywordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setKeyword(e.target.value);
    setPage(1);
  };

  return (
    <section className="question-list-page">
      <div className="question-list-header">
        <h1 className="page-title">{t('nav.questions')}</h1>
        <Link to="/questions/new" className="btn btn-primary">
          {t('question.new')}
        </Link>
      </div>

      <div className="question-list-filters">
        <input
          type="search"
          className="filter-keyword-input"
          placeholder={t('question.filter.keywordPlaceholder')}
          value={keyword}
          onChange={handleKeywordChange}
          aria-label={t('question.filter.keywordPlaceholder')}
        />

        {tags.length > 0 && (
          <div
            className="filter-tags"
            role="group"
            aria-label={t('question.filter.tagPlaceholder')}
          >
            {tags.map((tag) => (
              <TagChip
                key={tag.id}
                tag={tag}
                selected={selectedTagIds.includes(tag.id)}
                onToggle={handleTagToggle}
              />
            ))}
          </div>
        )}
      </div>

      {isLoading ? (
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      ) : loadError ? (
        <p role="alert" className="alert alert-error">
          {loadError}
        </p>
      ) : questions.length === 0 ? (
        <p className="question-list-empty">{t('question.list.empty')}</p>
      ) : (
        <>
          <ul className="question-list">
            {questions.map((q) => (
              <li key={q.id} className="question-list-item">
                <Link to={`/questions/${q.id}`} className="question-list-link">
                  <span className="question-list-title">{q.title}</span>
                  <div className="question-list-meta">
                    <span className="question-date">
                      {new Date(q.created_at).toLocaleDateString('ja-JP')}
                    </span>
                    {q.tags.length > 0 && (
                      <ul className="question-tags" aria-label={t('question.tagsLabel')}>
                        {q.tags.map((tag) => (
                          <li key={tag} className="tag-badge">
                            {tag}
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                </Link>
              </li>
            ))}
          </ul>

          {totalPages > 1 && (
            <div className="pagination" role="navigation" aria-label="ページネーション">
              <button
                type="button"
                className="btn btn-secondary"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                aria-label={t('question.pagination.prev')}
              >
                {t('question.pagination.prev')}
              </button>
              <span className="pagination-info">
                {t('question.pagination.pageInfo', { page, total: totalPages })}
              </span>
              <button
                type="button"
                className="btn btn-secondary"
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page >= totalPages}
                aria-label={t('question.pagination.next')}
              >
                {t('question.pagination.next')}
              </button>
            </div>
          )}
        </>
      )}
    </section>
  );
}
