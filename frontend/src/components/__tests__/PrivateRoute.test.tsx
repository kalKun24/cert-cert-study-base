import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import PrivateRoute from '../PrivateRoute';
import * as AuthContextModule from '../../context/AuthContext';
import type { AuthUser } from '../../utils/auth';

vi.mock('../../context/AuthContext', () => ({
  useAuth: vi.fn(),
}));

const mockUseAuth = vi.mocked(AuthContextModule.useAuth);

const mockUser: AuthUser = {
  id: 'user-1',
  username: 'testuser',
  display_name: 'Test User',
  email: 'test@example.com',
  role: 'user',
  is_active: true,
};

const adminUser: AuthUser = {
  ...mockUser,
  id: 'admin-1',
  username: 'admin',
  role: 'admin',
};

function renderWithRouter(
  initialEntries: string[],
  requiredRoles?: Parameters<typeof PrivateRoute>[0]['requiredRoles'],
) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route path="/login" element={<div>ログインページ</div>} />
        <Route path="/" element={<PrivateRoute requiredRoles={requiredRoles} />}>
          <Route index element={<div>保護されたコンテンツ</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
  );
}

afterEach(() => {
  vi.restoreAllMocks();
});

describe('PrivateRoute', () => {
  describe('認証確認中の場合', () => {
    it('isAuthLoading=true のとき何もレンダリングしない', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: false,
        isAuthLoading: true,
        user: null,
        token: null,
        login: vi.fn(),
        logout: vi.fn(),
      });

      renderWithRouter(['/']);

      expect(screen.queryByText('ログインページ')).not.toBeInTheDocument();
      expect(screen.queryByText('保護されたコンテンツ')).not.toBeInTheDocument();
    });
  });

  describe('未認証の場合', () => {
    it('/login にリダイレクトされる', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: false,
        isAuthLoading: false,
        user: null,
        token: null,
        login: vi.fn(),
        logout: vi.fn(),
      });

      renderWithRouter(['/']);

      expect(screen.getByText('ログインページ')).toBeInTheDocument();
      expect(screen.queryByText('保護されたコンテンツ')).not.toBeInTheDocument();
    });
  });

  describe('認証済みの場合', () => {
    it('requiredRoles なし: Outlet がレンダリングされる', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        isAuthLoading: false,
        user: mockUser,
        token: 'token',
        login: vi.fn(),
        logout: vi.fn(),
      });

      renderWithRouter(['/']);

      expect(screen.getByText('保護されたコンテンツ')).toBeInTheDocument();
      expect(screen.queryByText('ログインページ')).not.toBeInTheDocument();
    });

    it('ロールが requiredRoles に含まれる場合: Outlet がレンダリングされる', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        isAuthLoading: false,
        user: adminUser,
        token: 'token',
        login: vi.fn(),
        logout: vi.fn(),
      });

      renderWithRouter(['/'], ['admin']);

      expect(screen.getByText('保護されたコンテンツ')).toBeInTheDocument();
    });

    it('ロールが requiredRoles に含まれない場合: / にリダイレクトされる（ルートにフォールバック）', () => {
      mockUseAuth.mockReturnValue({
        isAuthenticated: true,
        isAuthLoading: false,
        user: mockUser,
        token: 'token',
        login: vi.fn(),
        logout: vi.fn(),
      });

      // /admin ルートにアクセスしようとしているが admin ロールがない
      render(
        <MemoryRouter initialEntries={['/admin']}>
          <Routes>
            <Route path="/login" element={<div>ログインページ</div>} />
            <Route path="/" element={<div>ホームページ</div>} />
            <Route path="/admin" element={<PrivateRoute requiredRoles={['admin']} />}>
              <Route index element={<div>管理者コンテンツ</div>} />
            </Route>
          </Routes>
        </MemoryRouter>,
      );

      expect(screen.getByText('ホームページ')).toBeInTheDocument();
      expect(screen.queryByText('管理者コンテンツ')).not.toBeInTheDocument();
    });
  });
});
