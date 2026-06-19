import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { User } from '../types/user';
import { updateUserStatus } from '../utils/userApi';

interface UserToggleStatusButtonProps {
  user: User;
  onSuccess: (updated: User) => void;
}

export default function UserToggleStatusButton({
  user,
  onSuccess,
}: UserToggleStatusButtonProps) {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);

  const handleClick = async () => {
    const confirmMsg = user.is_active
      ? t('user.confirm.deactivate', { name: user.display_name })
      : t('user.confirm.activate', { name: user.display_name });

    if (!window.confirm(confirmMsg)) return;

    setIsLoading(true);
    try {
      const updated = await updateUserStatus(user.id, { is_active: !user.is_active });
      onSuccess(updated);
    } catch {
      window.alert(t('user.error.statusUpdateFailed'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <button
      type="button"
      className={`btn btn-sm ${user.is_active ? 'btn-warning' : 'btn-success'}`}
      onClick={handleClick}
      disabled={isLoading}
      aria-label={
        user.is_active
          ? t('user.confirm.deactivate', { name: user.display_name })
          : t('user.confirm.activate', { name: user.display_name })
      }
    >
      {user.is_active ? t('user.action.deactivate') : t('user.action.activate')}
    </button>
  );
}
