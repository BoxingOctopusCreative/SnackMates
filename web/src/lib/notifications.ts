import { useCallback, useEffect, useRef, useState } from "react";
import { api, getToken } from "@/lib/api";
import {
  findNewAcceptedNotifications,
  notificationKey,
  notifySnackMateAccepted,
} from "@/lib/notification-alerts";
import { subscribeRealtimeStream } from "@/lib/realtime-stream";

export const NOTIFICATIONS_CHANGED = "snackmates:notifications-changed";

export type NotificationItem = Awaited<ReturnType<typeof api.notifications>>["items"][number];

export function notifyNotificationsChanged() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(NOTIFICATIONS_CHANGED));
  }
}

export function useNotifications(options?: { enabled?: boolean }) {
  const enabled = options?.enabled ?? true;

  const [items, setItems] = useState<NotificationItem[]>([]);
  const [loading, setLoading] = useState(false);
  const seenRef = useRef<Set<string>>(new Set());
  const initializedRef = useRef(false);
  const loadGenerationRef = useRef(0);

  const load = useCallback(async () => {
    if (!enabled) return;
    const generation = ++loadGenerationRef.current;
    try {
      const data = await api.notifications(getToken());
      if (generation !== loadGenerationRef.current) return;
      const nextItems = data.items;
      const accepted = findNewAcceptedNotifications(
        seenRef.current,
        nextItems,
        initializedRef.current,
      );
      for (const item of accepted) {
        const user = item.friendship.user;
        notifySnackMateAccepted(user?.display_name ?? "A snack mate", user?.username);
      }
      seenRef.current = new Set(nextItems.map(notificationKey));
      initializedRef.current = true;
      setItems(nextItems);
    } catch {
      if (generation !== loadGenerationRef.current) return;
      seenRef.current = new Set();
      initializedRef.current = true;
      setItems([]);
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
    window.addEventListener(NOTIFICATIONS_CHANGED, onChange);
    const unsubscribe = subscribeRealtimeStream(notifyNotificationsChanged);

    function onVisibilityChange() {
      if (document.visibilityState === "visible") {
        load();
      }
    }
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      loadGenerationRef.current += 1;
      window.removeEventListener(NOTIFICATIONS_CHANGED, onChange);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      unsubscribe();
    };
  }, [enabled, load]);

  return { items, loading, load, refresh };
}
