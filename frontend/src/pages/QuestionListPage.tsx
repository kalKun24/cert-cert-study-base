import { useState, useEffect, useCallback, useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { fetchQuestions } from '../utils/questionApi';
import { fetchTags } from '../utils/tagApi';
import { useTeam } from '../context/TeamContext';
import { Question } from '../types/question';
import { Tag } from '../types/tag';
import TagChip from '../components/TagChip';
import Paginator from '../components/Paginator';
import { QuestionListSkeleton } from '../components/Skeleton';

const PER_PAGE = 20;

/** URL クエリの tag_ids パラメータ名 */
const PARAM_TAG_IDS = 'tag_ids';
/** URL クエリの keyword パラメータ名 */
const PARAM_KEYWORD = 'keyword';
/** URL クエリの page パラメータ名 */
const PARAM_PAGE = 'page';

export default function QuestionListPage() {
  const { t } = useTranslation();
  const { activeTeam } = useTeam();
  const [searchParams, setSearchParams] = useSearchParams();

  // URL クエリパラメータから初期値を復元
  const initialTagIds = (): string[] => {
    const raw = searchParams.get(PARAM_TAG_IDS);
    if (!raw) return [];
    return raw.split(',').filter(Boolean);
  };
  const initialKeyword = searchParams.get(PARAM_KEYWORD) ?? '';
  const initialPage = parseInt(searchParams.get(PARAM_PAGE) ?? '1', 10) || 1;

  const [questions, setQuestions] = useState<Question[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const tagMap = useMemo(() => new Map(tags.map((t) => [t.id, t.name])), [tags]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [keyword, setKeyword] = useState(initialKeyword);
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>(initialTagIds);
  const [page, setPage] = useState(initialPage);
  const [totalPages, setTotalPages] = useState(1);

  /** URL クエリパラメータを現在の状態で同期する */
  const syncSearchParams = useCallback(
    (currentPage: number, currentKeyword: string, currentTagIds: string[]) => {
      const next = new URLSearchParams();
      if (currentKeyword.trim()) next.set(PARAM_KEYWORD, currentKeyword.trim());
      if (currentTagIds.length > 0) next.set(PARAM_TAG_IDS, currentTagIds.join(','));
      if (currentPage > 1) next.set(PARAM_PAGE, String(currentPage));
      setSearchParams(next, { replace: true });
    },
    [setSearchParams]
  );

  const loadQuestions = useCallback(
    (currentPage: number, currentKeyword: string, currentTagIds: string[]) => {
      if (!activeTeam) return;
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

      fetchQuestions(activeTeam.id, params)
        .then((data) => {
          if (isMounted) {
            setQuestions(data.items);
            setTotalPages(data.total_pages);
          }
        })
        .catch(() => {
          if (isMounted) {
            setLoadError(t('question.error.fetchFailed'));
            setTotalPages(1);
          }
        })
        .finally(() => {
          if (isMounted) setIsLoading(false);
        });

      return () => {
        isMounted = false;
      };
    },
    [activeTeam, t]
  );

  useEffect(() => {
    if (!activeTeam) return;
    fetchTags(activeTeam.id)
      .then(setTags)
      .catch(() => {
        // タグ取得エラーは非致命的。無視して続行。
      });
  }, [activeTeam?.id]);

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [page]);

  useEffect(() => {
    syncSearchParams(page, keyword, selectedTagIds);
    const cleanup = loadQuestions(page, keyword, selectedTagIds);
    return cleanup;
  }, [page, keyword, selectedTagIds, loadQuestions, syncSearchParams]);

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
    <section className="question-list-page content-wide page-container-wide">
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
        <QuestionListSkeleton />
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
                            {tagMap.get(tag) ?? tag}
                          </li>
                        ))}
                      </ul>
                    )}
                  </div>
                </Link>
              </li>
            ))}
          </ul>

          <Paginator page={page} totalPages={totalPages} onPageChange={setPage} />
        </>
      )}
    </section>
  );
}
