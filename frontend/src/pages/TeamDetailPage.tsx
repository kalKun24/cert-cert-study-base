import { useState, useEffect, useCallback } from 'react';
import { Link, useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../context/AuthContext';
import { fetchTeam, deleteTeam, changeMemberRole } from '../utils/teamApi';
import { leaveTeam, sendInvitation } from '../utils/invitationApi';
import { useTeam } from '../context/TeamContext';
import { TeamDetail, TeamMember } from '../types/team';
import MemberRemoveButton from '../components/MemberRemoveButton';
import TeamOwnerRoleModal from '../components/TeamOwnerRoleModal';

interface RoleModalState {
  member: TeamMember;
  action: 'grant' | 'revoke';
}

export default function TeamDetailPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const { refreshTeams } = useTeam();

  const [team, setTeam] = useState<TeamDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [deleteError, setDeleteError] = useState('');
  const [leaveError, setLeaveError] = useState('');
  const [roleChangeError, setRoleChangeError] = useState('');
  const [roleModal, setRoleModal] = useState<RoleModalState | null>(null);
  const [isRoleChanging, setIsRoleChanging] = useState(false);

  // 招待フォーム
  const [inviteeIdentifier, setInviteeIdentifier] = useState('');
  const [isInviting, setIsInviting] = useState(false);
  const [inviteError, setInviteError] = useState('');

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
    user?.role === 'admin' ||
    (team !== null &&
      team.members.some((m) => m.user_id === user?.id && m.role === 'owner'));

  const ownerCount = team ? team.members.filter((m) => m.role === 'owner').length : 0;

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

  const isMember =
    team !== null && user !== null && team.members.some((m) => m.user_id === user.id);

  const handleLeave = async () => {
    if (!team) return;
    if (!window.confirm(t('team.leave.confirm', { name: team.name }))) return;

    setLeaveError('');
    try {
      await leaveTeam(team.id);
      await refreshTeams();
      navigate('/', { replace: true });
    } catch {
      setLeaveError(t('team.leave.failed'));
    }
  };

  const handleRoleChangeConfirm = async () => {
    if (!team || !roleModal) return;
    setIsRoleChanging(true);
    setRoleChangeError('');
    const newRole = roleModal.action === 'grant' ? 'owner' : 'member';
    try {
      await changeMemberRole(team.id, roleModal.member.user_id, newRole);
      setRoleModal(null);
      loadTeam();
    } catch {
      setRoleChangeError(t('team.error.roleChangeFailed'));
      setRoleModal(null);
    } finally {
      setIsRoleChanging(false);
    }
  };

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!team || !inviteeIdentifier.trim()) return;
    setIsInviting(true);
    setInviteError('');
    try {
      await sendInvitation(team.id, inviteeIdentifier.trim());
      setInviteeIdentifier('');
      loadTeam();
    } catch {
      setInviteError(t('team.error.inviteFailed'));
    } finally {
      setIsInviting(false);
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

          {leaveError && (
            <p role="alert" className="alert alert-error">
              {leaveError}
            </p>
          )}

          {roleChangeError && (
            <p role="alert" className="alert alert-error">
              {roleChangeError}
            </p>
          )}

          {isMember && (
            <div style={{ marginBottom: 'var(--space-4)' }}>
              <button
                type="button"
                className="btn btn-secondary"
                onClick={handleLeave}
              >
                {t('team.leave.button')}
              </button>
            </div>
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
                    <th scope="col">{t('user.table.role')}</th>
                    {isOwnerOrAdmin && <th scope="col">{t('user.table.actions')}</th>}
                  </tr>
                </thead>
                <tbody>
                  {team.members.map((member) => {
                    const isCurrentUser = user?.id === member.user_id;
                    const isMemberOwner = member.role === 'owner';
                    const canRevokeOwner = isMemberOwner && ownerCount > 1;
                    const isSoleOwner = isMemberOwner && ownerCount === 1;

                    return (
                      <tr key={member.user_id}>
                        <td data-label={t('team.detail.userId')}>{member.user_id}</td>
                        <td data-label={t('team.detail.joinedAt')}>
                          {new Date(member.joined_at).toLocaleDateString('ja-JP')}
                        </td>
                        <td data-label={t('user.table.role')}>
                          <span
                            className={`member-role-badge member-role-badge--${member.role}`}
                          >
                            {member.role === 'owner'
                              ? t('team.member.role.owner')
                              : t('team.member.role.member')}
                          </span>
                        </td>
                        {isOwnerOrAdmin && (
                          <td>
                            <div className="table-actions">
                              <MemberRemoveButton
                                teamId={team.id}
                                userId={member.user_id}
                                onRemoved={loadTeam}
                              />
                              {!isCurrentUser && (
                                <>
                                  {!isMemberOwner && (
                                    <button
                                      type="button"
                                      className="btn btn-sm btn-secondary"
                                      onClick={() =>
                                        setRoleModal({ member, action: 'grant' })
                                      }
                                      disabled={isRoleChanging}
                                    >
                                      {t('team.member.grantOwner')}
                                    </button>
                                  )}
                                  {isMemberOwner && (
                                    <div className="tooltip-wrapper">
                                      <button
                                        type="button"
                                        className="btn btn-sm btn-secondary"
                                        onClick={() =>
                                          canRevokeOwner
                                            ? setRoleModal({ member, action: 'revoke' })
                                            : undefined
                                        }
                                        disabled={isRoleChanging || isSoleOwner}
                                        title={
                                          isSoleOwner
                                            ? t('team.member.grantOwnerDisabledTooltip')
                                            : undefined
                                        }
                                        aria-describedby={
                                          isSoleOwner
                                            ? `tooltip-sole-owner-${member.user_id}`
                                            : undefined
                                        }
                                      >
                                        {t('team.member.revokeOwner')}
                                      </button>
                                      {isSoleOwner && (
                                        <span
                                          role="tooltip"
                                          id={`tooltip-sole-owner-${member.user_id}`}
                                          className="tooltip"
                                        >
                                          {t('team.member.grantOwnerDisabledTooltip')}
                                        </span>
                                      )}
                                    </div>
                                  )}
                                </>
                              )}
                            </div>
                          </td>
                        )}
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            )}

            {isOwnerOrAdmin && (
              <div className="team-invite-form-wrapper">
                <h3 className="section-title">{t('team.member.inviteTitle')}</h3>
                <form onSubmit={handleInvite} className="team-invite-form">
                  <div className="form-group">
                    <label htmlFor="invitee-identifier">
                      {t('team.member.inviteIdentifierLabel')}
                    </label>
                    <input
                      id="invitee-identifier"
                      type="text"
                      value={inviteeIdentifier}
                      onChange={(e) => setInviteeIdentifier(e.target.value)}
                      placeholder={t('team.member.inviteIdentifierPlaceholder')}
                      className="form-input"
                      required
                    />
                  </div>
                  {inviteError && (
                    <p role="alert" className="alert alert-error">
                      {inviteError}
                    </p>
                  )}
                  <div className="form-actions">
                    <button
                      type="submit"
                      className="btn btn-primary"
                      disabled={isInviting || !inviteeIdentifier.trim()}
                    >
                      {t('team.member.inviteButton')}
                    </button>
                  </div>
                </form>
              </div>
            )}
          </section>
        </>
      ) : null}

      {roleModal && (
        <TeamOwnerRoleModal
          targetUserName={roleModal.member.user_id}
          action={roleModal.action}
          onConfirm={handleRoleChangeConfirm}
          onCancel={() => setRoleModal(null)}
        />
      )}
    </section>
  );
}
