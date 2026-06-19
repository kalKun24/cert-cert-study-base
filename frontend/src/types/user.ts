export type UserRole = 'admin' | 'teamowner' | 'user';

export interface User {
  id: string;
  username: string;
  display_name: string;
  email: string;
  role: UserRole;
  is_active: boolean;
  is_team_owner?: boolean;
  max_teams?: number;
  created_at: string;
  updated_at: string;
}

export interface UsersResponse {
  data: User[];
  error: string | null;
}

export interface UserResponse {
  data: User;
  error: string | null;
}

export interface DeleteUserResponse {
  data: { message: string };
  error: string | null;
}

export interface CreateUserRequest {
  username: string;
  display_name: string;
  email: string;
  password: string;
  role: UserRole;
}

export interface UpdateUserRequest {
  display_name?: string;
  email?: string;
  role?: UserRole;
  password?: string;
}

export interface UpdateUserStatusRequest {
  is_active: boolean;
}

export interface UpdateTeamOwnerStatusRequest {
  is_team_owner: boolean;
  max_teams?: number;
}
