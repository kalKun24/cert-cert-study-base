import apiClient from './apiClient';
import {
  Question,
  QuestionListData,
  QuestionListResponse,
  QuestionResponse,
  CreateQuestionRequest,
  UpdateQuestionRequest,
} from '../types/question';

export interface FetchQuestionsParams {
  page?: number;
  per_page?: number;
  keyword?: string;
  tag_ids?: string;
}

export async function fetchQuestions(
  teamId: string,
  params: FetchQuestionsParams
): Promise<QuestionListData> {
  const res = await apiClient.get<QuestionListResponse>(`/teams/${teamId}/questions`, { params });
  return res.data.data;
}

export async function fetchQuestion(teamId: string, id: string): Promise<Question> {
  const res = await apiClient.get<QuestionResponse>(`/teams/${teamId}/questions/${id}`);
  return res.data.data;
}

export async function createQuestion(
  teamId: string,
  req: CreateQuestionRequest
): Promise<Question> {
  const res = await apiClient.post<QuestionResponse>(`/teams/${teamId}/questions`, req);
  return res.data.data;
}

export async function updateQuestion(
  teamId: string,
  id: string,
  req: UpdateQuestionRequest
): Promise<Question> {
  const res = await apiClient.put<QuestionResponse>(`/teams/${teamId}/questions/${id}`, req);
  return res.data.data;
}

export async function deleteQuestion(teamId: string, id: string): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/questions/${id}`);
}
