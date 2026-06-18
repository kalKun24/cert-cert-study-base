import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import apiClient from '../utils/apiClient';
import { Question, QuestionResponse } from '../types/question';
import CommentSection from '../components/CommentSection';

export default function QuestionDetailPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();

  const [question, setQuestion] = useState<Question | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');

  useEffect(() => {
    if (!id) return;
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');
    apiClient
      .get<QuestionResponse>(`/questions/${id}`)
      .then((res) => {
        if (isMounted) setQuestion(res.data.data);
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
        <h1 className="question-detail-title">{question.title}</h1>
        <div className="question-meta">
          <span className="question-author">{question.display_name}</span>
          <span className="question-date">
            {new Date(question.created_at).toLocaleDateString('ja-JP')}
          </span>
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
      </header>

      <section className="question-detail-section" aria-label={t('question.section.body')}>
        <h2 className="question-section-heading">{t('question.section.body')}</h2>
        <div className="question-content">{question.body}</div>
      </section>

      <section className="question-detail-section" aria-label={t('question.section.answer')}>
        <h2 className="question-section-heading">{t('question.section.answer')}</h2>
        <div className="question-content">{question.answer}</div>
      </section>

      <section className="question-detail-section" aria-label={t('question.section.explanation')}>
        <h2 className="question-section-heading">{t('question.section.explanation')}</h2>
        <div className="question-content">{question.explanation}</div>
      </section>

      {question.discussion_notes && (
        <section
          className="question-detail-section"
          aria-label={t('question.section.discussionNotes')}
        >
          <h2 className="question-section-heading">{t('question.section.discussionNotes')}</h2>
          <div className="question-content">{question.discussion_notes}</div>
        </section>
      )}

      <CommentSection questionId={question.id} />
    </article>
  );
}
