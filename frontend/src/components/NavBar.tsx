import { useTranslation } from 'react-i18next';
import { Link, NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import apiClient from '../utils/apiClient';

// ロールの表示ラベル（翻訳ファイルではなく内部マッピング）
const ROLE_LABELS: Record<string, string> = {
  admin: '管理者',
  teamowner: 'チームオーナー',
  user: 'ユーザー',
};

// ロールの data 属性値
const ROLE_DATA: Record<string, string> = {
  admin: 'admin',
  teamowner: 'teamowner',
  user: 'user',
};

interface NavBarProps {
  /** モバイル用ドロワーの開閉制御 */
  isMobileMenuOpen: boolean;
  onMobileMenuToggle: () => void;
}

export default function NavBar({ isMobileMenuOpen, onMobileMenuToggle }: NavBarProps) {
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
    <header className="topbar" role="banner">
      <nav
        className="topbar-inner"
        aria-label={t('nav.topbar')}
      >
        {/* ロゴ（左端） */}
        <div className="topbar-logo">
          <Link to="/" className="topbar-logo-link">
            StudyBase
          </Link>
        </div>

        {/* 水平ナビリンク（中央・デスクトップのみ表示） */}
        <ul className="topbar-nav" role="list">
          <li>
            <NavLink to="/" end className={({ isActive }) => isActive ? 'topbar-nav-link topbar-nav-link--active' : 'topbar-nav-link'}>
              {t('nav.home')}
            </NavLink>
          </li>
          <li>
            <NavLink to="/questions" className={({ isActive }) => isActive ? 'topbar-nav-link topbar-nav-link--active' : 'topbar-nav-link'}>
              {t('nav.questions')}
            </NavLink>
          </li>
          <li>
            <NavLink to="/tags" className={({ isActive }) => isActive ? 'topbar-nav-link topbar-nav-link--active' : 'topbar-nav-link'}>
              {t('nav.tags')}
            </NavLink>
          </li>
          {/* ユーザー管理は admin のみ表示 */}
          {user?.role === 'admin' && (
            <li>
              <NavLink to="/admin/users" className={({ isActive }) => isActive ? 'topbar-nav-link topbar-nav-link--active' : 'topbar-nav-link'}>
                {t('nav.users')}
              </NavLink>
            </li>
          )}
        </ul>

        {/* ユーザーメニュー（右端） */}
        <div className="topbar-user">
          {user && (
            <>
              <span className="topbar-user-info">
                {user.display_name}
                <span
                  className="role-badge"
                  data-role={ROLE_DATA[user.role] ?? 'user'}
                >
                  {ROLE_LABELS[user.role] ?? user.role}
                </span>
              </span>
              <Link to="/profile/edit" className="btn btn-secondary btn-sm">
                {t('nav.profile')}
              </Link>
              <button
                type="button"
                onClick={handleLogout}
                className="btn btn-secondary btn-sm"
              >
                {t('nav.logout')}
              </button>
            </>
          )}
        </div>

        {/* ハンバーガーボタン（モバイルのみ表示） */}
        <button
          type="button"
          className="topbar-hamburger"
          onClick={onMobileMenuToggle}
          aria-label={isMobileMenuOpen ? t('nav.closeMenu') : t('nav.openMenu')}
          aria-expanded={isMobileMenuOpen}
          aria-controls="mobile-menu"
        >
          {/* ハンバーガーアイコン（SVG） */}
          <svg
            width="20"
            height="20"
            viewBox="0 0 20 20"
            fill="none"
            aria-hidden="true"
            focusable="false"
          >
            {isMobileMenuOpen ? (
              /* 閉じるアイコン（×） */
              <path
                d="M4 4L16 16M16 4L4 16"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
              />
            ) : (
              /* ハンバーガーアイコン（≡） */
              <>
                <path d="M3 5H17" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
                <path d="M3 10H17" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
                <path d="M3 15H17" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
              </>
            )}
          </svg>
        </button>
      </nav>
    </header>
  );
}
