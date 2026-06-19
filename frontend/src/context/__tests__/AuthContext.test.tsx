import { render, screen, act, waitFor } from '@testing-library/react';
import { AuthProvider, useAuth } from '../AuthContext';
import * as authUtils from '../../utils/auth';

vi.mock('../../utils/auth', () => ({
  getToken: vi.fn(),
  getStoredUser: vi.fn(),
  clearAuthStorage: vi.fn(),
  isTokenExpired: vi.fn(),
  setToken: vi.fn(),
  storeUser: vi.fn(),
  removeToken: vi.fn(),
  removeUser: vi.fn(),
}));

const mockGetToken = vi.mocked(authUtils.getToken);
const mockGetStoredUser = vi.mocked(authUtils.getStoredUser);
const mockClearAuthStorage = vi.mocked(authUtils.clearAuthStorage);
const mockIsTokenExpired = vi.mocked(authUtils.isTokenExpired);
const mockSetToken = vi.mocked(authUtils.setToken);
const mockStoreUser = vi.mocked(authUtils.storeUser);

const mockUser: authUtils.AuthUser = {
  id: 'user-1',
  username: 'testuser',
  display_name: 'Test User',
  email: 'test@example.com',
  role: 'user',
  is_active: true,
};

const VALID_TOKEN = 'valid.token.here';

function AuthStateDisplay() {
  const { user, token, isAuthenticated } = useAuth();
  return (
    <div>
      <span data-testid="is-authenticated">{String(isAuthenticated)}</span>
      <span data-testid="user">{user ? user.username : 'null'}</span>
      <span data-testid="token">{token ?? 'null'}</span>
    </div>
  );
}

function LoginButton() {
  const { login } = useAuth();
  return (
    <button onClick={() => login(mockUser, VALID_TOKEN)}>ログイン</button>
  );
}

function LogoutButton() {
  const { logout } = useAuth();
  return <button onClick={() => logout()}>ログアウト</button>;
}

afterEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
});

describe('AuthProvider', () => {
  describe('初期状態', () => {
    it('localStorageにトークンがない場合、isAuthenticated=false、user=nullである', async () => {
      mockGetToken.mockReturnValue(null);

      render(
        <AuthProvider>
          <AuthStateDisplay />
        </AuthProvider>,
      );

      await waitFor(() => {
        expect(screen.getByTestId('is-authenticated').textContent).toBe('false');
        expect(screen.getByTestId('user').textContent).toBe('null');
        expect(screen.getByTestId('token').textContent).toBe('null');
      });
    });

    it('localStorageに有効なトークンとユーザーがある場合、マウント後にisAuthenticated=trueになる', async () => {
      mockGetToken.mockReturnValue(VALID_TOKEN);
      mockIsTokenExpired.mockReturnValue(false);
      mockGetStoredUser.mockReturnValue(mockUser);

      render(
        <AuthProvider>
          <AuthStateDisplay />
        </AuthProvider>,
      );

      await waitFor(() => {
        expect(screen.getByTestId('is-authenticated').textContent).toBe('true');
        expect(screen.getByTestId('user').textContent).toBe(mockUser.username);
        expect(screen.getByTestId('token').textContent).toBe(VALID_TOKEN);
      });
    });

    it('localStorageにトークンがあっても期限切れの場合、clearAuthStorageが呼ばれisAuthenticated=falseになる', async () => {
      mockGetToken.mockReturnValue(VALID_TOKEN);
      mockIsTokenExpired.mockReturnValue(true);

      render(
        <AuthProvider>
          <AuthStateDisplay />
        </AuthProvider>,
      );

      await waitFor(() => {
        expect(mockClearAuthStorage).toHaveBeenCalled();
        expect(screen.getByTestId('is-authenticated').textContent).toBe('false');
      });
    });

    it('localStorageにトークンがあるがユーザーデータがない場合、clearAuthStorageが呼ばれる', async () => {
      mockGetToken.mockReturnValue(VALID_TOKEN);
      mockIsTokenExpired.mockReturnValue(false);
      mockGetStoredUser.mockReturnValue(null);

      render(
        <AuthProvider>
          <AuthStateDisplay />
        </AuthProvider>,
      );

      await waitFor(() => {
        expect(mockClearAuthStorage).toHaveBeenCalled();
        expect(screen.getByTestId('is-authenticated').textContent).toBe('false');
      });
    });
  });

  describe('login()', () => {
    it('login()を呼ぶとuser・token・isAuthenticatedが更新される', async () => {
      mockGetToken.mockReturnValue(null);

      render(
        <AuthProvider>
          <AuthStateDisplay />
          <LoginButton />
        </AuthProvider>,
      );

      await waitFor(() => {
        expect(screen.getByTestId('is-authenticated').textContent).toBe('false');
      });

      await act(async () => {
        screen.getByRole('button', { name: 'ログイン' }).click();
      });

      expect(mockSetToken).toHaveBeenCalledWith(VALID_TOKEN);
      expect(mockStoreUser).toHaveBeenCalledWith(mockUser);
      expect(screen.getByTestId('is-authenticated').textContent).toBe('true');
      expect(screen.getByTestId('user').textContent).toBe(mockUser.username);
      expect(screen.getByTestId('token').textContent).toBe(VALID_TOKEN);
    });
  });

  describe('logout()', () => {
    it('logout()を呼ぶとuser=null・token=null・isAuthenticated=falseになる', async () => {
      mockGetToken.mockReturnValue(VALID_TOKEN);
      mockIsTokenExpired.mockReturnValue(false);
      mockGetStoredUser.mockReturnValue(mockUser);

      render(
        <AuthProvider>
          <AuthStateDisplay />
          <LogoutButton />
        </AuthProvider>,
      );

      await waitFor(() => {
        expect(screen.getByTestId('is-authenticated').textContent).toBe('true');
      });

      await act(async () => {
        screen.getByRole('button', { name: 'ログアウト' }).click();
      });

      expect(mockClearAuthStorage).toHaveBeenCalled();
      expect(screen.getByTestId('is-authenticated').textContent).toBe('false');
      expect(screen.getByTestId('user').textContent).toBe('null');
      expect(screen.getByTestId('token').textContent).toBe('null');
    });
  });
});

describe('useAuth', () => {
  it('AuthProvider の外で useAuth() を呼ぶと Error が throw される', () => {
    function BadComponent() {
      useAuth();
      return null;
    }

    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() => render(<BadComponent />)).toThrow(
      'useAuth は AuthProvider の内側で使用してください',
    );
    spy.mockRestore();
  });
});
