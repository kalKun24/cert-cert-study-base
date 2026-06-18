export interface Tag {
  id: string;
  name: string;
  created_at: string;
}

export interface TagListResponse {
  data: Tag[];
  error: string | null;
}
