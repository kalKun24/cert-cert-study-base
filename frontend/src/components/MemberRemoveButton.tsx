import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { removeMember } from '../utils/teamApi';

interface MemberRemoveButtonProps {
  teamId: string;
  userId: string;
  onRemoved: () => void;
}

export default function MemberRemoveButton({
  teamId,
  userId,
  onRemoved,
}: MemberRemoveButtonProps) {
  const { t } = useTranslation();
  const [isRemoving, setIsRemoving] = useState(false);
  const [removeError, setRemoveError] = useState('');

  const handleRemove = async () => {
    if (!window.confirm(t('team.member.removeConfirm'))) return;

    setRemoveError('');
    setIsRemoving(true);
    try {
      await removeMember(teamId, userId);
      onRemoved();
    } catch {
      setRemoveError(t('team.error.removeFailed'));
      setIsRemoving(false);
    }
  };

  return (
    <>
      <button
        type="button"
        className="btn btn-danger btn-sm"
        onClick={handleRemove}
        disabled={isRemoving}
        aria-label={t('team.member.removeButton')}
      >
        {t('team.member.removeButton')}
      </button>
      {removeError && (
        <p role="alert" className="form-error">
          {removeError}
        </p>
      )}
    </>
  );
}
