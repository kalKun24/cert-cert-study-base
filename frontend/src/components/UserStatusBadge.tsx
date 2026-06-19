import { useTranslation } from 'react-i18next';

interface UserStatusBadgeProps {
  isActive: boolean;
}

export default function UserStatusBadge({ isActive }: UserStatusBadgeProps) {
  const { t } = useTranslation();

  return (
    <span className={`status-badge ${isActive ? 'status-badge-active' : 'status-badge-inactive'}`}>
      {isActive ? t('user.status.active') : t('user.status.inactive')}
    </span>
  );
}
