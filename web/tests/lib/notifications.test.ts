import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { api } from "@/lib/api";
import {
  NOTIFICATIONS_CHANGED,
  notifyNotificationsChanged,
  useNotifications,
} from "@/lib/notifications";

vi.mock("@/lib/api", () => ({
  API_URL: "http://localhost:8080",
  api: {
    notifications: vi.fn(),
  },
  getToken: vi.fn(() => "token"),
}));

vi.mock("@/lib/notification-alerts", () => ({
  findNewAcceptedNotifications: vi.fn(() => []),
  notificationKey: (item: { type: string; id: string }) => `${item.type}:${item.id}`,
  notifySnackMateAccepted: vi.fn(),
}));

const sampleNotifications = {
  unread_count: 1,
  items: [
    {
      id: "n1",
      type: "snack_mate_request",
      created_at: "2026-01-01T00:00:00Z",
      friendship: {
        id: "f1",
        status: "pending",
        user: {
          id: "u1",
          username: "alice",
          display_name: "Alice",
        },
      },
    },
  ],
};

class MockEventSource {
  static instances: MockEventSource[] = [];
  url: string;
  readyState = 0;
  listeners = new Map<string, Set<() => void>>();

  constructor(url: string) {
    this.url = url;
    MockEventSource.instances.push(this);
  }

  addEventListener(type: string, listener: () => void) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(listener);
  }

  close() {
    this.readyState = 2;
  }

  emit(type: string) {
    for (const listener of this.listeners.get(type) ?? []) {
      listener();
    }
  }
}

describe("useNotifications", () => {
  beforeEach(() => {
    MockEventSource.instances = [];
    vi.stubGlobal("EventSource", MockEventSource);
    vi.mocked(api.notifications).mockResolvedValue(sampleNotifications);
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.unstubAllGlobals();
  });

  it("loads notifications on mount and opens the push stream", async () => {
    const { result } = renderHook(() => useNotifications());

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1);
    });

    expect(api.notifications).toHaveBeenCalledTimes(1);
    expect(MockEventSource.instances).toHaveLength(1);
    expect(MockEventSource.instances[0]?.url).toContain("/api/v1/notifications/stream");
    expect(MockEventSource.instances[0]?.url).toContain("access_token=token");
  });

  it("reloads when the notifications changed event fires", async () => {
    const { result } = renderHook(() => useNotifications());

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1);
    });

    act(() => {
      notifyNotificationsChanged();
    });

    await waitFor(() => {
      expect(api.notifications).toHaveBeenCalledTimes(2);
    });
  });

  it("reloads when the server pushes a refresh event", async () => {
    const { result } = renderHook(() => useNotifications());

    await waitFor(() => {
      expect(result.current.items).toHaveLength(1);
    });

    const source = MockEventSource.instances[0];
    act(() => {
      source?.emit("refresh");
    });

    await waitFor(() => {
      expect(api.notifications).toHaveBeenCalledTimes(2);
    });
  });

  it("dispatches the notifications changed event", () => {
    const listener = vi.fn();
    window.addEventListener(NOTIFICATIONS_CHANGED, listener);

    notifyNotificationsChanged();

    expect(listener).toHaveBeenCalledTimes(1);
    window.removeEventListener(NOTIFICATIONS_CHANGED, listener);
  });
});
