import { useCallback, useEffect, useRef, useState } from "react";
import { api, Conversation, getToken, Message } from "@/lib/api";
import { subscribeRealtimeStream } from "@/lib/realtime-stream";

export const MESSAGES_CHANGED = "snackmates:messages-changed";

export function notifyMessagesChanged() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(MESSAGES_CHANGED));
  }
}

export function useMessages(options?: { enabled?: boolean }) {
  const enabled = options?.enabled ?? true;

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const loadGenerationRef = useRef(0);

  const load = useCallback(async () => {
    if (!enabled) return;
    const generation = ++loadGenerationRef.current;
    try {
      const data = await api.conversations(getToken());
      if (generation !== loadGenerationRef.current) return;
      setConversations(data.conversations);
      setUnreadCount(data.unread_count);
    } catch {
      if (generation !== loadGenerationRef.current) return;
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
    const unsubscribe = subscribeRealtimeStream(notifyMessagesChanged);

    function onVisibilityChange() {
      if (document.visibilityState === "visible") {
        load();
      }
    }
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      loadGenerationRef.current += 1;
      window.removeEventListener(MESSAGES_CHANGED, onChange);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      unsubscribe();
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
