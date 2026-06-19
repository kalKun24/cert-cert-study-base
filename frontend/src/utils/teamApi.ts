import apiClient from './apiClient';
import {
  Team,
  TeamDetail,
  TeamsResponse,
  TeamResponse,
  TeamDetailResponse,
  CreateTeamRequest,
  UpdateTeamRequest,
  AddMemberRequest,
} from '../types/team';

export async function fetchTeams(): Promise<Team[]> {
  const res = await apiClient.get<TeamsResponse>('/teams');
  return res.data.data;
}

export async function fetchTeam(id: string): Promise<TeamDetail> {
  const res = await apiClient.get<TeamDetailResponse>(`/teams/${id}`);
  return res.data.data;
}

export async function createTeam(req: CreateTeamRequest): Promise<Team> {
  const res = await apiClient.post<TeamResponse>('/teams', req);
  return res.data.data;
}

export async function updateTeam(id: string, req: UpdateTeamRequest): Promise<Team> {
  const res = await apiClient.put<TeamResponse>(`/teams/${id}`, req);
  return res.data.data;
}

export async function deleteTeam(id: string): Promise<void> {
  await apiClient.delete(`/teams/${id}`);
}

export async function addMember(teamId: string, req: AddMemberRequest): Promise<void> {
  await apiClient.post(`/teams/${teamId}/members`, req);
}

export async function removeMember(teamId: string, userId: string): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/members/${userId}`);
}

export async function changeMemberRole(
  teamId: string,
  userId: string,
  role: 'owner' | 'member',
): Promise<void> {
  await apiClient.patch(`/teams/${teamId}/members/${userId}/role`, { role });
}
