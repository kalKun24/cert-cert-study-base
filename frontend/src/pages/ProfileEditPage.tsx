import { useState, FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../context/AuthContext';
import apiClient from '../utils/apiClient';
import { AuthUser } from '../utils/auth';

interface UpdateProfileResponse {
  data: {
    id: string;
    username: string;
    display_name: string;
    role: string;
  };
  error: string | null;
}

interface UpdatePasswordResponse {
  data: null;
  error: string | null;
}

export default function ProfileEditPage() {
  const { t } = useTranslation();
  const { user, login, token, logout } = useAuth();
  const navigate = useNavigate();

  // 表示名フォームの状態
  const [displayName, setDisplayName] = useState(user?.display_name ?? '');
  const [displayNameError, setDisplayNameError] = useState('');
  const [displayNameSuccess, setDisplayNameSuccess] = useState('');
  const [isDisplayNameSubmitting, setIsDisplayNameSubmitting] = useState(false);

  // パスワードフォームの状態
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [isPasswordSubmitting, setIsPasswordSubmitting] = useState(false);

  const validateDisplayName = (): string => {
    if (!displayName.trim()) return t('profile.validation.displayNameRequired');
    return '';
  };

  const validatePassword = (): string => {
    if (!currentPassword) return t('profile.validation.currentPasswordRequired');
    if (!newPassword) return t('profile.validation.newPasswordRequired');
    if (newPassword.length < 8) return t('profile.validation.newPasswordTooShort');
    if (newPassword.length > 72) return t('profile.validation.newPasswordTooLong');
    if (!confirmPassword) return t('profile.validation.confirmPasswordRequired');
    if (newPassword !== confirmPassword) return t('profile.validation.passwordMismatch');
    return '';
  };

  const handleDisplayNameSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const validationError = validateDisplayName();
    if (validationError) {
      setDisplayNameError(validationError);
      setDisplayNameSuccess('');
      return;
    }

    setIsDisplayNameSubmitting(true);
    setDisplayNameError('');
    setDisplayNameSuccess('');

    try {
      const res = await apiClient.patch<UpdateProfileResponse>('/users/me/profile', {
        display_name: displayName.trim(),
      });

      // AuthContext のユーザー情報を更新
      if (user && token) {
        const updatedUser: AuthUser = {
          ...user,
          display_name: res.data.data.display_name,
        };
        login(updatedUser, token);
      }

      setDisplayNameSuccess(t('profile.displayName.success'));
    } catch {
      setDisplayNameError(t('profile.error.displayNameUpdateFailed'));
    } finally {
      setIsDisplayNameSubmitting(false);
    }
  };

  const handlePasswordSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const validationError = validatePassword();
    if (validationError) {
      setPasswordError(validationError);
      return;
    }

    setIsPasswordSubmitting(true);
    setPasswordError('');

    try {
      await apiClient.patch<UpdatePasswordResponse>('/users/me/password', {
        current_password: currentPassword,
        new_password: newPassword,
      });

      // パスワード変更成功後はログアウトしてログイン画面へリダイレクト
      logout();
      navigate('/login', { replace: true, state: { message: t('profile.password.successMessage') } });
    } catch (err: unknown) {
      if (
        typeof err === 'object' &&
        err !== null &&
        'response' in err &&
        typeof (err as { response?: { status?: number } }).response?.status === 'number'
      ) {
        const status = (err as { response: { status: number } }).response.status;
        if (status === 422) {
          setPasswordError(t('profile.error.currentPasswordIncorrect'));
        } else {
          setPasswordError(t('profile.error.passwordChangeFailed'));
        }
      } else {
        setPasswordError(t('profile.error.passwordChangeFailed'));
      }
      setIsPasswordSubmitting(false);
    }
  };

  return (
    <section className="profile-edit-page page-container-full">
      <h1 className="page-title">{t('profile.pageTitle')}</h1>

      {/* 表示名の変更セクション */}
      <div className="card">
        <h2 className="card-title">{t('profile.displayName.sectionTitle')}</h2>
        <form onSubmit={handleDisplayNameSubmit} noValidate>
          {displayNameError && (
            <div className="alert alert-error" role="alert">
              {displayNameError}
            </div>
          )}
          {displayNameSuccess && (
            <div className="alert alert-success" role="status">
              {displayNameSuccess}
            </div>
          )}

          <div className="form-group">
            <label htmlFor="display-name">{t('profile.displayName.label')}</label>
            <input
              id="display-name"
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder={t('profile.displayName.placeholder')}
              disabled={isDisplayNameSubmitting}
              autoComplete="name"
            />
          </div>

          <div className="form-actions">
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isDisplayNameSubmitting}
            >
              {isDisplayNameSubmitting ? t('common.loading') : t('profile.displayName.saveButton')}
            </button>
          </div>
        </form>
      </div>

      {/* パスワードの変更セクション */}
      <div className="card">
        <h2 className="card-title">{t('profile.password.sectionTitle')}</h2>
        <form onSubmit={handlePasswordSubmit} noValidate>
          {passwordError && (
            <div className="alert alert-error" role="alert">
              {passwordError}
            </div>
          )}

          <div className="form-group">
            <label htmlFor="current-password">{t('profile.password.currentLabel')}</label>
            <input
              id="current-password"
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              disabled={isPasswordSubmitting}
              autoComplete="current-password"
            />
          </div>

          <div className="form-group">
            <label htmlFor="new-password">{t('profile.password.newLabel')}</label>
            <input
              id="new-password"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              disabled={isPasswordSubmitting}
              autoComplete="new-password"
            />
          </div>

          <div className="form-group">
            <label htmlFor="confirm-password">{t('profile.password.confirmLabel')}</label>
            <input
              id="confirm-password"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              disabled={isPasswordSubmitting}
              autoComplete="new-password"
            />
          </div>

          <div className="form-actions">
            <button
              type="submit"
              className="btn btn-primary"
              disabled={isPasswordSubmitting}
            >
              {isPasswordSubmitting ? t('common.loading') : t('profile.password.changeButton')}
            </button>
          </div>
        </form>
      </div>
    </section>
  );
}
