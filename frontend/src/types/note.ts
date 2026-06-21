export type NoteStatus = 'draft' | 'private' | 'published';

export interface Note {
  id: string;
  team_id: string;
  title: string;
  body: string;
  discussion_points: string;
  memo: string;
  tags: string[];
  status: NoteStatus;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface NoteListData {
  items: Note[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface NoteListResponse {
  data: NoteListData;
  error: string | null;
}

export interface NoteResponse {
  data: Note;
  error: string | null;
}

export interface CreateNoteRequest {
  title: string;
  body?: string;
  discussion_points?: string;
  memo?: string;
  tags?: string[];
  status?: NoteStatus;
}

export interface UpdateNoteRequest {
  title?: string;
  body?: string;
  discussion_points?: string;
  memo?: string;
  tags?: string[];
  status?: NoteStatus;
}
