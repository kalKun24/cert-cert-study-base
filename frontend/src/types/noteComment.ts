export interface NoteComment {
  id: string;
  note_id: string;
  body: string;
  created_by: string;
  display_name: string;
  created_at: string;
  updated_at: string;
}

export interface NoteCommentListResponse {
  data: NoteComment[];
  error: string | null;
}

export interface NoteCommentResponse {
  data: NoteComment;
  error: string | null;
}
