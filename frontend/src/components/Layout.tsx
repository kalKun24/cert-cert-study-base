import { useState, useEffect, useCallback } from 'react';
import { Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { NavLink } from 'react-router-dom';
import NavBar from './NavBar';
import { useAuth } from '../context/AuthContext';

export default function Layout() {
  const { t } = useTranslation();
  const { user } = useAuth();
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

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
