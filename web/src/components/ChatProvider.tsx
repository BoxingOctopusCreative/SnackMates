"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { useCurrentUserId } from "@/components/ConversationThread";
import {
  CHATS_CHANGED,
  ChatsChangedDetail,
  Chat,
  ChatMessage,
  fetchChatMessages,
  openChatWith,
  postChatMessage,
  useChats,
} from "@/lib/chats";

type ChatContextValue = {
  chats: Chat[];
  unreadCount: number;
  loading: boolean;
  refresh: () => Promise<void>;
  chatOpen: boolean;
  setChatOpen: (open: boolean) => void;
  activeChatId: string | null;
  activeMessages: ChatMessage[];
  activeOtherUser: Chat["other_user"] | null;
  openChat: (chatId: string) => Promise<void>;
  startChat: (username: string) => Promise<string>;
  sendChatMessage: (body: string) => Promise<void>;
  closeChat: () => void;
};

const ChatContext = createContext<ChatContextValue | null>(null);

export function ChatProvider({ children }: { children: React.ReactNode }) {
  const currentUserId = useCurrentUserId();
  const { chats, unreadCount, loading, refresh, load } = useChats();
  const [activeChatId, setActiveChatId] = useState<string | null>(null);
  const [activeMessages, setActiveMessages] = useState<ChatMessage[]>([]);
  const [activeOtherUser, setActiveOtherUser] = useState<Chat["other_user"] | null>(null);
  const [chatOpen, setChatOpen] = useState(false);

  const openChat = useCallback(
    async (chatId: string) => {
      const data = await fetchChatMessages(chatId);
      const fromList = chats.find((c) => c.id === chatId)?.other_user;
      setActiveChatId(chatId);
      setActiveMessages(data.messages);
      setActiveOtherUser(data.chat.other_user ?? fromList ?? null);
      await load();
    },
    [chats, load],
  );

  const startChat = useCallback(
    async (username: string) => {
      const id = await openChatWith(username);
      await openChat(id);
      return id;
    },
    [openChat],
  );

  const sendChatMessage = useCallback(
    async (body: string) => {
      if (!activeChatId) return;

      const optimisticId = `pending-${crypto.randomUUID()}`;
      const optimistic: ChatMessage = {
        id: optimisticId,
        chat_id: activeChatId,
        sender_id: currentUserId ?? "",
        body,
        created_at: new Date().toISOString(),
      };

      setActiveMessages((prev) => [...prev, optimistic]);

      try {
        const msg = await postChatMessage(activeChatId, body, { refreshThread: false });
        setActiveMessages((prev) => prev.map((m) => (m.id === optimisticId ? msg : m)));
      } catch (err) {
        setActiveMessages((prev) => prev.filter((m) => m.id !== optimisticId));
        throw err;
      }
    },
    [activeChatId, currentUserId],
  );

  const closeChat = useCallback(() => {
    setActiveChatId(null);
    setActiveMessages([]);
    setActiveOtherUser(null);
  }, []);

  useEffect(() => {
    if (!chatOpen || !activeChatId) return;

    const chatId = activeChatId;

    async function reloadActiveChat() {
      try {
        const data = await fetchChatMessages(chatId);
        const fromList = chats.find((c) => c.id === chatId)?.other_user;
        setActiveMessages(data.messages);
        setActiveOtherUser(data.chat.other_user ?? fromList ?? null);
      } catch {
        // keep showing the last loaded thread on transient errors
      }
    }

    function onChatsChanged(event: Event) {
      const detail = (event as CustomEvent<ChatsChangedDetail>).detail;
      if (detail?.refreshThread === false) return;
      reloadActiveChat();
    }

    window.addEventListener(CHATS_CHANGED, onChatsChanged);
    return () => window.removeEventListener(CHATS_CHANGED, onChatsChanged);
  }, [chatOpen, activeChatId, chats]);

  const value = useMemo(
    () => ({
      chats,
      unreadCount,
      loading,
      refresh,
      chatOpen,
      setChatOpen,
      activeChatId,
      activeMessages,
      activeOtherUser,
      openChat,
      startChat,
      sendChatMessage,
      closeChat,
    }),
    [
      chats,
      unreadCount,
      loading,
      refresh,
      chatOpen,
      activeChatId,
      activeMessages,
      activeOtherUser,
      openChat,
      startChat,
      sendChatMessage,
      closeChat,
    ],
  );

  return <ChatContext.Provider value={value}>{children}</ChatContext.Provider>;
}

export function useChatContext() {
  const ctx = useContext(ChatContext);
  if (!ctx) {
    throw new Error("useChatContext must be used within ChatProvider");
  }
  return ctx;
}
