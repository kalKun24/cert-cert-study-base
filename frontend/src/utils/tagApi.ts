import apiClient from './apiClient';
import { Tag, TagListResponse, TagResponse } from '../types/tag';

export async function fetchTags(): Promise<Tag[]> {
  const res = await apiClient.get<TagListResponse>('/tags');
  return res.data.data;
}

export async function createTag(name: string): Promise<Tag> {
  const res = await apiClient.post<TagResponse>('/tags', { name });
  return res.data.data;
}

export async function updateTag(id: string, name: string): Promise<Tag> {
  const res = await apiClient.put<TagResponse>(`/tags/${id}`, { name });
  return res.data.data;
}

export async function deleteTag(id: string): Promise<void> {
  await apiClient.delete(`/tags/${id}`);
}
