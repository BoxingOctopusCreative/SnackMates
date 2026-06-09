"use client";

import { usePathname } from "next/navigation";
import { ThemeSwitcher } from "@/components/ThemeSwitcher";

/** Public utility pages without AppShell / profile menu — only these get a floating theme toggle. */
const FLOATING_THEME_SWITCHER_ROUTES = [
  "/forgot-password",
  "/reset-password",
  "/verify-email",
  "/confirm-account",
] as const;

export function ConditionalThemeSwitcher() {
  const pathname = usePathname();
  const showFloatingSwitcher = FLOATING_THEME_SWITCHER_ROUTES.includes(
    pathname as (typeof FLOATING_THEME_SWITCHER_ROUTES)[number],
  );

  if (!showFloatingSwitcher) return null;

  return <ThemeSwitcher variant="floating" />;
}
