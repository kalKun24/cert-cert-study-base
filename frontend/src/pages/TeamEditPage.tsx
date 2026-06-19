import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../context/AuthContext';
import { fetchTeam, updateTeam } from '../utils/teamApi';
import { TeamDetail } from '../types/team';
import TeamForm from '../components/TeamForm';

export default function TeamEditPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [team, setTeam] = useState<TeamDetail | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');

  useEffect(() => {
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

  if (!id) return null;

  if (isLoading) {
    return (
      <section className="team-form-page page-container-full">
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      </section>
    );
  }

  if (loadError) {
    return (
      <section className="team-form-page page-container-full">
        <p role="alert" className="alert alert-error">
          {loadError}
        </p>
      </section>
    );
  }

  if (!team) return null;

  const isOwnerOrAdmin = user?.role === 'admin' || user?.id === team.owner_id;

  if (!isOwnerOrAdmin) {
    return (
      <section className="team-form-page page-container-full">
        <p role="alert" className="alert alert-error">
          {t('errors.forbidden')}
        </p>
      </section>
    );
  }

  const handleSubmit = async (name: string, description: string) => {
    setIsSubmitting(true);
    setSubmitError('');
    try {
      await updateTeam(team.id, { name, description });
      navigate(`/teams/${team.id}`);
    } catch {
      setSubmitError(t('team.error.updateFailed'));
      setIsSubmitting(false);
    }
  };

  return (
    <section className="team-form-page page-container-full">
      <h1 className="page-title">{t('team.form.editTitle')}</h1>
      <TeamForm
        initialName={team.name}
        initialDescription={team.description}
        isSubmitting={isSubmitting}
        submitError={submitError}
        onSubmit={handleSubmit}
        onCancel={() => navigate(`/teams/${team.id}`)}
      />
    </section>
  );
}
