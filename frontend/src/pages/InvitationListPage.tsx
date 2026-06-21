import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Invitation } from '../types/team';
import { fetchMyInvitations, respondInvitation } from '../utils/invitationApi';
import { useTeam } from '../context/TeamContext';

export default function InvitationListPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { refreshTeams } = useTeam();

  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [statusMessage, setStatusMessage] = useState<{ text: string; isError: boolean } | null>(
    null,
  );
  const [processingId, setProcessingId] = useState<string | null>(null);

  const loadInvitations = useCallback(async () => {
    setIsLoading(true);
    try {
      const all = await fetchMyInvitations();
      const pending = all.filter((inv) => inv.status === 'pending');
      if (pending.length === 0) {
        navigate('/no-team', { replace: true });
        return;
      }
      setInvitations(pending);
    } catch {
      setStatusMessage({ text: t('team.error.fetchFailed'), isError: true });
    } finally {
      setIsLoading(false);
    }
  }, [navigate, t]);

  useEffect(() => {
    void loadInvitations();
  }, [loadInvitations]);

  const handleRespond = async (invitation: Invitation, status: 'accepted' | 'rejected') => {
    setProcessingId(invitation.id);
    setStatusMessage(null);
    try {
      await respondInvitation(invitation.id, status);
      if (status === 'accepted') {
        let remainingCount = 0;
        setInvitations((prev) => {
          const next = prev.filter((inv) => inv.id !== invitation.id);
          remainingCount = next.length;
          return next;
        });
        if (remainingCount === 0) {
          await refreshTeams();
          navigate('/', { replace: true });
        } else {
          setStatusMessage({ text: t('team.invitation.acceptSuccess'), isError: false });
        }
      } else {
        setStatusMessage({ text: t('team.invitation.rejectSuccess'), isError: false });
        setInvitations((prev) => prev.filter((inv) => inv.id !== invitation.id));
        if (invitations.length <= 1) {
          navigate('/no-team', { replace: true });
        }
      }
    } catch {
      const message =
        status === 'accepted'
          ? t('team.invitation.acceptFailed')
          : t('team.invitation.rejectFailed');
      setStatusMessage({ text: message, isError: true });
    } finally {
      setProcessingId(null);
    }
  };

  return (
    <section className="page-container-narrow" aria-labelledby="invitation-page-title">
      <h1 id="invitation-page-title" className="page-title">
        {t('team.invitation.pageTitle')}
      </h1>

      {statusMessage && (
        <p
          role={statusMessage.isError ? 'alert' : 'status'}
          className={`alert ${statusMessage.isError ? 'alert-error' : 'alert-success'}`}
        >
          {statusMessage.text}
        </p>
      )}

      {isLoading ? (
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      ) : invitations.length === 0 ? (
        <p className="team-list-empty">{t('team.invitation.empty')}</p>
      ) : (
        <ul role="list" style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {invitations.map((inv) => (
            <li key={inv.id} className="card" style={{ marginBottom: 0 }}>
              <dl style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '0.5rem 1rem', marginBottom: '1rem' }}>
                <dt className="form-label" style={{ margin: 0 }}>
                  {t('team.invitation.teamId')}
                </dt>
                <dd style={{ margin: 0 }}>{inv.team_id}</dd>

                <dt className="form-label" style={{ margin: 0 }}>
                  {t('team.invitation.invitedBy')}
                </dt>
                <dd style={{ margin: 0 }}>{inv.invited_by}</dd>

                <dt className="form-label" style={{ margin: 0 }}>
                  {t('team.invitation.invitedAt')}
                </dt>
                <dd style={{ margin: 0 }}>
                  {new Date(inv.created_at).toLocaleDateString('ja-JP')}
                </dd>
              </dl>

              <div className="form-actions">
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={() => handleRespond(inv, 'accepted')}
                  disabled={processingId !== null}
                >
                  {t('team.invitation.acceptButton')}
                </button>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => handleRespond(inv, 'rejected')}
                  disabled={processingId !== null}
                >
                  {t('team.invitation.rejectButton')}
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
