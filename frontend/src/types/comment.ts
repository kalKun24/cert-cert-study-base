export interface Comment {
  id: string;
  question_id: string;
  body: string;
  created_by: string;
  display_name: string;
  created_at: string;
  updated_at: string;
}

export interface CommentListResponse {
  data: Comment[];
  error: string | null;
}

export interface CommentResponse {
  data: Comment;
  error: string | null;
}
