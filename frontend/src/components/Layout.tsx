import { useState, useEffect, useCallback, useRef } from 'react';
import { Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { NavLink } from 'react-router-dom';
import NavBar from './NavBar';
import { useAuth } from '../context/AuthContext';

/** フォーカス可能な要素のセレクタ */
const FOCUSABLE_SELECTORS =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

export default function Layout() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const drawerRef = useRef<HTMLDivElement>(null);

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

  return (
    <div className="app-layout">
      <NavBar
        isMobileMenuOpen={isMobileMenuOpen}
        onMobileMenuToggle={handleMobileMenuToggle}
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
        hidden={!isMobileMenuOpen}
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
                to="/tags"
                className={({ isActive }) =>
                  isActive ? 'mobile-nav-link mobile-nav-link--active' : 'mobile-nav-link'
                }
                onClick={closeMobileMenu}
              >
                {t('nav.tags')}
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
          </ul>
        </nav>
      </div>

      <div className="app-body">
        <main className="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
