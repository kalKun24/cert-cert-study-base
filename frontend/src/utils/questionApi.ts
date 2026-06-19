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

export async function fetchQuestions(params: FetchQuestionsParams): Promise<QuestionListData> {
  const res = await apiClient.get<QuestionListResponse>('/questions', { params });
  return res.data.data;
}

export async function fetchQuestion(id: string): Promise<Question> {
  const res = await apiClient.get<QuestionResponse>(`/questions/${id}`);
  return res.data.data;
}

export async function createQuestion(req: CreateQuestionRequest): Promise<Question> {
  const res = await apiClient.post<QuestionResponse>('/questions', req);
  return res.data.data;
}

export async function updateQuestion(id: string, req: UpdateQuestionRequest): Promise<Question> {
  const res = await apiClient.put<QuestionResponse>(`/questions/${id}`, req);
  return res.data.data;
}

export async function deleteQuestion(id: string): Promise<void> {
  await apiClient.delete(`/questions/${id}`);
}
