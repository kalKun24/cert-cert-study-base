export interface Question {
  id: string;
  title: string;
  body: string;
  answer: string;
  explanation: string;
  discussion_notes: string;
  tags: string[];
  created_by: string;
  display_name: string;
  is_visible_to: 'all' | 'members' | 'admin';
  created_at: string;
  updated_at: string;
}

export interface QuestionListResponse {
  data: Question[];
  error: string | null;
}

export interface QuestionResponse {
  data: Question;
  error: string | null;
}
