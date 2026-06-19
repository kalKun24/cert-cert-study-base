import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import ReactMarkdown from 'react-markdown';
import rehypeSanitize from 'rehype-sanitize';
import { fetchQuestion, deleteQuestion } from '../utils/questionApi';
import { useAuth } from '../context/AuthContext';
import { Question } from '../types/question';
import CommentSection from '../components/CommentSection';

function MarkdownContent({ source }: { source: string }) {
  return (
    <div className="markdown-content">
      <ReactMarkdown rehypePlugins={[rehypeSanitize]}>{source}</ReactMarkdown>
    </div>
  );
}

export default function QuestionDetailPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [question, setQuestion] = useState<Question | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState('');

  useEffect(() => {
    if (!id) return;
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');
    fetchQuestion(id)
      .then((q) => {
        if (isMounted) setQuestion(q);
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
  }, [id, t]);

  const canEdit = user !== null && question !== null &&
    (user.id === question.created_by || user.role === 'admin');

  const handleDelete = async () => {
    if (!question) return;
    const confirmed = window.confirm(t('question.confirm.delete'));
    if (!confirmed) return;

    setIsDeleting(true);
    setDeleteError('');
    try {
      await deleteQuestion(question.id);
      navigate('/questions');
    } catch {
      setDeleteError(t('question.error.deleteFailed'));
      setIsDeleting(false);
    }
  };

  if (isLoading) {
    return (
      <p role="status" className="page-loading">
        {t('common.loading')}
      </p>
    );
  }

  if (loadError) {
    return (
      <div>
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
      <div>
        <p className="alert alert-error">{t('errors.notFound')}</p>
        <Link to="/questions" className="btn btn-secondary">
          {t('question.backToList')}
        </Link>
      </div>
    );
  }

  return (
    <article className="question-detail">
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
                aria-label={t('common.delete')}
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
            <ul className="question-tags" aria-label={t('question.tagsLabel')}>
              {question.tags.map((tag) => (
                <li key={tag} className="tag-badge">
                  {tag}
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

      <section className="question-detail-section" aria-label={t('question.section.body')}>
        <h2 className="question-section-heading">{t('question.section.body')}</h2>
        <MarkdownContent source={question.body} />
      </section>

      <section className="question-detail-section" aria-label={t('question.section.answer')}>
        <h2 className="question-section-heading">{t('question.section.answer')}</h2>
        <MarkdownContent source={question.answer} />
      </section>

      <section className="question-detail-section" aria-label={t('question.section.explanation')}>
        <h2 className="question-section-heading">{t('question.section.explanation')}</h2>
        <MarkdownContent source={question.explanation} />
      </section>

      {question.memo && (
        <section
          className="question-detail-section"
          aria-label={t('question.section.discussionNotes')}
        >
          <h2 className="question-section-heading">{t('question.section.discussionNotes')}</h2>
          <MarkdownContent source={question.memo} />
        </section>
      )}

      <CommentSection questionId={question.id} />
    </article>
  );
}
