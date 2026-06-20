import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { fetchQuestions } from '../utils/questionApi';
import { fetchTags } from '../utils/tagApi';
import { fetchTeams } from '../utils/teamApi';
import { useTeam } from '../context/TeamContext';
import { Question } from '../types/question';
import { DashboardSkeleton } from '../components/Skeleton';

interface DashboardStats {
  questions: number;
  tags: number;
  teams: number;
}

export default function HomePage() {
  const { t } = useTranslation();
  const { activeTeam } = useTeam();

  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [recentQuestions, setRecentQuestions] = useState<Question[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let isMounted = true;

    Promise.allSettled([
      activeTeam ? fetchQuestions(activeTeam.id, { page: 1, per_page: 5 }) : Promise.resolve(null),
      activeTeam ? fetchTags(activeTeam.id) : Promise.resolve([]),
      fetchTeams(),
    ]).then(([questionsResult, tagsResult, teamsResult]) => {
      if (!isMounted) return;

      setStats({
        questions: questionsResult.status === 'fulfilled' && questionsResult.value !== null ? questionsResult.value.total : 0,
        tags: tagsResult.status === 'fulfilled' ? tagsResult.value.length : 0,
        teams: teamsResult.status === 'fulfilled' ? teamsResult.value.length : 0,
      });

      if (questionsResult.status === 'fulfilled' && questionsResult.value !== null) {
        setRecentQuestions(questionsResult.value.items.slice(0, 5));
      }

      setIsLoading(false);
    });

    return () => {
      isMounted = false;
    };
  }, [activeTeam?.id]);

  if (isLoading) {
    return <DashboardSkeleton />;
  }

  return (
    <div className="dashboard content-wide page-container-wide">
      <div className="dashboard-header">
        <h1 className="page-title">{t('home.title')}</h1>
        <p className="dashboard-tagline">{t('app.tagline')}</p>
      </div>

      {/* 統計カード */}
      {stats !== null && (
        <div className="dashboard-stats" aria-label="統計情報">
          <div className="stat-card">
            <span className="stat-label">{t('home.stats.questions')}</span>
            <span className="stat-value">{stats.questions}</span>
          </div>
          <div className="stat-card">
            <span className="stat-label">{t('home.stats.tags')}</span>
            <span className="stat-value">{stats.tags}</span>
          </div>
          <div className="stat-card">
            <span className="stat-label">{t('home.stats.teams')}</span>
            <span className="stat-value">{stats.teams}</span>
          </div>
        </div>
      )}

      {/* 最近の問題 */}
      <section className="dashboard-recent">
        <div className="dashboard-section-header">
          <h2 className="section-title">{t('home.recentQuestions.title')}</h2>
          <Link to="/questions" className="btn btn-secondary btn-sm">
            {t('home.recentQuestions.viewAll')}
          </Link>
        </div>

        {recentQuestions.length === 0 ? (
          <p className="question-list-empty">{t('home.recentQuestions.empty')}</p>
        ) : (
          <ul className="question-list">
            {recentQuestions.map((q) => (
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
        )}
      </section>
    </div>
  );
}
