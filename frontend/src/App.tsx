import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { TeamProvider } from './context/TeamContext';
import PrivateRoute from './components/PrivateRoute';
import Layout from './components/Layout';
import LoginPage from './pages/LoginPage';
import HomePage from './pages/HomePage';
import NoTeamPage from './pages/NoTeamPage';
import InvitationListPage from './pages/InvitationListPage';
import QuestionListPage from './pages/QuestionListPage';
import QuestionDetailPage from './pages/QuestionDetailPage';
import QuestionCreatePage from './pages/QuestionCreatePage';
import QuestionEditPage from './pages/QuestionEditPage';
import TeamListPage from './pages/TeamListPage';
import TeamDetailPage from './pages/TeamDetailPage';
import TeamCreatePage from './pages/TeamCreatePage';
import TeamEditPage from './pages/TeamEditPage';
import TagManagePage from './pages/TagManagePage';
import UserListPage from './pages/UserListPage';
import UserCreatePage from './pages/UserCreatePage';
import UserEditPage from './pages/UserEditPage';
import ProfileEditPage from './pages/ProfileEditPage';
import { useTeam } from './context/TeamContext';
import { useEffect, useState } from 'react';
import { fetchMyInvitations } from './utils/invitationApi';

/**
 * ログイン後のルーティング分岐コンポーネント。
 * "/" ルートにのみ適用し、所属チームと招待の状態に応じてページを振り分ける。
 */
function TeamSelectionGate() {
  const { teams, isLoading } = useTeam();
  const [pendingCount, setPendingCount] = useState<number | null>(null);
  const [invLoading, setInvLoading] = useState(true);

  useEffect(() => {
    if (isLoading) return;
    if (teams.length > 0) {
      setInvLoading(false);
      return;
    }
    let cancelled = false;
    setInvLoading(true);
    fetchMyInvitations()
      .then((all) => {
        if (!cancelled) {
          setPendingCount(all.filter((inv) => inv.status === 'pending').length);
        }
      })
      .catch(() => {
        if (!cancelled) setPendingCount(0);
      })
      .finally(() => {
        if (!cancelled) setInvLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [isLoading, teams.length]);

  if (isLoading || invLoading) return null;

  if (teams.length > 0) {
    return <HomePage />;
  }

  if (pendingCount !== null && pendingCount > 0) {
    return <InvitationListPage />;
  }

  return <NoTeamPage />;
}

/** PrivateRoute内でTeamProviderをラップするためのラッパーコンポーネント */
function PrivateWithTeam() {
  return (
    <TeamProvider>
      <Outlet />
    </TeamProvider>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />

          <Route element={<PrivateRoute />}>
            <Route element={<PrivateWithTeam />}>
              <Route element={<Layout />}>
                <Route path="/" element={<TeamSelectionGate />} />
                <Route path="/no-team" element={<NoTeamPage />} />
                <Route path="/invitations" element={<InvitationListPage />} />
                <Route path="/questions" element={<QuestionListPage />} />
                <Route path="/questions/new" element={<QuestionCreatePage />} />
                <Route path="/questions/:id" element={<QuestionDetailPage />} />
                <Route path="/questions/:id/edit" element={<QuestionEditPage />} />
                <Route path="/tags" element={<TagManagePage />} />
                <Route path="/teams" element={<TeamListPage />} />
                <Route element={<PrivateRoute requiredRoles={['admin', 'teamowner']} />}>
                  <Route path="/teams/new" element={<TeamCreatePage />} />
                </Route>
                <Route path="/teams/:id" element={<TeamDetailPage />} />
                <Route path="/teams/:id/edit" element={<TeamEditPage />} />
                <Route path="/profile/edit" element={<ProfileEditPage />} />
                <Route element={<PrivateRoute requiredRoles={['admin']} />}>
                  <Route path="/admin/users" element={<UserListPage />} />
                  <Route path="/admin/users/new" element={<UserCreatePage />} />
                  <Route path="/admin/users/:id/edit" element={<UserEditPage />} />
                </Route>
              </Route>
            </Route>
          </Route>

          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
