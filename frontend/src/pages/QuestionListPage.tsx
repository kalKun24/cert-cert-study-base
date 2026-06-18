import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import apiClient from '../utils/apiClient';
import { Question, QuestionListResponse } from '../types/question';

export default function QuestionListPage() {
  const { t } = useTranslation();

  const [questions, setQuestions] = useState<Question[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');

  useEffect(() => {
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');
    apiClient
      .get<QuestionListResponse>('/questions')
      .then((res) => {
        if (isMounted) setQuestions(res.data.data);
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
  }, [t]);

  if (isLoading) {
    return (
      <p role="status" className="page-loading">
        {t('common.loading')}
      </p>
    );
  }

  if (loadError) {
    return (
      <p role="alert" className="alert alert-error">
        {loadError}
      </p>
    );
  }

  return (
    <section className="question-list-page">
      <h1 className="page-title">{t('nav.questions')}</h1>

      {questions.length === 0 ? (
        <p className="question-list-empty">{t('question.list.empty')}</p>
      ) : (
        <ul className="question-list">
          {questions.map((q) => (
            <li key={q.id} className="question-list-item">
              <Link to={`/questions/${q.id}`} className="question-list-link">
                <span className="question-list-title">{q.title}</span>
                <div className="question-list-meta">
                  <span className="question-author">{q.display_name}</span>
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
      )}
    </section>
  );
}
