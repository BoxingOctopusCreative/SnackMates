"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { restoreSessionFromCookie } from "@/lib/session";

export function useExistingSessionRedirect() {
  const router = useRouter();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    let cancelled = false;

    restoreSessionFromCookie().then((token) => {
      if (cancelled) return;
      if (token) {
        router.replace("/dashboard");
        return;
      }
      setChecking(false);
    });

    return () => {
      cancelled = true;
    };
  }, [router]);

  return checking;
}
