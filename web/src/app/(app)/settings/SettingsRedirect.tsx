"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/AuthGate";
import { useSettingsModal } from "@/components/SettingsModalProvider";

export function SettingsRedirect() {
  const router = useRouter();
  const { user } = useAuth();
  const { openSettings } = useSettingsModal();

  useEffect(() => {
    openSettings();
    router.replace(`/users/${user.username}`);
  }, [openSettings, router, user.username]);

  return null;
}
