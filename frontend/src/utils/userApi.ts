import apiClient from './apiClient';
import {
  User,
  UsersResponse,
  UserResponse,
  CreateUserRequest,
  UpdateUserRequest,
  UpdateUserStatusRequest,
  UpdateTeamOwnerStatusRequest,
} from '../types/user';

export async function fetchUsers(): Promise<User[]> {
  const res = await apiClient.get<UsersResponse>('/users');
  return res.data.data;
}

export async function fetchUser(id: string): Promise<User> {
  const res = await apiClient.get<UserResponse>(`/users/${id}`);
  return res.data.data;
}

export async function createUser(req: CreateUserRequest): Promise<User> {
  const res = await apiClient.post<UserResponse>('/users', req);
  return res.data.data;
}

export async function updateUser(id: string, req: UpdateUserRequest): Promise<User> {
  const res = await apiClient.put<UserResponse>(`/users/${id}`, req);
  return res.data.data;
}

export async function deleteUser(id: string): Promise<void> {
  await apiClient.delete(`/users/${id}`);
}

export async function updateUserStatus(id: string, req: UpdateUserStatusRequest): Promise<User> {
  const res = await apiClient.patch<UserResponse>(`/users/${id}/status`, req);
  return res.data.data;
}

export async function updateTeamOwnerStatus(
  id: string,
  req: UpdateTeamOwnerStatusRequest,
): Promise<User> {
  const res = await apiClient.patch<UserResponse>(`/admin/users/${id}/team-owner`, req);
  return res.data.data;
}
