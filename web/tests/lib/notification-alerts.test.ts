import { describe, expect, it, vi } from "vitest";
import {
  findNewAcceptedNotifications,
  notificationKey,
  notifySnackMateAccepted,
} from "@/lib/notification-alerts";
import type { NotificationItem } from "@/lib/notifications";

vi.mock("@react-spectrum/toast", () => ({
  ToastQueue: {
    positive: vi.fn(),
  },
}));

const acceptedItem: NotificationItem = {
  id: "f1",
  type: "snack_mate_accepted",
  created_at: "2026-01-01T00:00:00Z",
  friendship: {
    id: "f1",
    requester_id: "u1",
    addressee_id: "u2",
    status: "accepted",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-02T00:00:00Z",
    user: {
      id: "u2",
      username: "bruno",
      display_name: "Bruno",
    },
  },
};

describe("notification alerts", () => {
  it("builds stable notification keys", () => {
    expect(notificationKey(acceptedItem)).toBe("snack_mate_accepted:f1");
  });

  it("does not alert before initialization", () => {
    expect(findNewAcceptedNotifications(new Set(), [acceptedItem], false)).toEqual([]);
  });

  it("alerts only for newly seen acceptances", () => {
    const seen = new Set(["snack_mate_accepted:f1"]);
    const next = [
      acceptedItem,
      {
        ...acceptedItem,
        id: "f2",
        friendship: { ...acceptedItem.friendship, id: "f2" },
      },
    ];

    expect(findNewAcceptedNotifications(seen, next, true)).toEqual([next[1]]);
  });

  it("queues a positive toast with profile action", async () => {
    const { ToastQueue } = await import("@react-spectrum/toast");
    notifySnackMateAccepted("Bruno", "bruno");

    expect(ToastQueue.positive).toHaveBeenCalledWith(
      "Bruno accepted your snack mate request!",
      expect.objectContaining({
        actionLabel: "View profile",
        shouldCloseOnAction: true,
      }),
    );
  });
});
