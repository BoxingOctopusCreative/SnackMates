import { useCallback, useEffect, useRef, useState } from "react";
import { API_URL, api, Chat, ChatMessage, getToken } from "@/lib/api";

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

function chatStreamUrl(token: string) {
  const params = new URLSearchParams({ access_token: token });
  return `${API_URL}/api/v1/notifications/stream?${params.toString()}`;
}

export function useChats(options?: { enabled?: boolean }) {
  const enabled = options?.enabled ?? true;

  const [chats, setChats] = useState<Chat[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const sourceRef = useRef<EventSource | null>(null);

  const load = useCallback(async () => {
    if (!enabled) return;
    try {
      const data = await api.chats(getToken());
      setChats(data.chats);
      setUnreadCount(data.unread_count);
    } catch {
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

    function connectStream() {
      const token = getToken();
      if (!token) return;

      sourceRef.current?.close();
      const source = new EventSource(chatStreamUrl(token));
      sourceRef.current = source;

      source.addEventListener("refresh", () => {
        notifyChatsChanged();
      });
    }

    connectStream();

    function onVisibilityChange() {
      if (document.visibilityState !== "visible") return;
      load();
      if (!sourceRef.current || sourceRef.current.readyState === EventSource.CLOSED) {
        connectStream();
      }
    }
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      window.removeEventListener(CHATS_CHANGED, onChange);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      sourceRef.current?.close();
      sourceRef.current = null;
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
