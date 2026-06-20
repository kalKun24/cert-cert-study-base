export type QuestionStatus = 'draft' | 'private' | 'published';

export interface Question {
  id: string;
  team_id: string;
  title: string;
  body: string;
  answer: string;
  explanation: string;
  memo: string;
  tags: string[];
  status: QuestionStatus;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface QuestionListData {
  items: Question[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface QuestionListResponse {
  data: QuestionListData;
  error: string | null;
}

export interface QuestionResponse {
  data: Question;
  error: string | null;
}

export interface CreateQuestionRequest {
  title: string;
  body: string;
  answer?: string;
  explanation?: string;
  memo?: string;
  tags?: string[];
  status?: QuestionStatus;
}

export interface UpdateQuestionRequest {
  title?: string;
  body?: string;
  answer?: string;
  explanation?: string;
  memo?: string;
  tags?: string[];
  status?: QuestionStatus;
}
