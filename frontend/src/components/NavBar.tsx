import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import apiClient from '../utils/apiClient';

const ROLE_LABELS: Record<string, string> = {
  admin: '管理者',
  teamowner: 'チームオーナー',
  user: 'ユーザー',
};

export default function NavBar() {
  const { t } = useTranslation();
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await apiClient.post('/auth/logout');
    } catch {
      // サーバーサイドはステートレスなので失敗しても続行
    }
    logout();
    navigate('/login', { replace: true });
  };

  return (
    <nav className="navbar">
      <div className="navbar-brand">
        <Link to="/">{t('app.title')}</Link>
      </div>

      <div className="navbar-links">
        <Link to="/">{t('nav.home')}</Link>
        <Link to="/questions">{t('nav.questions')}</Link>
        <Link to="/tags">{t('nav.tags')}</Link>
        <Link to="/teams">{t('nav.teams')}</Link>
      </div>

      <div className="navbar-user">
        {user && (
          <>
            <span className="user-info">
              {user.display_name}
              <span className="role-badge">{ROLE_LABELS[user.role] ?? user.role}</span>
            </span>
            <button onClick={handleLogout} className="btn btn-secondary">
              {t('nav.logout')}
            </button>
          </>
        )}
      </div>
    </nav>
  );
}
