import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import PrivateRoute from './components/PrivateRoute';
import Layout from './components/Layout';
import LoginPage from './pages/LoginPage';
import HomePage from './pages/HomePage';
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

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />

          <Route element={<PrivateRoute />}>
            <Route element={<Layout />}>
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
              <Route path="/profile/edit" element={<ProfileEditPage />} />
              <Route element={<PrivateRoute requiredRoles={['admin']} />}>
                <Route path="/admin/users" element={<UserListPage />} />
                <Route path="/admin/users/new" element={<UserCreatePage />} />
                <Route path="/admin/users/:id/edit" element={<UserEditPage />} />
              </Route>
            </Route>
          </Route>

          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
