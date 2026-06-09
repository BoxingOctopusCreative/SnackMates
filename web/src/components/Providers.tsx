"use client";

import { ThemeProvider } from "@/components/ThemeProvider";
import { ConditionalThemeSwitcher } from "@/components/ConditionalThemeSwitcher";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider>
      <ConditionalThemeSwitcher />
      {children}
    </ThemeProvider>
  );
}
