import React, { createContext, useContext, useEffect, useState } from 'react';
import { api, tokenStore } from './api';

export type UserInfo = {
  id: number;
  username: string;
  nickname?: string;
  email?: string;
  role?: string;
  avatar?: string;
};

export type AdminInfo = {
  id: number;
  username: string;
  nickname?: string;
  role?: string;
};

type AuthContextType = {
  user: UserInfo | null;
  admin: AdminInfo | null;
  loading: boolean;
  setUser: (u: UserInfo | null) => void;
  setAdmin: (a: AdminInfo | null) => void;
  loginUser: (username: string, password: string) => Promise<void>;
  loginAdmin: (username: string, password: string) => Promise<void>;
  logoutUser: () => Promise<void>;
  logoutAdmin: () => Promise<void>;
  refreshUser: () => Promise<void>;
};

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [admin, setAdmin] = useState<AdminInfo | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshUser = async () => {
    if (!tokenStore.getUser()) {
      setUser(null);
      return;
    }
    try {
      const res = await api.getProfile();
      setUser(res.data);
    } catch {
      tokenStore.clearUser();
      setUser(null);
    }
  };

  useEffect(() => {
    (async () => {
      await refreshUser();
      if (tokenStore.getAdmin()) {
        try {
          const res = await api.adminDashboard();
          if (res.code === 200 || res.code === 0) {
            setAdmin({ id: 0, username: 'admin' });
          }
        } catch {
          tokenStore.clearAdmin();
        }
      }
      setLoading(false);
    })();
  }, []);

  const loginUser = async (username: string, password: string) => {
    const res = await api.userLogin(username, password);
    tokenStore.setUser(res.data.token);
    setUser(res.data.user);
  };

  const loginAdmin = async (username: string, password: string) => {
    const res = await api.adminLogin(username, password);
    tokenStore.setAdmin(res.data.token);
    setAdmin(res.data.admin);
  };

  const logoutUser = async () => {
    try { await api.userLogout(); } catch {}
    tokenStore.clearUser();
    setUser(null);
  };

  const logoutAdmin = async () => {
    try { await api.adminLogout(); } catch {}
    tokenStore.clearAdmin();
    setAdmin(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user, admin, loading,
        setUser, setAdmin,
        loginUser, loginAdmin,
        logoutUser, logoutAdmin,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}