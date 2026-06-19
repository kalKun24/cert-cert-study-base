import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { User } from '../types/user';
import UserStatusBadge from './UserStatusBadge';

interface UserTableProps {
  users: User[];
  onStatusToggle: (user: User) => void;
  onDelete: (user: User) => void;
}

export default function UserTable({ users, onStatusToggle, onDelete }: UserTableProps) {
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
              <td>{user.username}</td>
              <td>{user.display_name}</td>
              <td>{t(`user.role.${user.role}`)}</td>
              <td>
                <UserStatusBadge isActive={user.is_active} />
              </td>
              <td>{new Date(user.created_at).toLocaleDateString('ja-JP')}</td>
              <td>
                <div className="table-actions">
                  <Link
                    to={`/admin/users/${user.id}/edit`}
                    className="btn btn-secondary btn-sm"
                  >
                    {t('common.edit')}
                  </Link>
                  <button
                    type="button"
                    className={`btn btn-sm ${user.is_active ? 'btn-warning' : 'btn-success'}`}
                    onClick={() => onStatusToggle(user)}
                  >
                    {user.is_active ? t('user.action.deactivate') : t('user.action.activate')}
                  </button>
                  <button
                    type="button"
                    className="btn btn-danger btn-sm"
                    onClick={() => onDelete(user)}
                  >
                    {t('common.delete')}
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
