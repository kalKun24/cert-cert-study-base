import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { User } from '../types/user';
import { fetchUsers } from '../utils/userApi';
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

  const handleStatusToggled = (updated: User) => {
    setUsers((prev) => prev.map((u) => (u.id === updated.id ? updated : u)));
  };

  const handleDeleted = (deletedId: string) => {
    setUsers((prev) => prev.filter((u) => u.id !== deletedId));
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
          onStatusToggled={handleStatusToggled}
          onDeleted={handleDeleted}
        />
      )}
    </section>
  );
}
