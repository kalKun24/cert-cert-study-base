import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import apiClient from '../utils/apiClient';

const ROLE_DATA: Record<string, string> = {
  admin: 'admin',
  teamowner: 'teamowner',
  user: 'user',
};

const ROLE_LABELS: Record<string, string> = {
  admin: '管理者',
  teamowner: 'チームオーナー',
  user: 'ユーザー',
};

interface NavBarProps {
  onMenuToggle?: () => void;
  isSidebarOpen?: boolean;
}

export default function NavBar({ onMenuToggle, isSidebarOpen = false }: NavBarProps) {
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
    <nav className="navbar" aria-label={t('app.title')}>
      {/* モバイル: ハンバーガーボタン */}
      <button
        type="button"
        className="navbar-menu-toggle"
        onClick={onMenuToggle}
        aria-label={isSidebarOpen ? t('nav.closeMenu') : t('nav.openMenu')}
        aria-expanded={isSidebarOpen}
        aria-controls="sidebar"
      >
        ☰
      </button>

      <div className="navbar-brand">
        <Link to="/">{t('app.title')}</Link>
      </div>

      <div className="navbar-user">
        {user && (
          <>
            <span className="user-info">
              {user.display_name}
              <span
                className="role-badge"
                data-role={ROLE_DATA[user.role] ?? 'user'}
              >
                {ROLE_LABELS[user.role] ?? user.role}
              </span>
            </span>
            <Link to="/profile/edit" className="btn btn-secondary">
              {t('nav.profile')}
            </Link>
            <button type="button" onClick={handleLogout} className="btn btn-secondary">
              {t('nav.logout')}
            </button>
          </>
        )}
      </div>
    </nav>
  );
}
