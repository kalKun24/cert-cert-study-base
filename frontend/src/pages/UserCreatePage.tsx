import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { CreateUserRequest, UpdateUserRequest } from '../types/user';
import { createUser } from '../utils/userApi';
import UserForm from '../components/UserForm';

export default function UserCreatePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState('');

  const handleSubmit = async (data: CreateUserRequest | UpdateUserRequest) => {
    setIsSubmitting(true);
    setSubmitError('');
    try {
      await createUser(data as CreateUserRequest);
      navigate('/admin/users');
    } catch {
      setSubmitError(t('user.error.createFailed'));
      setIsSubmitting(false);
    }
  };

  return (
    <section className="user-form-page page-container-full">
      <h1 className="page-title">{t('user.form.createTitle')}</h1>
      <UserForm
        mode="create"
        isSubmitting={isSubmitting}
        submitError={submitError}
        onSubmit={handleSubmit}
        onCancel={() => navigate('/admin/users')}
      />
    </section>
  );
}
