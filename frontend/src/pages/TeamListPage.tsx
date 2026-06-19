import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../context/AuthContext';
import { fetchTeams } from '../utils/teamApi';
import { Team } from '../types/team';

export default function TeamListPage() {
  const { t } = useTranslation();
  const { user } = useAuth();

  const [teams, setTeams] = useState<Team[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');

  useEffect(() => {
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');

    fetchTeams()
      .then((data) => {
        if (isMounted) setTeams(data);
      })
      .catch(() => {
        if (isMounted) setLoadError(t('team.error.fetchFailed'));
      })
      .finally(() => {
        if (isMounted) setIsLoading(false);
      });

    return () => {
      isMounted = false;
    };
  }, [t]);

  const canCreateTeam = user?.role === 'admin' || user?.role === 'teamowner';

  return (
    <section className="team-list-page content-wide">
      <div className="team-list-header">
        <h1 className="page-title">{t('team.list.title')}</h1>
        {canCreateTeam && (
          <Link to="/teams/new" className="btn btn-primary">
            {t('team.list.createButton')}
          </Link>
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
      ) : teams.length === 0 ? (
        <p className="team-list-empty">{t('team.list.empty')}</p>
      ) : (
        <ul className="team-list">
          {teams.map((team) => (
            <li key={team.id} className="team-list-item">
              <Link to={`/teams/${team.id}`} className="team-list-link">
                <span className="team-list-name">{team.name}</span>
                {team.description && (
                  <span className="team-list-description">{team.description}</span>
                )}
              </Link>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
