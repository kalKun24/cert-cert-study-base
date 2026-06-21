import { useState, useEffect, useCallback } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { fetchNotes } from '../utils/noteApi';
import { fetchTags } from '../utils/tagApi';
import { useTeam } from '../context/TeamContext';
import { Note } from '../types/note';
import { Tag } from '../types/tag';
import TagChip from '../components/TagChip';
import Paginator from '../components/Paginator';
import { QuestionListSkeleton } from '../components/Skeleton';

const PER_PAGE = 20;

const PARAM_TAG_IDS = 'tag_ids';
const PARAM_KEYWORD = 'keyword';
const PARAM_PAGE = 'page';

export default function NoteListPage() {
  const { t } = useTranslation();
  const { activeTeam } = useTeam();
  const [searchParams, setSearchParams] = useSearchParams();

  const initialTagNames = (): string[] => {
    const raw = searchParams.get(PARAM_TAG_IDS);
    if (!raw) return [];
    return raw.split(',').filter(Boolean);
  };
  const initialKeyword = searchParams.get(PARAM_KEYWORD) ?? '';
  const initialPage = parseInt(searchParams.get(PARAM_PAGE) ?? '1', 10) || 1;

  const [notes, setNotes] = useState<Note[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [keyword, setKeyword] = useState(initialKeyword);
  const [selectedTagNames, setSelectedTagNames] = useState<string[]>(initialTagNames);
  const [page, setPage] = useState(initialPage);
  const [totalPages, setTotalPages] = useState(1);

  const syncSearchParams = useCallback(
    (currentPage: number, currentKeyword: string, currentTagNames: string[]) => {
      const next = new URLSearchParams();
      if (currentKeyword.trim()) next.set(PARAM_KEYWORD, currentKeyword.trim());
      if (currentTagNames.length > 0) next.set(PARAM_TAG_IDS, currentTagNames.join(','));
      if (currentPage > 1) next.set(PARAM_PAGE, String(currentPage));
      setSearchParams(next, { replace: true });
    },
    [setSearchParams]
  );

  const loadNotes = useCallback(
    (currentPage: number, currentKeyword: string, currentTagNames: string[]) => {
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
      if (currentTagNames.length > 0) {
        params.tag_ids = currentTagNames.join(',');
      }

      fetchNotes(activeTeam.id, params)
        .then((data) => {
          if (isMounted) {
            setNotes(data.items);
            setTotalPages(data.total_pages);
          }
        })
        .catch(() => {
          if (isMounted) {
            setLoadError(t('note.error.fetchFailed'));
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
        // タグ取得エラーは非致命的
      });
  }, [activeTeam?.id]);

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [page]);

  useEffect(() => {
    syncSearchParams(page, keyword, selectedTagNames);
    const cleanup = loadNotes(page, keyword, selectedTagNames);
    return cleanup;
  }, [page, keyword, selectedTagNames, loadNotes, syncSearchParams]);

  const handleTagToggle = (tagName: string) => {
    setSelectedTagNames((prev) =>
      prev.includes(tagName) ? prev.filter((n) => n !== tagName) : [...prev, tagName]
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
        <h1 className="page-title">{t('nav.notes')}</h1>
        <Link to="/notes/new" className="btn btn-primary">
          {t('note.new')}
        </Link>
      </div>

      <div className="question-list-filters">
        <input
          type="search"
          className="filter-keyword-input"
          placeholder={t('note.filter.keywordPlaceholder')}
          value={keyword}
          onChange={handleKeywordChange}
          aria-label={t('note.filter.keywordPlaceholder')}
        />

        {tags.length > 0 && (
          <div
            className="filter-tags"
            role="group"
            aria-label={t('note.filter.tagPlaceholder')}
          >
            {tags.map((tag) => (
              <TagChip
                key={tag.id}
                tag={tag}
                selected={selectedTagNames.includes(tag.name)}
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
      ) : notes.length === 0 ? (
        <p className="question-list-empty">{t('note.list.empty')}</p>
      ) : (
        <>
          <ul className="question-list">
            {notes.map((note) => (
              <li key={note.id} className="question-list-item">
                <Link to={`/notes/${note.id}`} className="question-list-link">
                  <span className="question-list-title">{note.title}</span>
                  <div className="question-list-meta">
                    <span className="question-date">
                      {new Date(note.created_at).toLocaleDateString('ja-JP')}
                    </span>
                    <span className="question-status">
                      {t(`note.status.${note.status}`)}
                    </span>
                    {note.tags.length > 0 && (
                      <ul className="question-tags" aria-label={t('note.tagsLabel')}>
                        {note.tags.map((tag) => (
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

          <Paginator page={page} totalPages={totalPages} onPageChange={setPage} />
        </>
      )}
    </section>
  );
}
