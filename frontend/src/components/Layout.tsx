import { useState, useEffect, useCallback, useRef } from 'react';
import { Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { NavLink } from 'react-router-dom';
import NavBar from './NavBar';
import { useAuth } from '../context/AuthContext';

const FOCUSABLE_SELECTORS = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

export default function Layout() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const sidebarRef = useRef<HTMLElement>(null);

  const closeSidebar = useCallback(() => {
    setIsSidebarOpen(false);
  }, []);

  // ESCキーでサイドバーを閉じる
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isSidebarOpen) {
        closeSidebar();
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [isSidebarOpen, closeSidebar]);

  // モバイル: サイドバーオーバーレイ表示中のフォーカストラップ
  useEffect(() => {
    if (!isSidebarOpen) return;

    const sidebar = sidebarRef.current;
    if (!sidebar) return;

    // サイドバー内のフォーカス可能な要素を取得
    const getFocusableElements = (): HTMLElement[] =>
      Array.from(sidebar.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTORS));

    // サイドバーを開いたとき最初の要素にフォーカスを当てる
    const focusableElements = getFocusableElements();
    if (focusableElements.length > 0) {
      focusableElements[0].focus();
    }

    // Tabキーをサイドバー内でループさせる
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
  }, [isSidebarOpen]);

  const handleMenuToggle = () => {
    setIsSidebarOpen((prev) => !prev);
  };

  return (
    <div className="app-layout">
      <NavBar onMenuToggle={handleMenuToggle} isSidebarOpen={isSidebarOpen} />
      <div className="app-body">
        {/* モバイル: サイドバーオーバーレイ（タップで閉じる） */}
        {isSidebarOpen && (
          <div
            className="sidebar-overlay"
            onClick={closeSidebar}
            aria-hidden="true"
          />
        )}

        <aside
          id="sidebar"
          ref={sidebarRef}
          className={`sidebar${isSidebarOpen ? ' is-open' : ''}`}
        >
          <nav aria-label={t('nav.sidebar')}>
            <ul className="sidebar-menu">
              <li>
                <NavLink to="/" end onClick={closeSidebar}>
                  {t('nav.home')}
                </NavLink>
              </li>
              <li>
                <NavLink to="/questions" onClick={closeSidebar}>
                  {t('nav.questions')}
                </NavLink>
              </li>
              <li>
                <NavLink to="/tags" onClick={closeSidebar}>
                  {t('nav.tags')}
                </NavLink>
              </li>
              <li>
                <NavLink to="/teams" onClick={closeSidebar}>
                  {t('nav.teams')}
                </NavLink>
              </li>
              {user?.role === 'admin' && (
                <li>
                  <NavLink to="/admin/users" onClick={closeSidebar}>
                    {t('nav.users')}
                  </NavLink>
                </li>
              )}
            </ul>
          </nav>
        </aside>

        <main className="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
