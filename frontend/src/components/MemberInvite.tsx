import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { addMember } from '../utils/teamApi';

interface MemberInviteProps {
  teamId: string;
  onInvited: () => void;
}

export default function MemberInvite({ teamId, onInvited }: MemberInviteProps) {
  const { t } = useTranslation();

  const [userId, setUserId] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');
  const [validationError, setValidationError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!userId.trim()) {
      setValidationError(t('team.validation.userIdRequired'));
      return;
    }
    setValidationError('');
    setSubmitError('');
    setIsSubmitting(true);

    try {
      await addMember(teamId, { user_id: userId.trim() });
      setUserId('');
      onInvited();
    } catch {
      setSubmitError(t('team.error.inviteFailed'));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} noValidate className="member-invite-form">
      <h3 className="member-invite-title">{t('team.member.inviteTitle')}</h3>

      <div className="form-field member-invite-field">
        <label htmlFor="member-user-id" className="form-label">
          {t('team.member.userIdLabel')}
        </label>
        <div className="member-invite-input-row">
          <input
            id="member-user-id"
            type="text"
            className="form-input"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
            placeholder={t('team.member.userIdPlaceholder')}
            aria-describedby={validationError ? 'member-user-id-error' : undefined}
            aria-invalid={!!validationError}
            disabled={isSubmitting}
          />
          <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
            {t('team.member.inviteButton')}
          </button>
        </div>
        {validationError && (
          <p id="member-user-id-error" role="alert" className="form-error">
            {validationError}
          </p>
        )}
      </div>

      {submitError && (
        <p role="alert" className="alert alert-error">
          {submitError}
        </p>
      )}
    </form>
  );
}
