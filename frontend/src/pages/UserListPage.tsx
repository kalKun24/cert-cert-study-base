import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { User } from '../types/user';
import { fetchUsers, updateUserStatus, deleteUser } from '../utils/userApi';
import UserTable from '../components/UserTable';

export default function UserListPage() {
  const { t } = useTranslation();

  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState('');

  const loadUsers = () => {
    setIsLoading(true);
    setFetchError('');
    fetchUsers()
      .then(setUsers)
      .catch(() => setFetchError(t('user.error.fetchFailed')))
      .finally(() => setIsLoading(false));
  };

  useEffect(() => {
    loadUsers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleStatusToggle = async (user: User) => {
    const confirmMsg = user.is_active
      ? t('user.confirm.deactivate', { name: user.display_name })
      : t('user.confirm.activate', { name: user.display_name });
    if (!window.confirm(confirmMsg)) return;
    try {
      const updated = await updateUserStatus(user.id, { is_active: !user.is_active });
      setUsers((prev) => prev.map((u) => (u.id === updated.id ? updated : u)));
    } catch {
      window.alert(t('user.error.statusUpdateFailed'));
    }
  };

  const handleDelete = async (user: User) => {
    const confirmed = window.confirm(
      t('user.confirm.delete', { name: user.display_name }),
    );
    if (!confirmed) return;
    try {
      await deleteUser(user.id);
      setUsers((prev) => prev.filter((u) => u.id !== user.id));
    } catch {
      window.alert(t('user.error.deleteFailed'));
    }
  };

  return (
    <section className="user-list-page">
      <div className="page-header">
        <h1 className="page-title">{t('user.list.title')}</h1>
        <Link to="/admin/users/new" className="btn btn-primary">
          {t('user.list.createButton')}
        </Link>
      </div>

      {isLoading ? (
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      ) : fetchError ? (
        <p role="alert" className="alert alert-error">
          {fetchError}
        </p>
      ) : users.length === 0 ? (
        <p className="user-list-empty">{t('user.list.empty')}</p>
      ) : (
        <UserTable
          users={users}
          onStatusToggle={handleStatusToggle}
          onDelete={handleDelete}
        />
      )}
    </section>
  );
}
