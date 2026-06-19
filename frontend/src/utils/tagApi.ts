import apiClient from './apiClient';
import { Tag, TagListResponse } from '../types/tag';

export async function fetchTags(): Promise<Tag[]> {
  const res = await apiClient.get<TagListResponse>('/tags');
  return res.data.data;
}
