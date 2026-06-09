"use client";

import { useState } from "react";
import { Button, Flex } from "@adobe/react-spectrum";
import { api, ApiError, FriendshipView, getToken } from "@/lib/api";
import { notifyNotificationsChanged } from "@/lib/notifications";

type SnackMateActionButtonProps = {
  username: string;
  friendship?: FriendshipView;
  onChange?: (friendship?: FriendshipView) => void;
};

export function SnackMateActionButton({
  username,
  friendship,
  onChange,
}: SnackMateActionButtonProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function run(action: () => Promise<void>) {
    setLoading(true);
    setError("");
    try {
      await action();
      notifyNotificationsChanged();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  async function sendRequest() {
    await run(async () => {
      const res = await api.requestFriend(username, getToken());
      onChange?.({ id: res.id, status: "pending_outgoing" });
    });
  }

  async function acceptRequest() {
    if (!friendship?.id) return;
    await run(async () => {
      await api.acceptFriend(friendship.id!, getToken());
      onChange?.({ id: friendship.id, status: "friends" });
    });
  }

  async function declineRequest() {
    if (!friendship?.id) return;
    await run(async () => {
      await api.declineFriend(friendship.id!, getToken());
      onChange?.(undefined);
    });
  }

  async function removeMate() {
    if (!friendship?.id) return;
    await run(async () => {
      await api.removeFriend(friendship.id!, getToken());
      onChange?.(undefined);
    });
  }

  const status = friendship?.status ?? "none";

  return (
    <Flex direction="column" gap="size-100" alignItems="end">
      {(status === "none" || status === "declined") && (
        <Button variant="accent" onPress={sendRequest} isDisabled={loading}>
          {loading ? "Sending..." : "Request Snack Mate"}
        </Button>
      )}
      {status === "pending_outgoing" && (
        <Flex gap="size-100" wrap>
          <Button variant="secondary" isDisabled>
            Request Sent
          </Button>
          <Button variant="negative" onPress={removeMate} isDisabled={loading}>
            Cancel
          </Button>
        </Flex>
      )}
      {status === "pending_incoming" && (
        <Flex gap="size-100" wrap>
          <Button variant="accent" onPress={acceptRequest} isDisabled={loading}>
            Accept
          </Button>
          <Button variant="secondary" onPress={declineRequest} isDisabled={loading}>
            Decline
          </Button>
        </Flex>
      )}
      {status === "friends" && (
        <Flex gap="size-100" wrap>
          <Button variant="secondary" isDisabled>
            Snack Mates
          </Button>
          <Button variant="negative" onPress={removeMate} isDisabled={loading}>
            Remove
          </Button>
        </Flex>
      )}
      {error && (
        <span style={{ color: "var(--spectrum-global-color-red-600)", fontSize: "0.875rem" }}>
          {error}
        </span>
      )}
    </Flex>
  );
}
