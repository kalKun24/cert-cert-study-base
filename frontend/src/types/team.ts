export interface Team {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
}

export interface TeamMember {
  team_id: string;
  user_id: string;
  joined_at: string;
}

export interface TeamDetail extends Team {
  members: TeamMember[];
}

export interface TeamsResponse {
  data: Team[];
  error: string | null;
}

export interface TeamResponse {
  data: Team;
  error: string | null;
}

export interface TeamDetailResponse {
  data: TeamDetail;
  error: string | null;
}

export interface CreateTeamRequest {
  name: string;
  description: string;
}

export interface UpdateTeamRequest {
  name?: string;
  description?: string;
}

export interface AddMemberRequest {
  user_id: string;
}
