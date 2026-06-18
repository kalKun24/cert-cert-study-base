import { useState, FormEvent } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../context/AuthContext';
import apiClient from '../utils/apiClient';
import { AuthUser } from '../utils/auth';

interface LoginResponse {
  data: {
    token: string;
    user: AuthUser;
  };
  error: string | null;
}

export default function LoginPage() {
  const { t } = useTranslation();
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const from = (location.state as { from?: string })?.from ?? '/';

  const validate = (): string => {
    if (!username.trim()) return t('login.validation.usernameRequired');
    if (!password) return t('login.validation.passwordRequired');
    if (password.length < 8) return t('login.validation.passwordTooShort');
    return '';
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const validationError = validate();
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsLoading(true);
    setError('');
    try {
      const res = await apiClient.post<LoginResponse>('/auth/login', { username, password });
      login(res.data.data.user, res.data.data.token);
      navigate(from, { replace: true });
    } catch (err: unknown) {
      if (
        typeof err === 'object' &&
        err !== null &&
        'response' in err &&
        typeof (err as { response?: { status?: number } }).response?.status === 'number'
      ) {
        const status = (err as { response: { status: number } }).response.status;
        if (status === 401) {
          setError(t('login.error.invalidCredentials'));
        } else {
          setError(t('common.error'));
        }
      } else {
        setError(t('common.error'));
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <h1 className="login-title">{t('app.title')}</h1>
        <p className="login-subtitle">{t('app.description')}</p>

        <form onSubmit={handleSubmit} noValidate>
          {error && (
            <div className="alert alert-error" role="alert">
              {error}
            </div>
          )}

          <div className="form-group">
            <label htmlFor="username">{t('login.username')}</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
              disabled={isLoading}
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">{t('login.password')}</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              disabled={isLoading}
            />
          </div>

          <button type="submit" className="btn btn-primary" disabled={isLoading}>
            {isLoading ? t('common.loading') : t('nav.login')}
          </button>
        </form>
      </div>
    </div>
  );
}
