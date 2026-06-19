import { useState, useEffect, useCallback } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../context/AuthContext';
import { fetchTeam, deleteTeam } from '../utils/teamApi';
import { TeamDetail } from '../types/team';
import MemberInvite from '../components/MemberInvite';
import MemberRemoveButton from '../components/MemberRemoveButton';

export default function TeamDetailPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [team, setTeam] = useState<TeamDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [deleteError, setDeleteError] = useState('');

  const loadTeam = useCallback(() => {
    if (!id) return;
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');

    fetchTeam(id)
      .then((data) => {
        if (isMounted) setTeam(data);
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
  }, [id, t]);

  useEffect(() => {
    return loadTeam();
  }, [loadTeam]);

  if (!id) return null;

  const isOwnerOrAdmin =
    user?.role === 'admin' || (team !== null && user?.id === team.owner_id);

  const handleDelete = async () => {
    if (!team) return;
    if (!window.confirm(t('team.detail.deleteConfirm', { name: team.name }))) return;

    setDeleteError('');
    try {
      await deleteTeam(team.id);
      navigate('/teams', { replace: true });
    } catch {
      setDeleteError(t('team.error.deleteFailed'));
    }
  };

  return (
    <section className="team-detail-page page-container-full">
      <Link to="/teams" className="back-link">
        {t('team.detail.backToList')}
      </Link>

      {isLoading ? (
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      ) : loadError ? (
        <p role="alert" className="alert alert-error">
          {loadError}
        </p>
      ) : team ? (
        <>
          <div className="team-detail-header">
            <h1 className="page-title">{team.name}</h1>
            {isOwnerOrAdmin && (
              <div className="team-detail-actions">
                <Link to={`/teams/${team.id}/edit`} className="btn btn-secondary">
                  {t('team.detail.editButton')}
                </Link>
                <button
                  type="button"
                  className="btn btn-danger"
                  onClick={handleDelete}
                >
                  {t('team.detail.deleteButton')}
                </button>
              </div>
            )}
          </div>

          {team.description && (
            <p className="team-detail-description">{team.description}</p>
          )}

          {deleteError && (
            <p role="alert" className="alert alert-error">
              {deleteError}
            </p>
          )}

          <section className="team-members-section" aria-labelledby="members-heading">
            <h2 id="members-heading" className="section-title">
              {t('team.detail.membersTitle')}
            </h2>

            {team.members.length === 0 ? (
              <p className="team-members-empty">{t('team.detail.membersEmpty')}</p>
            ) : (
              <table className="team-members-table">
                <thead>
                  <tr>
                    <th scope="col">{t('team.detail.userId')}</th>
                    <th scope="col">{t('team.detail.joinedAt')}</th>
                    {isOwnerOrAdmin && <th scope="col" />}
                  </tr>
                </thead>
                <tbody>
                  {team.members.map((member) => (
                    <tr key={member.user_id}>
                      <td data-label={t('team.detail.userId')}>{member.user_id}</td>
                      <td data-label={t('team.detail.joinedAt')}>
                        {new Date(member.joined_at).toLocaleDateString('ja-JP')}
                      </td>
                      {isOwnerOrAdmin && (
                        <td>
                          <MemberRemoveButton
                            teamId={team.id}
                            userId={member.user_id}
                            onRemoved={loadTeam}
                          />
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            )}

            {isOwnerOrAdmin && (
              <MemberInvite teamId={team.id} onInvited={loadTeam} />
            )}
          </section>
        </>
      ) : null}
    </section>
  );
}
