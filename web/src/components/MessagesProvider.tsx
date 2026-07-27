"use client";

import { createContext, useContext, useMemo } from "react";
import { Conversation, useMessages } from "@/lib/messages";

type MessagesContextValue = {
  conversations: Conversation[];
  unreadCount: number;
  loading: boolean;
  refresh: () => Promise<void>;
  load: () => Promise<void>;
};

const MessagesContext = createContext<MessagesContextValue | null>(null);

/** Shared inbox state + single realtime subscription for the mail icon and /messages. */
export function MessagesProvider({ children }: { children: React.ReactNode }) {
  const { conversations, unreadCount, loading, refresh, load } = useMessages();

  const value = useMemo(
    () => ({
      conversations,
      unreadCount,
      loading,
      refresh,
      load,
    }),
    [conversations, unreadCount, loading, refresh, load],
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
