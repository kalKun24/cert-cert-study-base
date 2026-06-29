import { render, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import NavBar from '../NavBar';
import * as AuthContextModule from '../../context/AuthContext';
import * as TeamContextModule from '../../context/TeamContext';
import type { AuthUser } from '../../utils/auth';
import type { Team } from '../../types/team';

// react-i18next をモック（キーをそのまま返す）
vi.mock('react-i18next', () => {
  const t = (key: string) => key;
  return {
    useTranslation: () => ({
      t,
      i18n: { language: 'ja' },
    }),
  };
});

// AuthContext をモック
vi.mock('../../context/AuthContext', () => ({
  useAuth: vi.fn(),
}));

// TeamContext をモック
vi.mock('../../context/TeamContext', () => ({
  useTeam: vi.fn(),
}));

// apiClient をモック（ログアウト時の呼び出し回避）
vi.mock('../../utils/apiClient', () => ({
  default: { post: vi.fn() },
}));

const mockUseAuth = vi.mocked(AuthContextModule.useAuth);
const mockUseTeam = vi.mocked(TeamContextModule.useTeam);

const mockUser: AuthUser = {
  id: 'user-1',
  username: 'testuser',
  display_name: 'Test User',
  email: 'test@example.com',
  role: 'user',
  is_active: true,
};

function makeTeam(id: string, name: string): Team {
  return {
    id,
    name,
    description: '',
    owner_id: 'owner-1',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  };
}

const teamA = makeTeam('team-1', 'チームA');
const teamB = makeTeam('team-2', 'チームB');

function setupTeam(teams: Team[], setActiveTeam = vi.fn()) {
  mockUseTeam.mockReturnValue({
    teams,
    activeTeam: teams[0] ?? null,
    setActiveTeam,
    isLoading: false,
    refreshTeams: vi.fn(),
  });
  return setActiveTeam;
}

function renderNavBar() {
  return render(
    <MemoryRouter>
      <NavBar isMobileMenuOpen={false} onMobileMenuToggle={vi.fn()} invitationCount={0} />
    </MemoryRouter>,
  );
}

// デスクトップ用チーム切り替えセレクトの id
const DESKTOP_TEAM_SELECT_ID = 'topbar-team-select';

beforeEach(() => {
  mockUseAuth.mockReturnValue({
    isAuthenticated: true,
    isAuthLoading: false,
    user: mockUser,
    token: 'token',
    login: vi.fn(),
    logout: vi.fn(),
  });
});

afterEach(() => {
  vi.clearAllMocks();
});

describe('NavBar デスクトップのチーム切り替え表示境界', () => {
  it('所属チームが0件のときセレクトを表示しない', () => {
    setupTeam([]);

    renderNavBar();

    expect(document.getElementById(DESKTOP_TEAM_SELECT_ID)).toBeNull();
  });

  it('所属チームが1件のときセレクトを表示しない', () => {
    setupTeam([teamA]);

    renderNavBar();

    expect(document.getElementById(DESKTOP_TEAM_SELECT_ID)).toBeNull();
  });

  it('所属チームが2件以上のときセレクトを表示する', () => {
    setupTeam([teamA, teamB]);

    renderNavBar();

    expect(document.getElementById(DESKTOP_TEAM_SELECT_ID)).toBeInTheDocument();
  });

  it('チームを選択すると setActiveTeam が正しいチームで呼ばれる', () => {
    const setActiveTeam = setupTeam([teamA, teamB]);

    renderNavBar();

    const select = document.getElementById(DESKTOP_TEAM_SELECT_ID);
    expect(select).not.toBeNull();
    fireEvent.change(select as HTMLSelectElement, { target: { value: 'team-2' } });

    expect(setActiveTeam).toHaveBeenCalledTimes(1);
    expect(setActiveTeam).toHaveBeenCalledWith(teamB);
  });
});
