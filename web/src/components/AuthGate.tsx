"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, getToken, saveToken, User } from "@/lib/api";
import { AppShell } from "@/components/AppShell";
import { ProgressCircle, View } from "@adobe/react-spectrum";

function consumeTokenFromURL() {
  if (typeof window === "undefined") return;
  const params = new URLSearchParams(window.location.search);
  const urlToken = params.get("token");
  if (!urlToken) return;

  saveToken(urlToken);
  params.delete("token");
  const query = params.toString();
  const nextUrl = query ? `${window.location.pathname}?${query}` : window.location.pathname;
  window.history.replaceState({}, "", nextUrl);
}

type AuthContextValue = {
  user: User;
  updateUser: (patch: Partial<User>) => void;
  refreshUser: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    const next = await api.me(getToken());
    setUser(next);
  }, []);

  const updateUser = useCallback((patch: Partial<User>) => {
    setUser((current) => (current ? { ...current, ...patch } : current));
  }, []);

  useEffect(() => {
    consumeTokenFromURL();

    api
      .me(getToken())
      .then(setUser)
      .catch(() => {
        router.replace("/login");
      })
      .finally(() => setLoading(false));
  }, [router]);

  if (loading) {
    return (
      <View minHeight="100vh" UNSAFE_style={{ display: "grid", placeItems: "center" }}>
        <ProgressCircle isIndeterminate aria-label="Loading" />
      </View>
    );
  }

  if (!user) return null;

  return (
    <AuthContext.Provider value={{ user, updateUser, refreshUser }}>
      <AppShell user={user}>{children}</AppShell>
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return context;
}
