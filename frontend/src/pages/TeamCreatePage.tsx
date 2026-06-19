import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { createTeam } from '../utils/teamApi';
import TeamForm from '../components/TeamForm';

export default function TeamCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');

  const handleSubmit = async (name: string, description: string) => {
    setIsSubmitting(true);
    setSubmitError('');
    try {
      const created = await createTeam({ name, description });
      navigate(`/teams/${created.id}`);
    } catch {
      setSubmitError(t('team.error.createFailed'));
      setIsSubmitting(false);
    }
  };

  return (
    <section className="team-form-page">
      <h1 className="page-title">{t('team.form.createTitle')}</h1>
      <TeamForm
        isSubmitting={isSubmitting}
        submitError={submitError}
        onSubmit={handleSubmit}
        onCancel={() => navigate('/teams')}
      />
    </section>
  );
}
