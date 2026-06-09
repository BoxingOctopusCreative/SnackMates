"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { SettingsModal } from "@/components/SettingsModal";

type SettingsModalContextValue = {
  openSettings: () => void;
  closeSettings: () => void;
};

const SettingsModalContext = createContext<SettingsModalContextValue | null>(null);

export function SettingsModalProvider({ children }: { children: ReactNode }) {
  const [isOpen, setIsOpen] = useState(false);
  const openSettings = useCallback(() => setIsOpen(true), []);
  const closeSettings = useCallback(() => setIsOpen(false), []);

  const value = useMemo<SettingsModalContextValue>(
    () => ({ openSettings, closeSettings }),
    [openSettings, closeSettings],
  );

  return (
    <SettingsModalContext.Provider value={value}>
      {children}
      <SettingsModal isOpen={isOpen} onClose={() => setIsOpen(false)} />
    </SettingsModalContext.Provider>
  );
}

export function useSettingsModal() {
  const context = useContext(SettingsModalContext);
  if (!context) {
    throw new Error("useSettingsModal must be used within SettingsModalProvider");
  }
  return context;
}
