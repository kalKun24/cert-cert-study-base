import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { Team } from '../types/team';
import { fetchTeams } from '../utils/teamApi';

interface TeamContextValue {
  teams: Team[];
  activeTeam: Team | null;
  setActiveTeam: (team: Team) => void;
  isLoading: boolean;
  refreshTeams: () => Promise<void>;
}

const TeamContext = createContext<TeamContextValue | null>(null);

const ACTIVE_TEAM_KEY = 'activeTeamId';

export function TeamProvider({ children }: { children: ReactNode }) {
  const [teams, setTeams] = useState<Team[]>([]);
  const [activeTeam, setActiveTeamState] = useState<Team | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const loadTeams = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await fetchTeams();
      setTeams(data);

      const savedId = localStorage.getItem(ACTIVE_TEAM_KEY);
      const restored = savedId ? data.find((t) => t.id === savedId) ?? null : null;
      setActiveTeamState(restored ?? data[0] ?? null);
    } catch {
      setTeams([]);
      setActiveTeamState(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadTeams();
  }, [loadTeams]);

  const setActiveTeam = useCallback((team: Team) => {
    localStorage.setItem(ACTIVE_TEAM_KEY, team.id);
    setActiveTeamState(team);
  }, []);

  const refreshTeams = useCallback(async () => {
    await loadTeams();
  }, [loadTeams]);

  return (
    <TeamContext.Provider value={{ teams, activeTeam, setActiveTeam, isLoading, refreshTeams }}>
      {children}
    </TeamContext.Provider>
  );
}

export function useTeam(): TeamContextValue {
  const ctx = useContext(TeamContext);
  if (!ctx) throw new Error('useTeam は TeamProvider の内側で使用してください');
  return ctx;
}
