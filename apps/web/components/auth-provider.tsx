"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

import { api, tokens, type Account } from "@/lib/api";

interface AuthContextValue {
  user: Account | null;
  roles: string[];
  permissions: string[];
  loading: boolean;
  login: (identifier: string, password: string) => Promise<void>;
  register: (input: { username: string; email: string; password: string; displayName?: string }) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
  setAccount: (user: Account) => void;
  hasPermission: (slug: string) => boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<Account | null>(null);
  const [roles, setRoles] = useState<string[]>([]);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

  const hydrate = useCallback(async () => {
    if (!tokens.access()) {
      setLoading(false);
      return;
    }
    try {
      const { data } = await api.me();
      setUser(data.user);
      setRoles(data.roles ?? []);
      setPermissions(data.permissions ?? []);
    } catch {
      tokens.clear();
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void hydrate();
  }, [hydrate]);

  const login = useCallback(async (identifier: string, password: string) => {
    const { data } = await api.login({ identifier, password });
    tokens.set(data.tokens.accessToken, data.tokens.refreshToken);
    setUser(data.user);
    setRoles(data.roles ?? []);
    setPermissions(data.permissions ?? []);
  }, []);

  const register = useCallback(
    async (input: { username: string; email: string; password: string; displayName?: string }) => {
      const { data } = await api.register(input);
      tokens.set(data.tokens.accessToken, data.tokens.refreshToken);
      setUser(data.user);
      setRoles(data.roles ?? []);
      setPermissions(data.permissions ?? []);
    },
    [],
  );

  const logout = useCallback(async () => {
    const refreshToken = tokens.refresh();
    if (refreshToken) {
      try {
        await api.logout(refreshToken);
      } catch {
        /* ignore */
      }
    }
    tokens.clear();
    setUser(null);
    setRoles([]);
    setPermissions([]);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      roles,
      permissions,
      loading,
      login,
      register,
      logout,
      refresh: hydrate,
      setAccount: setUser,
      hasPermission: (slug: string) => permissions.includes(slug),
    }),
    [user, roles, permissions, loading, login, register, logout, hydrate],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
