import { useState, useEffect, useMemo } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { fetchQuestion, deleteQuestion } from '../utils/questionApi';
import { fetchTags } from '../utils/tagApi';
import { Tag } from '../types/tag';
import { useAuth } from '../context/AuthContext';
import { useTeam } from '../context/TeamContext';
import { Question } from '../types/question';
import CommentSection from '../components/CommentSection';
import AccordionSection from '../components/AccordionSection';
import { QuestionDetailSkeleton } from '../components/Skeleton';
import MarkdownPreviewContent from '../components/MarkdownPreviewContent';
import 'highlight.js/styles/github-dark-dimmed.css';
import 'katex/dist/katex.min.css';

function MarkdownContent({ source }: { source: string }) {
  return (
    <div className="markdown-content">
      <MarkdownPreviewContent value={source} />
    </div>
  );
}

export default function QuestionDetailPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const { activeTeam } = useTeam();

  const [question, setQuestion] = useState<Question | null>(null);
  const [tags, setTags] = useState<Tag[]>([]);
  const tagMap = useMemo(() => new Map(tags.map((t) => [t.id, t.name])), [tags]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState('');

  useEffect(() => {
    if (!id || !activeTeam) return;
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');
    Promise.all([fetchQuestion(activeTeam.id, id), fetchTags(activeTeam.id)])
      .then(([q, ts]) => {
        if (isMounted) {
          setQuestion(q);
          setTags(ts);
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
  }, [id, activeTeam, t]);

  const canEdit = user !== null && question !== null &&
    (user.id === question.created_by || user.role === 'admin');

  const handleDelete = async () => {
    if (!question) return;
    const confirmed = window.confirm(t('question.confirm.delete'));
    if (!confirmed) return;

    setIsDeleting(true);
    setDeleteError('');
    try {
      await deleteQuestion(question.team_id, question.id);
      navigate('/questions');
    } catch {
      setDeleteError(t('question.error.deleteFailed'));
      setIsDeleting(false);
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
        <Link to="/questions" className="btn btn-secondary">
          {t('question.backToList')}
        </Link>
      </div>
    );
  }

  if (!question) {
    return (
      <div className="page-container-narrow">
        <p className="alert alert-error">{t('errors.notFound')}</p>
        <Link to="/questions" className="btn btn-secondary">
          {t('question.backToList')}
        </Link>
      </div>
    );
  }

  return (
    <article className="question-detail page-container-narrow">
      <header className="question-detail-header">
        <Link to="/questions" className="question-back-link">
          {t('question.backToList')}
        </Link>
        <div className="question-detail-title-row">
          <h1 className="question-detail-title">{question.title}</h1>
          {canEdit && (
            <div className="question-detail-actions">
              <Link
                to={`/questions/${question.id}/edit`}
                className="btn btn-secondary"
              >
                {t('common.edit')}
              </Link>
              <button
                type="button"
                className="btn btn-danger"
                onClick={handleDelete}
                disabled={isDeleting}
                aria-label={`${t('common.delete')} - ${question.title}`}
                aria-busy={isDeleting}
              >
                {t('common.delete')}
              </button>
            </div>
          )}
        </div>
        <div className="question-meta">
          <span className="question-date">
            {new Date(question.created_at).toLocaleDateString('ja-JP')}
          </span>
          <span className="question-status">{t(`question.status.${question.status}`)}</span>
          {question.tags.length > 0 && (
            <ul className="question-tags" aria-label={t('question.tagsLabel')} role="list">
              {question.tags.map((tag) => (
                <li key={tag} className="tag-badge" role="listitem">
                  {tagMap.get(tag) ?? tag}
                </li>
              ))}
            </ul>
          )}
        </div>
        {deleteError && (
          <p role="alert" className="alert alert-error">
            {deleteError}
          </p>
        )}
      </header>

      {/* 問題文: デフォルトで開いた状態 */}
      <AccordionSection
        title={t('question.section.body')}
        defaultOpen={true}
        className="question-section--body"
      >
        <MarkdownContent source={question.body} />
      </AccordionSection>

      {/* 解答: デフォルトで閉じた状態 */}
      <AccordionSection
        title={t('question.section.answer')}
        defaultOpen={false}
        className="question-section--answer"
      >
        <MarkdownContent source={question.answer} />
      </AccordionSection>

      {/* 解説: デフォルトで閉じた状態 */}
      <AccordionSection
        title={t('question.section.explanation')}
        defaultOpen={false}
        className="question-section--explanation"
      >
        <MarkdownContent source={question.explanation} />
      </AccordionSection>

      {/* 議論点・メモ: 内容がある場合のみ、デフォルトで閉じた状態 */}
      {question.memo && (
        <AccordionSection
          title={t('question.section.discussionNotes')}
          defaultOpen={false}
          className="question-section--memo"
        >
          <MarkdownContent source={question.memo} />
        </AccordionSection>
      )}

      <CommentSection teamId={question.team_id} questionId={question.id} />
    </article>
  );
}
