import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { User } from '../types/user';
import { deleteUser } from '../utils/userApi';

interface UserDeleteButtonProps {
  user: User;
  onSuccess: (deletedId: string) => void;
}

export default function UserDeleteButton({ user, onSuccess }: UserDeleteButtonProps) {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);

  const handleClick = async () => {
    const confirmed = window.confirm(
      t('user.confirm.delete', { name: user.display_name }),
    );
    if (!confirmed) return;

    setIsLoading(true);
    try {
      await deleteUser(user.id);
      onSuccess(user.id);
    } catch {
      window.alert(t('user.error.deleteFailed'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <button
      type="button"
      className="btn btn-danger btn-sm"
      onClick={handleClick}
      disabled={isLoading}
      aria-label={t('user.confirm.delete', { name: user.display_name })}
    >
      {t('common.delete')}
    </button>
  );
}
