import { useTranslation } from 'react-i18next';
import { useTeam } from '../context/TeamContext';

export default function NoTeamPage() {
  const { t } = useTranslation();
  const { refreshTeams, isLoading } = useTeam();

  const handleRefresh = async () => {
    await refreshTeams();
  };

  return (
    <section className="page-container-narrow" aria-labelledby="no-team-title">
      <div className="card" style={{ textAlign: 'center' }}>
        <h1 id="no-team-title" className="page-title">
          {t('team.noTeam.title')}
        </h1>
        <p>{t('team.noTeam.description')}</p>
        <div className="form-actions" style={{ justifyContent: 'center' }}>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleRefresh}
            disabled={isLoading}
          >
            {t('team.noTeam.refreshButton')}
          </button>
        </div>
      </div>
    </section>
  );
}
