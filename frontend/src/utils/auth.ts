import { UserRole } from '../types/user';

export type { UserRole };

const TOKEN_KEY = 'cert_study_token';
const USER_KEY = 'cert_study_user';

export interface AuthUser {
  id: string;
  username: string;
  display_name: string;
  email: string;
  role: UserRole;
  is_active: boolean;
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token);
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY);
}

export function getStoredUser(): AuthUser | null {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthUser;
  } catch {
    return null;
  }
}

export function storeUser(user: AuthUser): void {
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function removeUser(): void {
  localStorage.removeItem(USER_KEY);
}

export function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.exp * 1000 < Date.now();
  } catch {
    return true;
  }
}

export function clearAuthStorage(): void {
  removeToken();
  removeUser();
}
