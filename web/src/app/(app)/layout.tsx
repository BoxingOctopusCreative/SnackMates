"use client";

import { AuthProvider } from "@/components/AuthGate";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>;
}
