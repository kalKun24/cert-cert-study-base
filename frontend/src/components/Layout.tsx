import { useState, useEffect, useCallback, useRef } from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { NavLink } from 'react-router-dom';
import NavBar from './NavBar';
import { useAuth } from '../context/AuthContext';
import { useInvitationCount } from '../hooks/useInvitationCount';
import apiClient from '../utils/apiClient';

// ロールの表示ラベル
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

/** フォーカス可能な要素のセレクタ */
const FOCUSABLE_SELECTORS =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

export default function Layout() {
  const { t } = useTranslation();
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const drawerRef = useRef<HTMLDivElement>(null);
  const { count: invitationCount } = useInvitationCount();

  const handleLogout = async () => {
    try {
      await apiClient.post('/auth/logout');
    } catch {
      // サーバーサイドはステートレスなので失敗しても続行
    }
    logout();
    navigate('/login', { replace: true });
  };

  const closeMobileMenu = useCallback(() => {
    setIsMobileMenuOpen(false);
  }, []);

  const handleMobileMenuToggle = () => {
    setIsMobileMenuOpen((prev) => !prev);
  };

  // ESCキーでドロワーを閉じる
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isMobileMenuOpen) {
        closeMobileMenu();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isMobileMenuOpen, closeMobileMenu]);

  // モバイルドロワー開閉時のフォーカストラップ
  useEffect(() => {
    if (!isMobileMenuOpen) return;

    const drawer = drawerRef.current;
    if (!drawer) return;

    const getFocusableElements = (): HTMLElement[] =>
      Array.from(drawer.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTORS));

    // ドロワーを開いたとき最初の要素にフォーカスを当てる
    const focusableElements = getFocusableElements();
    if (focusableElements.length > 0) {
      focusableElements[0].focus();
    }

    // Tabキーをドロワー内でループさせる
    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;

      const elements = getFocusableElements();
      if (elements.length === 0) return;

      const first = elements[0];
      const last = elements[elements.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === first) {
          e.preventDefault();
          last.focus();
        }
      } else {
        if (document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };

    document.addEventListener('keydown', handleTabKey);
    return () => {
      document.removeEventListener('keydown', handleTabKey);
    };
  }, [isMobileMenuOpen]);

  // ドロワー開閉時に body のスクロールを制御する
  useEffect(() => {
    if (isMobileMenuOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isMobileMenuOpen]);

  // ドロワー開閉時に inert 属性を DOM 直接操作で設定する（型定義未整備のため）
  useEffect(() => {
    const drawer = drawerRef.current;
    if (!drawer) return;
    if (isMobileMenuOpen) {
      drawer.removeAttribute('inert');
    } else {
      drawer.setAttribute('inert', '');
    }
  }, [isMobileMenuOpen]);

  return (
    <div className="app-layout">
      <NavBar
        isMobileMenuOpen={isMobileMenuOpen}
        onMobileMenuToggle={handleMobileMenuToggle}
        invitationCount={invitationCount}
      />

      {/* モバイルドロワーオーバーレイ（タップで閉じる） */}
      {isMobileMenuOpen && (
        <div
          className="mobile-overlay"
          onClick={closeMobileMenu}
          aria-hidden="true"
        />
      )}

      {/* モバイルドロワーナビゲーション */}
      <div
        id="mobile-menu"
        ref={drawerRef}
        className={`mobile-drawer${isMobileMenuOpen ? ' mobile-drawer--open' : ''}`}
        aria-label={t('nav.mobileMenu')}
        role="dialog"
        aria-modal="true"
        aria-hidden={!isMobileMenuOpen}
      >
        <nav>
          <ul className="mobile-nav-list" role="list">
            <li>
              <NavLink
                to="/"
                end
                className={({ isActive }) =>
                  isActive ? 'mobile-nav-link mobile-nav-link--active' : 'mobile-nav-link'
                }
                onClick={closeMobileMenu}
              >
                {t('nav.home')}
              </NavLink>
            </li>
            <li>
              <NavLink
                to="/questions"
                className={({ isActive }) =>
                  isActive ? 'mobile-nav-link mobile-nav-link--active' : 'mobile-nav-link'
                }
                onClick={closeMobileMenu}
              >
                {t('nav.questions')}
              </NavLink>
            </li>
            <li>
              <NavLink
                to="/notes"
                className={({ isActive }) =>
                  isActive ? 'mobile-nav-link mobile-nav-link--active' : 'mobile-nav-link'
                }
                onClick={closeMobileMenu}
              >
                {t('nav.notes')}
              </NavLink>
            </li>
            <li>
              <NavLink
                to="/tags"
                className={({ isActive }) =>
                  isActive ? 'mobile-nav-link mobile-nav-link--active' : 'mobile-nav-link'
                }
                onClick={closeMobileMenu}
              >
                {t('nav.tags')}
              </NavLink>
            </li>
            <li>
              <NavLink
                to="/teams"
                className={({ isActive }) =>
                  isActive ? 'mobile-nav-link mobile-nav-link--active' : 'mobile-nav-link'
                }
                onClick={closeMobileMenu}
              >
                {t('nav.teams')}
              </NavLink>
            </li>
            {/* ユーザー管理は admin のみ表示 */}
            {user?.role === 'admin' && (
              <li>
                <NavLink
                  to="/admin/users"
                  className={({ isActive }) =>
                    isActive ? 'mobile-nav-link mobile-nav-link--active' : 'mobile-nav-link'
                  }
                  onClick={closeMobileMenu}
                >
                  {t('nav.users')}
                </NavLink>
              </li>
            )}
            {/* 招待一覧（pending 件数があるときのみ表示） */}
            {invitationCount > 0 && (
              <li>
                <NavLink
                  to="/invitations"
                  className={({ isActive }) =>
                    isActive ? 'mobile-nav-link mobile-nav-link--active' : 'mobile-nav-link'
                  }
                  onClick={closeMobileMenu}
                >
                  {t('nav.invitations')}
                  <span
                    className="nav-badge"
                    aria-label={t('nav.invitationsBadgeLabel', { count: invitationCount })}
                  >
                    {invitationCount}
                  </span>
                </NavLink>
              </li>
            )}
          </ul>
        </nav>

        {/* モバイルドロワー内ユーザー情報・ログアウト */}
        {user && (
          <div className="mobile-drawer-user">
            <span className="mobile-drawer-user-info">
              {user.display_name}
              <span
                className="role-badge"
                data-role={ROLE_DATA[user.role] ?? 'user'}
              >
                {ROLE_LABELS[user.role] ?? user.role}
              </span>
            </span>
            <button
              type="button"
              onClick={handleLogout}
              className="btn btn-secondary btn-sm mobile-drawer-logout"
            >
              {t('nav.logout')}
            </button>
          </div>
        )}
      </div>

      <div className="app-body">
        <main className="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
