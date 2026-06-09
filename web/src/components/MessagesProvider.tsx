"use client";

import { createContext, useContext, useMemo } from "react";
import { useMessages } from "@/lib/messages";

type MessagesContextValue = {
  unreadCount: number;
  loading: boolean;
  refresh: () => Promise<void>;
};

const MessagesContext = createContext<MessagesContextValue | null>(null);

/** Inbox unread badge for the header mail icon only. */
export function MessagesProvider({ children }: { children: React.ReactNode }) {
  const { unreadCount, loading, refresh } = useMessages();

  const value = useMemo(
    () => ({
      unreadCount,
      loading,
      refresh,
    }),
    [unreadCount, loading, refresh],
  );

  return <MessagesContext.Provider value={value}>{children}</MessagesContext.Provider>;
}

export function useMessagesInbox() {
  const ctx = useContext(MessagesContext);
  if (!ctx) {
    throw new Error("useMessagesInbox must be used within MessagesProvider");
  }
  return ctx;
}
