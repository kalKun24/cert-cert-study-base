import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import {
  AuthUser,
  getToken,
  getStoredUser,
  setToken,
  storeUser,
  clearAuthStorage,
  isTokenExpired,
} from '../utils/auth';

interface AuthContextValue {
  user: AuthUser | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (user: AuthUser, token: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setTokenState] = useState<string | null>(null);

  useEffect(() => {
    const savedToken = getToken();
    if (savedToken && !isTokenExpired(savedToken)) {
      const savedUser = getStoredUser();
      if (savedUser) {
        setTokenState(savedToken);
        setUser(savedUser);
      } else {
        clearAuthStorage();
      }
    } else {
      clearAuthStorage();
    }
  }, []);

  const login = (newUser: AuthUser, newToken: string) => {
    setToken(newToken);
    storeUser(newUser);
    setTokenState(newToken);
    setUser(newUser);
  };

  const logout = () => {
    clearAuthStorage();
    setTokenState(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, token, isAuthenticated: !!token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth は AuthProvider の内側で使用してください');
  return ctx;
}
