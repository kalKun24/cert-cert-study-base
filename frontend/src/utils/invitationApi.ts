import apiClient from './apiClient';
import { Invitation, InvitationsResponse, InvitationResponse } from '../types/team';

export async function fetchMyInvitations(): Promise<Invitation[]> {
  const res = await apiClient.get<InvitationsResponse>('/invitations/me');
  return res.data.data;
}

export async function respondInvitation(
  id: string,
  status: 'accepted' | 'rejected',
): Promise<Invitation> {
  const res = await apiClient.patch<InvitationResponse>(`/invitations/${id}`, { status });
  return res.data.data;
}

export async function leaveTeam(teamId: string): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/members/me`);
}
