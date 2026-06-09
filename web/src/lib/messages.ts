import { useCallback, useEffect, useRef, useState } from "react";
import { API_URL, api, Conversation, getToken, Message } from "@/lib/api";

export const MESSAGES_CHANGED = "snackmates:messages-changed";

export function notifyMessagesChanged() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(MESSAGES_CHANGED));
  }
}

function messageStreamUrl(token: string) {
  const params = new URLSearchParams({ access_token: token });
  return `${API_URL}/api/v1/notifications/stream?${params.toString()}`;
}

export function useMessages(options?: { enabled?: boolean }) {
  const enabled = options?.enabled ?? true;

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const sourceRef = useRef<EventSource | null>(null);

  const load = useCallback(async () => {
    if (!enabled) return;
    try {
      const data = await api.conversations(getToken());
      setConversations(data.conversations);
      setUnreadCount(data.unread_count);
    } catch {
      setConversations([]);
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
    window.addEventListener(MESSAGES_CHANGED, onChange);

    function connectStream() {
      const token = getToken();
      if (!token) return;

      sourceRef.current?.close();
      const source = new EventSource(messageStreamUrl(token));
      sourceRef.current = source;

      source.addEventListener("refresh", () => {
        notifyMessagesChanged();
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
      window.removeEventListener(MESSAGES_CHANGED, onChange);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      sourceRef.current?.close();
      sourceRef.current = null;
    };
  }, [enabled, load]);

  return { conversations, unreadCount, loading, load, refresh };
}

export async function fetchConversationMessages(conversationId: string) {
  return api.getConversation(conversationId, getToken());
}

export async function postMessage(conversationId: string, subject: string, body: string) {
  const msg = await api.sendMessage(conversationId, subject, body, getToken());
  notifyMessagesChanged();
  return msg;
}

export async function openConversationWith(username: string) {
  const { id } = await api.startConversation(username, getToken());
  notifyMessagesChanged();
  return id;
}

export type { Conversation, Message };
