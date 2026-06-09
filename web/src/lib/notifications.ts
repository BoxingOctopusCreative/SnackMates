import { useCallback, useEffect, useRef, useState } from "react";
import { API_URL, api, getToken } from "@/lib/api";
import {
  findNewAcceptedNotifications,
  notificationKey,
  notifySnackMateAccepted,
} from "@/lib/notification-alerts";

export const NOTIFICATIONS_CHANGED = "snackmates:notifications-changed";

export type NotificationItem = Awaited<ReturnType<typeof api.notifications>>["items"][number];

export function notifyNotificationsChanged() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(NOTIFICATIONS_CHANGED));
  }
}

function notificationStreamUrl(token: string) {
  const params = new URLSearchParams({ access_token: token });
  return `${API_URL}/api/v1/notifications/stream?${params.toString()}`;
}

export function useNotifications(options?: { enabled?: boolean }) {
  const enabled = options?.enabled ?? true;

  const [items, setItems] = useState<NotificationItem[]>([]);
  const [loading, setLoading] = useState(false);
  const sourceRef = useRef<EventSource | null>(null);
  const seenRef = useRef<Set<string>>(new Set());
  const initializedRef = useRef(false);

  const load = useCallback(async () => {
    if (!enabled) return;
    try {
      const data = await api.notifications(getToken());
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

    function connectStream() {
      const token = getToken();
      if (!token) return;

      sourceRef.current?.close();
      const source = new EventSource(notificationStreamUrl(token));
      sourceRef.current = source;

      source.addEventListener("refresh", () => {
        notifyNotificationsChanged();
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
      window.removeEventListener(NOTIFICATIONS_CHANGED, onChange);
      document.removeEventListener("visibilitychange", onVisibilityChange);
      sourceRef.current?.close();
      sourceRef.current = null;
    };
  }, [enabled, load]);

  return { items, loading, load, refresh };
}
