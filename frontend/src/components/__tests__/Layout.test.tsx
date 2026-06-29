import { render, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Layout from '../Layout';
import * as AuthContextModule from '../../context/AuthContext';
import * as TeamContextModule from '../../context/TeamContext';
import * as InvitationCountModule from '../../hooks/useInvitationCount';
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

// 招待件数フックをモック
vi.mock('../../hooks/useInvitationCount', () => ({
  useInvitationCount: vi.fn(),
}));

const mockUseAuth = vi.mocked(AuthContextModule.useAuth);
const mockUseTeam = vi.mocked(TeamContextModule.useTeam);
const mockUseInvitationCount = vi.mocked(InvitationCountModule.useInvitationCount);

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

function renderLayout() {
  return render(
    <MemoryRouter>
      <Layout />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  mockUseAuth.mockReturnValue({
    isAuthenticated: true,
    isAuthLoading: false,
    user: mockUser,
    token: 'token',
    login: vi.fn(),
    logout: vi.fn(),
  });
  mockUseInvitationCount.mockReturnValue({
    count: 0,
    refresh: vi.fn(),
  });
});

afterEach(() => {
  vi.clearAllMocks();
});

// モバイルドロワーのチーム切り替えセレクト（NavBar デスクトップ版と区別するため id で特定）
const MOBILE_TEAM_SELECT_ID = 'mobile-team-select';

describe('Layout モバイルドロワーのチーム切り替え', () => {
  it('所属チームが2つ以上のときチーム切り替えセレクトを表示する', () => {
    setupTeam([teamA, teamB]);

    renderLayout();

    const select = document.getElementById(MOBILE_TEAM_SELECT_ID);
    expect(select).toBeInTheDocument();
    // アクセシビリティ: label と aria-label が関連付けられている
    expect(select).toHaveAttribute('aria-label', 'nav.switchTeam');
  });

  it('モバイルドロワーでチームを選択すると setActiveTeam が呼ばれる', () => {
    const setActiveTeam = setupTeam([teamA, teamB]);

    renderLayout();

    const select = document.getElementById(MOBILE_TEAM_SELECT_ID);
    expect(select).not.toBeNull();
    fireEvent.change(select as HTMLSelectElement, { target: { value: 'team-2' } });

    expect(setActiveTeam).toHaveBeenCalledTimes(1);
    expect(setActiveTeam).toHaveBeenCalledWith(teamB);
  });

  it('所属チームが1つの場合は切り替えセレクトを表示しない', () => {
    setupTeam([teamA]);

    renderLayout();

    expect(document.getElementById(MOBILE_TEAM_SELECT_ID)).toBeNull();
  });

  it('所属チームが0件の場合は切り替えセレクトを表示しない', () => {
    setupTeam([]);

    renderLayout();

    expect(document.getElementById(MOBILE_TEAM_SELECT_ID)).toBeNull();
  });
});
