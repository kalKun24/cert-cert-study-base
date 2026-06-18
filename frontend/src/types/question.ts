export type QuestionStatus = 'draft' | 'private' | 'published';
export type VisibilityScope = 'all' | 'team';

export interface Question {
  id: string;
  title: string;
  body: string;
  answer: string;
  explanation: string;
  memo: string;
  tags: string[];
  status: QuestionStatus;
  visibility_scope: VisibilityScope;
  published_team_ids: string[];
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
  visibility_scope?: VisibilityScope;
  published_team_ids?: string[];
}

export interface UpdateQuestionRequest {
  title?: string;
  body?: string;
  answer?: string;
  explanation?: string;
  memo?: string;
  tags?: string[];
  status?: QuestionStatus;
  visibility_scope?: VisibilityScope;
  published_team_ids?: string[];
}
