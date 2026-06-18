import apiClient from './apiClient';
import { Comment, CommentListResponse, CommentResponse } from '../types/comment';

export async function fetchComments(questionId: string): Promise<Comment[]> {
  const res = await apiClient.get<CommentListResponse>(`/questions/${questionId}/comments`);
  return res.data.data;
}

export async function postComment(questionId: string, body: string): Promise<Comment> {
  const res = await apiClient.post<CommentResponse>(`/questions/${questionId}/comments`, {
    body,
  });
  return res.data.data;
}

export async function updateComment(
  questionId: string,
  commentId: string,
  body: string
): Promise<Comment> {
  const res = await apiClient.put<CommentResponse>(
    `/questions/${questionId}/comments/${commentId}`,
    { body }
  );
  return res.data.data;
}

export async function deleteComment(questionId: string, commentId: string): Promise<void> {
  await apiClient.delete(`/questions/${questionId}/comments/${commentId}`);
}
