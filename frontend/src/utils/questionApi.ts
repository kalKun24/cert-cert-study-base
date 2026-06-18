import apiClient from './apiClient';
import {
  Question,
  QuestionResponse,
  CreateQuestionRequest,
  UpdateQuestionRequest,
} from '../types/question';

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
