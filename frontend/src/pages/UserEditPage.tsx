import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { User, UpdateUserRequest, CreateUserRequest } from '../types/user';
import { fetchUser, updateUser } from '../utils/userApi';
import UserForm from '../components/UserForm';

export default function UserEditPage() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');

  useEffect(() => {
    if (!id) return;
    let isMounted = true;
    setIsLoading(true);
    setLoadError('');

    fetchUser(id)
      .then((data) => {
        if (isMounted) setUser(data);
      })
      .catch(() => {
        if (isMounted) setLoadError(t('user.error.fetchFailed'));
      })
      .finally(() => {
        if (isMounted) setIsLoading(false);
      });

    return () => {
      isMounted = false;
    };
  }, [id, t]);

  if (!id) return null;

  if (isLoading) {
    return (
      <section className="user-form-page page-container-full">
        <p role="status" className="page-loading">
          {t('common.loading')}
        </p>
      </section>
    );
  }

  if (loadError) {
    return (
      <section className="user-form-page page-container-full">
        <p role="alert" className="alert alert-error">
          {loadError}
        </p>
      </section>
    );
  }

  if (!user) return null;

  const handleSubmit = async (data: CreateUserRequest | UpdateUserRequest) => {
    setIsSubmitting(true);
    setSubmitError('');
    try {
      await updateUser(user.id, data as UpdateUserRequest);
      navigate('/admin/users');
    } catch {
      setSubmitError(t('user.error.updateFailed'));
      setIsSubmitting(false);
    }
  };

  return (
    <section className="user-form-page page-container-full">
      <h1 className="page-title">{t('user.form.editTitle')}</h1>
      <UserForm
        mode="edit"
        initialUsername={user.username}
        initialDisplayName={user.display_name}
        initialEmail={user.email}
        initialRole={user.role}
        isSubmitting={isSubmitting}
        submitError={submitError}
        onSubmit={handleSubmit}
        onCancel={() => navigate('/admin/users')}
      />
    </section>
  );
}
