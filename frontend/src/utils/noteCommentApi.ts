import apiClient from './apiClient';
import { NoteComment, NoteCommentListResponse, NoteCommentResponse } from '../types/noteComment';

export async function fetchNoteComments(
  teamId: string,
  noteId: string
): Promise<NoteComment[]> {
  const res = await apiClient.get<NoteCommentListResponse>(
    `/teams/${teamId}/notes/${noteId}/comments`
  );
  return res.data.data;
}

export async function postNoteComment(
  teamId: string,
  noteId: string,
  body: string
): Promise<NoteComment> {
  const res = await apiClient.post<NoteCommentResponse>(
    `/teams/${teamId}/notes/${noteId}/comments`,
    { body }
  );
  return res.data.data;
}

export async function updateNoteComment(
  teamId: string,
  noteId: string,
  commentId: string,
  body: string
): Promise<NoteComment> {
  const res = await apiClient.put<NoteCommentResponse>(
    `/teams/${teamId}/notes/${noteId}/comments/${commentId}`,
    { body }
  );
  return res.data.data;
}

export async function deleteNoteComment(
  teamId: string,
  noteId: string,
  commentId: string
): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/notes/${noteId}/comments/${commentId}`);
}
