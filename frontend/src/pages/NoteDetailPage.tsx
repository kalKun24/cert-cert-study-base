import { useState, useEffect, useMemo } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import ReactMarkdown from 'react-markdown';
import rehypeSanitize from 'rehype-sanitize';
import { fetchNote, deleteNote, updateNoteVisibility } from '../utils/noteApi';
import { fetchTags } from '../utils/tagApi';
import { Tag } from '../types/tag';
import { useAuth } from '../context/AuthContext';
import { useTeam } from '../context/TeamContext';
import { Note, NoteStatus } from '../types/note';
import NoteCommentSection from '../components/NoteCommentSection';
import AccordionSection from '../components/AccordionSection';
import { QuestionDetailSkeleton } from '../components/Skeleton';

const STATUS_CYCLE: Record<NoteStatus, NoteStatus> = {
  draft: 'private',
  private: 'published',
  published: 'draft',
};

function MarkdownContent({ source }: { source: string }) {
  return (
    <div className="markdown-content">
      <ReactMarkdown rehypePlugins={[rehypeSanitize]}>{source}</ReactMarkdown>
    </div>
  );
}

export default function NoteDetailPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { activeTeam } = useTeam();

  const [note, setNote] = useState<Note | null>(null);
  const [tags, setTags] = useState<Tag[]>([]);
  const tagMap = useMemo(() => new Map(tags.map((t) => [t.id, t.name])), [tags]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState('');
  const [isChangingVisibility, setIsChangingVisibility] = useState(false);
  const [visibilityError, setVisibilityError] = useState('');

  useEffect(() => {
    if (!id || !activeTeam) return;
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');
    Promise.all([fetchNote(activeTeam.id, id), fetchTags(activeTeam.id)])
      .then(([n, ts]) => {
        if (isMounted) {
          setNote(n);
          setTags(ts);
        }
      })
      .catch(() => {
        if (isMounted) setLoadError(t('note.error.fetchFailed'));
      })
      .finally(() => {
        if (isMounted) setIsLoading(false);
      });
    return () => {
      isMounted = false;
    };
  }, [id, activeTeam, t]);

  const canEdit =
    user !== null && note !== null && (user.id === note.created_by || user.role === 'admin');

  const handleDelete = async () => {
    if (!note) return;
    const confirmed = window.confirm(t('note.confirm.delete'));
    if (!confirmed) return;

    setIsDeleting(true);
    setDeleteError('');
    try {
      await deleteNote(note.team_id, note.id);
      navigate('/notes');
    } catch {
      setDeleteError(t('note.error.deleteFailed'));
      setIsDeleting(false);
    }
  };

  const handleVisibilityChange = async () => {
    if (!note || !activeTeam) return;
    const nextStatus = STATUS_CYCLE[note.status];
    setIsChangingVisibility(true);
    setVisibilityError('');
    try {
      const updated = await updateNoteVisibility(activeTeam.id, note.id, nextStatus);
      setNote(updated);
    } catch {
      setVisibilityError(t('note.error.visibilityFailed'));
    } finally {
      setIsChangingVisibility(false);
    }
  };

  if (isLoading) {
    return <QuestionDetailSkeleton />;
  }

  if (loadError) {
    return (
      <div className="page-container-narrow">
        <p role="alert" className="alert alert-error">
          {loadError}
        </p>
        <Link to="/notes" className="btn btn-secondary">
          {t('note.backToList')}
        </Link>
      </div>
    );
  }

  if (!note) {
    return (
      <div className="page-container-narrow">
        <p className="alert alert-error">{t('errors.notFound')}</p>
        <Link to="/notes" className="btn btn-secondary">
          {t('note.backToList')}
        </Link>
      </div>
    );
  }

  return (
    <article className="question-detail page-container-narrow">
      <header className="question-detail-header">
        <Link to="/notes" className="question-back-link">
          {t('note.backToList')}
        </Link>
        <div className="question-detail-title-row">
          <h1 className="question-detail-title">{note.title}</h1>
          {canEdit && (
            <div className="question-detail-actions">
              <Link to={`/notes/${note.id}/edit`} className="btn btn-secondary">
                {t('common.edit')}
              </Link>
              <button
                type="button"
                className="btn btn-danger"
                onClick={handleDelete}
                disabled={isDeleting}
                aria-label={`${t('common.delete')} - ${note.title}`}
                aria-busy={isDeleting}
              >
                {t('common.delete')}
              </button>
            </div>
          )}
        </div>
        <div className="question-meta">
          <span className="question-date">
            {new Date(note.created_at).toLocaleDateString('ja-JP')}
          </span>
          <span className="question-status">{t(`note.status.${note.status}`)}</span>
          {note.tags.length > 0 && (
            <ul className="question-tags" aria-label={t('note.tagsLabel')} role="list">
              {note.tags.map((tag) => (
                <li key={tag} className="tag-badge" role="listitem">
                  {tagMap.get(tag) ?? tag}
                </li>
              ))}
            </ul>
          )}
        </div>

        {/* ステータス変更ボタン（作成者または admin のみ） */}
        {canEdit && (
          <div className="question-detail-actions">
            <button
              type="button"
              className="btn btn-secondary btn-sm"
              onClick={handleVisibilityChange}
              disabled={isChangingVisibility}
              aria-busy={isChangingVisibility}
            >
              {t('note.visibility.changeButton')}
            </button>
          </div>
        )}

        {deleteError && (
          <p role="alert" className="alert alert-error">
            {deleteError}
          </p>
        )}
        {visibilityError && (
          <p role="alert" className="alert alert-error">
            {visibilityError}
          </p>
        )}
      </header>

      {/* 本文: デフォルトで開いた状態 */}
      <AccordionSection
        title={t('note.section.body')}
        defaultOpen={true}
        className="question-section--body"
      >
        <MarkdownContent source={note.body} />
      </AccordionSection>

      {/* 議論点: デフォルトで閉じた状態 */}
      <AccordionSection
        title={t('note.section.discussionPoints')}
        defaultOpen={false}
        className="question-section--answer"
      >
        <MarkdownContent source={note.discussion_points} />
      </AccordionSection>

      {/* メモ: 内容がある場合のみ、デフォルトで閉じた状態 */}
      {note.memo && (
        <AccordionSection
          title={t('note.section.memo')}
          defaultOpen={false}
          className="question-section--memo"
        >
          <MarkdownContent source={note.memo} />
        </AccordionSection>
      )}

      <NoteCommentSection teamId={note.team_id} noteId={note.id} />
    </article>
  );
}
