export interface Tag {
  id: string;
  name: string;
  created_at: string;
}

export interface TagListResponse {
  data: Tag[];
  error: string | null;
}

export interface TagResponse {
  data: Tag;
  error: string | null;
}

export interface TagDeleteResponse {
  data: { message: string };
  error: string | null;
}
