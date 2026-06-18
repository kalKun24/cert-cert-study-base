import { Outlet } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import NavBar from './NavBar';
import { useAuth } from '../context/AuthContext';

export default function Layout() {
  const { t } = useTranslation();
  const { user } = useAuth();

  return (
    <div className="app-layout">
      <NavBar />
      <div className="app-body">
        <aside className="sidebar">
          <ul className="sidebar-menu">
            <li>
              <Link to="/">{t('nav.home')}</Link>
            </li>
            <li>
              <Link to="/questions">{t('nav.questions')}</Link>
            </li>
            <li>
              <Link to="/tags">{t('nav.tags')}</Link>
            </li>
            {user?.role === 'admin' && (
              <li>
                <Link to="/users">{t('nav.users')}</Link>
              </li>
            )}
          </ul>
        </aside>
        <main className="main-content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
