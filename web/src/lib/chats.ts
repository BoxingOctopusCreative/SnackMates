import { useCallback, useEffect, useRef, useState } from "react";
import { api, Chat, ChatMessage, getToken } from "@/lib/api";
import { subscribeRealtimeStream } from "@/lib/realtime-stream";

export const CHATS_CHANGED = "snackmates:chats-changed";

export type ChatsChangedDetail = {
  /** When false, only the chat list refreshes (skip reloading the open thread). */
  refreshThread?: boolean;
};

export function notifyChatsChanged(detail?: ChatsChangedDetail) {
  if (typeof window !== "undefined") {
    window.dispatchEvent(
      new CustomEvent<ChatsChangedDetail>(CHATS_CHANGED, {
        detail: { refreshThread: detail?.refreshThread ?? true },
      }),
    );
  }
}

export function useChats(options?: { enabled?: boolean }) {
  const enabled = options?.enabled ?? true;

  const [chats, setChats] = useState<Chat[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const loadGenerationRef = useRef(0);

  const load = useCallback(async () => {
    if (!enabled) return;
    const generation = ++loadGenerationRef.current;
    try {
      const data = await api.chats(getToken());
      if (generation !== loadGenerationRef.current) return;
      setChats(data.chats);
      setUnreadCount(data.unread_count);
    } catch {
      if (generation !== loadGenerationRef.current) return;
      setChats([]);
      setUnreadCount(0);
    }
  }, [enabled]);

  const refresh = useCallback(async () => {
    if (!enabled) return;
    setLoading(true);
    try {
      await load();
    } finally {
      setLoading(false);
    }
  }, [enabled, load]);

  useEffect(() => {
    if (!enabled) return;

    load();

    const onChange = () => load();
    window.addEventListener(CHATS_CHANGED, onChange);
    const unsubscribe = subscribeRealtimeStream(notifyChatsChanged);

    function onVisibilityChange() {
      if (document.visibilityState === "visible") {
        load();
      }
    }
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      loadGenerationRef.current += 1;
      window.removeEventListener(CHATS_CHANGED, onChange);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      unsubscribe();
    };
  }, [enabled, load]);

  return { chats, unreadCount, loading, load, refresh };
}

export async function fetchChatMessages(chatId: string) {
  return api.getChat(chatId, getToken());
}

export async function postChatMessage(
  chatId: string,
  body: string,
  options?: { refreshThread?: boolean },
) {
  const msg = await api.sendChatMessage(chatId, body, getToken());
  notifyChatsChanged({ refreshThread: options?.refreshThread });
  return msg;
}

export async function openChatWith(username: string) {
  const { id } = await api.startChat(username, getToken());
  notifyChatsChanged({ refreshThread: false });
  return id;
}

export type { Chat, ChatMessage };
