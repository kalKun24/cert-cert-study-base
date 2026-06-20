import apiClient from './apiClient';
import { Comment, CommentListResponse, CommentResponse } from '../types/comment';

export async function fetchComments(teamId: string, questionId: string): Promise<Comment[]> {
  const res = await apiClient.get<CommentListResponse>(
    `/teams/${teamId}/questions/${questionId}/comments`
  );
  return res.data.data;
}

export async function postComment(
  teamId: string,
  questionId: string,
  body: string
): Promise<Comment> {
  const res = await apiClient.post<CommentResponse>(
    `/teams/${teamId}/questions/${questionId}/comments`,
    { body }
  );
  return res.data.data;
}

export async function updateComment(
  teamId: string,
  questionId: string,
  commentId: string,
  body: string
): Promise<Comment> {
  const res = await apiClient.put<CommentResponse>(
    `/teams/${teamId}/questions/${questionId}/comments/${commentId}`,
    { body }
  );
  return res.data.data;
}

export async function deleteComment(
  teamId: string,
  questionId: string,
  commentId: string
): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/questions/${questionId}/comments/${commentId}`);
}
