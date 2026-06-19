import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { User } from '../types/user';
import UserStatusBadge from './UserStatusBadge';
import UserToggleStatusButton from './UserToggleStatusButton';
import UserDeleteButton from './UserDeleteButton';

interface UserTableProps {
  users: User[];
  onStatusToggled: (updated: User) => void;
  onDeleted: (deletedId: string) => void;
}

export default function UserTable({ users, onStatusToggled, onDeleted }: UserTableProps) {
  const { t } = useTranslation();

  return (
    <div className="table-wrapper">
      <table className="user-table">
        <thead>
          <tr>
            <th scope="col">{t('user.table.username')}</th>
            <th scope="col">{t('user.table.displayName')}</th>
            <th scope="col">{t('user.table.role')}</th>
            <th scope="col">{t('user.table.status')}</th>
            <th scope="col">{t('user.table.createdAt')}</th>
            <th scope="col">{t('user.table.actions')}</th>
          </tr>
        </thead>
        <tbody>
          {users.map((user) => (
            <tr key={user.id}>
              <td data-label={t('user.table.username')}>{user.username}</td>
              <td data-label={t('user.table.displayName')}>{user.display_name}</td>
              <td data-label={t('user.table.role')}>{t(`user.role.${user.role}`)}</td>
              <td data-label={t('user.table.status')}>
                <UserStatusBadge isActive={user.is_active} />
              </td>
              <td data-label={t('user.table.createdAt')}>
                {new Date(user.created_at).toLocaleDateString('ja-JP')}
              </td>
              <td data-label={t('user.table.actions')}>
                <div className="table-actions">
                  <Link
                    to={`/admin/users/${user.id}/edit`}
                    className="btn btn-secondary btn-sm"
                  >
                    {t('common.edit')}
                  </Link>
                  <UserToggleStatusButton user={user} onSuccess={onStatusToggled} />
                  <UserDeleteButton user={user} onSuccess={onDeleted} />
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
