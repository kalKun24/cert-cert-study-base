import apiClient from './apiClient';
import { Tag, TagListResponse, TagResponse } from '../types/tag';

export async function fetchTags(teamId: string): Promise<Tag[]> {
  const res = await apiClient.get<TagListResponse>(`/teams/${teamId}/tags`);
  return res.data.data;
}

export async function createTag(teamId: string, name: string): Promise<Tag> {
  const res = await apiClient.post<TagResponse>(`/teams/${teamId}/tags`, { name });
  return res.data.data;
}

export async function updateTag(teamId: string, id: string, name: string): Promise<Tag> {
  const res = await apiClient.put<TagResponse>(`/teams/${teamId}/tags/${id}`, { name });
  return res.data.data;
}

export async function deleteTag(teamId: string, id: string): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/tags/${id}`);
}
