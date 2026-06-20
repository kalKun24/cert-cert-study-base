import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { User, UpdateUserRequest, CreateUserRequest } from '../types/user';
import { fetchUser, updateUser, updateTeamOwnerStatus } from '../utils/userApi';
import { useAuth } from '../context/AuthContext';
import UserForm from '../components/UserForm';

export default function UserEditPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user: currentUser } = useAuth();

  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');

  // チームオーナー権限フィールド
  const [isTeamOwner, setIsTeamOwner] = useState(false);
  const [maxTeams, setMaxTeams] = useState<number>(1);
  const [isTeamOwnerSaving, setIsTeamOwnerSaving] = useState(false);
  const [teamOwnerSaveError, setTeamOwnerSaveError] = useState('');
  const [teamOwnerSaveSuccess, setTeamOwnerSaveSuccess] = useState('');

  useEffect(() => {
    if (!id) return;
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');

    fetchUser(id)
      .then((data) => {
        if (isMounted) {
          setUser(data);
          setIsTeamOwner(data.is_team_owner ?? false);
          setMaxTeams(data.max_teams ?? 1);
        }
      })
      .catch(() => {
        if (isMounted) setLoadError(t('user.error.fetchFailed'));
      })
      .finally(() => {
        if (isMounted) setIsLoading(false);
      });

    return () => {
      isMounted = false;
    };
  }, [id, t]);

  if (!id) return null;

  if (isLoading) {
    return (
      <section className="user-form-page page-container-full">
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      </section>
    );
  }

  if (loadError) {
    return (
      <section className="user-form-page page-container-full">
        <p role="alert" className="alert alert-error">
          {loadError}
        </p>
      </section>
    );
  }

  if (!user) return null;

  const isAdmin = currentUser?.role === 'admin';

  const handleSubmit = async (data: CreateUserRequest | UpdateUserRequest) => {
    setIsSubmitting(true);
    setSubmitError('');
    try {
      await updateUser(user.id, data as UpdateUserRequest);
      navigate('/admin/users');
    } catch {
      setSubmitError(t('user.error.updateFailed'));
      setIsSubmitting(false);
    }
  };

  const handleTeamOwnerSave = async () => {
    setIsTeamOwnerSaving(true);
    setTeamOwnerSaveError('');
    setTeamOwnerSaveSuccess('');
    try {
      await updateTeamOwnerStatus(user.id, {
        is_team_owner: isTeamOwner,
        max_teams: isTeamOwner ? maxTeams : undefined,
      });
      setTeamOwnerSaveSuccess(t('user.teamOwner.saveSuccess'));
    } catch {
      setTeamOwnerSaveError(t('user.teamOwner.saveFailed'));
    } finally {
      setIsTeamOwnerSaving(false);
    }
  };

  return (
    <section className="user-form-page page-container-full">
      <h1 className="page-title">{t('user.form.editTitle')}</h1>
      <UserForm
        mode="edit"
        initialUsername={user.username}
        initialDisplayName={user.display_name}
        initialEmail={user.email}
        initialRole={user.role}
        isSubmitting={isSubmitting}
        submitError={submitError}
        onSubmit={handleSubmit}
        onCancel={() => navigate('/admin/users')}
      />

      {isAdmin && (
        <div className="team-owner-section">
          <h2 className="section-title">{t('user.teamOwner.label')}</h2>

          <div className="form-group">
            <div className="team-owner-toggle-row">
              <label className="team-owner-toggle-label" htmlFor="team-owner-toggle">
                {t('user.teamOwner.enabledLabel')}
              </label>
              <button
                id="team-owner-toggle"
                type="button"
                role="switch"
                aria-checked={isTeamOwner}
                className={`toggle-switch${isTeamOwner ? ' toggle-switch--on' : ''}`}
                onClick={() => {
                  setIsTeamOwner((prev) => !prev);
                  setTeamOwnerSaveSuccess('');
                }}
              >
                <span className="toggle-switch-thumb" />
              </button>
            </div>
          </div>

          {isTeamOwner && (
            <div className="form-group">
              <label htmlFor="max-teams-input">{t('user.teamOwner.maxTeamsLabel')}</label>
              <input
                id="max-teams-input"
                type="number"
                min={1}
                value={maxTeams}
                placeholder={t('user.teamOwner.maxTeamsPlaceholder')}
                onChange={(e) => setMaxTeams(Number(e.target.value))}
                className="form-input"
                style={{ maxWidth: '160px' }}
              />
            </div>
          )}

          {teamOwnerSaveError && (
            <p role="alert" className="alert alert-error">
              {teamOwnerSaveError}
            </p>
          )}
          {teamOwnerSaveSuccess && (
            <p role="status" className="alert alert-success">
              {teamOwnerSaveSuccess}
            </p>
          )}

          <div className="form-actions">
            <button
              type="button"
              className="btn btn-primary"
              onClick={handleTeamOwnerSave}
              disabled={isTeamOwnerSaving}
            >
              {t('user.teamOwner.saveButton')}
            </button>
          </div>
        </div>
      )}
    </section>
  );
}
