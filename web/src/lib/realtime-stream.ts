import { API_URL, getToken } from "@/lib/api";

type RefreshHandler = () => void;

let source: EventSource | null = null;
let subscriberCount = 0;
let visibilityBound = false;
const refreshHandlers = new Set<RefreshHandler>();

function streamUrl(token: string) {
  const params = new URLSearchParams({ access_token: token });
  return `${API_URL}/api/v1/notifications/stream?${params.toString()}`;
}

function connect() {
  const token = getToken();
  if (!token || typeof EventSource === "undefined") return;

  source?.close();
  const next = new EventSource(streamUrl(token));
  source = next;

  next.addEventListener("refresh", () => {
    for (const handler of refreshHandlers) {
      handler();
    }
  });
}

function onVisibilityChange() {
  if (document.visibilityState !== "visible") return;
  if (!source || source.readyState === EventSource.CLOSED) {
    connect();
  }
}

function ensureConnected() {
  if (subscriberCount === 0) return;
  if (!source || source.readyState === EventSource.CLOSED) {
    connect();
  }
  if (!visibilityBound) {
    document.addEventListener("visibilitychange", onVisibilityChange);
    visibilityBound = true;
  }
}

function maybeDisconnect() {
  if (subscriberCount > 0) return;
  source?.close();
  source = null;
  if (visibilityBound) {
    document.removeEventListener("visibilitychange", onVisibilityChange);
    visibilityBound = false;
  }
}

/**
 * Share one EventSource for /notifications/stream across notifications,
 * messages, and chats. Pass an onRefresh callback that reloads that domain
 * (typically by dispatching its *CHANGED event).
 */
export function subscribeRealtimeStream(onRefresh: RefreshHandler): () => void {
  refreshHandlers.add(onRefresh);
  subscriberCount += 1;
  ensureConnected();

  return () => {
    refreshHandlers.delete(onRefresh);
    subscriberCount = Math.max(0, subscriberCount - 1);
    maybeDisconnect();
  };
}

/** Test helper — current open EventSource count (0 or 1). */
export function realtimeStreamConnectionCountForTests() {
  return source ? 1 : 0;
}

/** Test helper — reset module state between tests. */
export function resetRealtimeStreamForTests() {
  source?.close();
  source = null;
  subscriberCount = 0;
  refreshHandlers.clear();
  if (visibilityBound) {
    document.removeEventListener("visibilitychange", onVisibilityChange);
    visibilityBound = false;
  }
}
