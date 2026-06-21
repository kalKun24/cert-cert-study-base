import apiClient from './apiClient';
import {
  Note,
  NoteListData,
  NoteListResponse,
  NoteResponse,
  CreateNoteRequest,
  UpdateNoteRequest,
  NoteStatus,
} from '../types/note';

export interface FetchNotesParams {
  page?: number;
  per_page?: number;
  keyword?: string;
  tag_ids?: string;
}

export async function fetchNotes(
  teamId: string,
  params: FetchNotesParams
): Promise<NoteListData> {
  const res = await apiClient.get<NoteListResponse>(`/teams/${teamId}/notes`, { params });
  return res.data.data;
}

export async function fetchNote(teamId: string, id: string): Promise<Note> {
  const res = await apiClient.get<NoteResponse>(`/teams/${teamId}/notes/${id}`);
  return res.data.data;
}

export async function createNote(
  teamId: string,
  req: CreateNoteRequest
): Promise<Note> {
  const res = await apiClient.post<NoteResponse>(`/teams/${teamId}/notes`, req);
  return res.data.data;
}

export async function updateNote(
  teamId: string,
  id: string,
  req: UpdateNoteRequest
): Promise<Note> {
  const res = await apiClient.put<NoteResponse>(`/teams/${teamId}/notes/${id}`, req);
  return res.data.data;
}

export async function deleteNote(teamId: string, id: string): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/notes/${id}`);
}

export async function updateNoteVisibility(
  teamId: string,
  id: string,
  status: NoteStatus
): Promise<Note> {
  const res = await apiClient.patch<NoteResponse>(
    `/teams/${teamId}/notes/${id}/visibility`,
    { status }
  );
  return res.data.data;
}
