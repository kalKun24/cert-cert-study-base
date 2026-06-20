import { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { TeamProvider, useTeam } from './context/TeamContext';
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
import { fetchMyInvitations } from './utils/invitationApi';

/**
 * チーム所属チェックゲート。
 * 所属チームがない場合は /invitations または /no-team にリダイレクトし、
 * すべてのコンテンツルートへの素通りを防ぐ。
 */
function TeamGate() {
  const { teams, isLoading } = useTeam();
  const [redirectTo, setRedirectTo] = useState<string | null>(null);
  const [invLoading, setInvLoading] = useState(false);

  useEffect(() => {
    if (isLoading) return;
    if (teams.length > 0) {
      setRedirectTo(null);
      return;
    }
    let cancelled = false;
    setInvLoading(true);
    fetchMyInvitations()
      .then((all) => {
        if (!cancelled) {
          const hasPending = all.some((inv) => inv.status === 'pending');
          setRedirectTo(hasPending ? '/invitations' : '/no-team');
        }
      })
      .catch(() => {
        if (!cancelled) setRedirectTo('/no-team');
      })
      .finally(() => {
        if (!cancelled) setInvLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [isLoading, teams.length]);

  if (isLoading || invLoading) {
    return <p className="page-loading">読み込み中...</p>;
  }

  if (redirectTo) {
    return <Navigate to={redirectTo} replace />;
  }

  return <Outlet />;
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
                {/* チーム所属不要なページ（TeamGate の外） */}
                <Route path="/no-team" element={<NoTeamPage />} />
                <Route path="/invitations" element={<InvitationListPage />} />
                <Route path="/profile/edit" element={<ProfileEditPage />} />

                {/* チーム所属必須なページ（TeamGate でガード） */}
                <Route element={<TeamGate />}>
                  <Route path="/" element={<HomePage />} />
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
                  <Route element={<PrivateRoute requiredRoles={['admin']} />}>
                    <Route path="/admin/users" element={<UserListPage />} />
                    <Route path="/admin/users/new" element={<UserCreatePage />} />
                    <Route path="/admin/users/:id/edit" element={<UserEditPage />} />
                  </Route>
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
