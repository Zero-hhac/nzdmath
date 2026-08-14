import React, { createContext, useContext, useEffect, useState } from 'react';
import { api, tokenStore } from './api';

export type UserInfo = {
  id: number;
  username: string;
  nickname?: string;
  email?: string;
  role?: string;
  avatar?: string;
  real_name?: string;
  class_name?: string;
  department?: string;
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
  isAdmin: boolean;
  loading: boolean;
  setUser: (u: UserInfo | null) => void;
  setAdmin: (a: AdminInfo | null) => void;
  loginUser: (username: string, password: string) => Promise<{ isAdmin: boolean; user: UserInfo }>;
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
      setAdmin(null);
      return;
    }
    try {
      const res = await api.getProfile();
      const u = res.data;
      setUser(u);
      if (u.role === 'admin' || u.role === 'super_admin' || u.username === 'admin') {
        setAdmin({ id: u.id, username: u.username, role: u.role });
      } else if (tokenStore.getAdmin()) {
        try {
          const dashRes = await api.adminDashboard();
          if (dashRes.code === 200 || dashRes.code === 0) {
            setAdmin({ id: u.id, username: u.username, role: 'admin' });
          }
        } catch {
          tokenStore.clearAdmin();
          setAdmin(null);
        }
      } else {
        setAdmin(null);
      }
    } catch {
      tokenStore.clearUser();
      tokenStore.clearAdmin();
      setUser(null);
      setAdmin(null);
    }
  };

  useEffect(() => {
    (async () => {
      await refreshUser();
      setLoading(false);
    })();
  }, []);

  const loginUser = async (username: string, password: string) => {
    const res = await api.userLogin(username, password);
    const data = res.data;
    tokenStore.setUser(data.token);
    setUser(data.user);

    const isAdm = Boolean(
      data.is_admin ||
      data.admin_token ||
      data.user?.role === 'admin' ||
      data.user?.role === 'super_admin' ||
      data.user?.username === 'admin'
    );

    if (isAdm) {
      if (data.admin_token) {
        tokenStore.setAdmin(data.admin_token);
      } else {
        tokenStore.setAdmin(data.token);
      }
      setAdmin({ id: data.user.id, username: data.user.username, role: data.user.role || 'admin' });
    } else {
      tokenStore.clearAdmin();
      setAdmin(null);
    }

    return { isAdmin: isAdm, user: data.user };
  };

  const loginAdmin = async (username: string, password: string) => {
    await loginUser(username, password);
  };

  const logoutUser = async () => {
    try { await api.userLogout(); } catch {}
    if (tokenStore.getAdmin()) {
      try { await api.adminLogout(); } catch {}
    }
    tokenStore.clearUser();
    tokenStore.clearAdmin();
    setUser(null);
    setAdmin(null);
  };

  const logoutAdmin = async () => {
    await logoutUser();
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        admin,
        isAdmin: Boolean(admin || user?.role === 'admin' || user?.role === 'super_admin' || user?.username === 'admin'),
        loading,
        setUser,
        setAdmin,
        loginUser,
        loginAdmin,
        logoutUser,
        logoutAdmin,
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
