"use client";

import { useState } from "react";
import Link from "next/link";
import { Icon } from "@iconify/react";
import {
  ActionButton,
  Avatar,
  Button,
  Flex,
  Text,
  View,
} from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { api, getToken } from "@/lib/api";
import { HeaderDropdown } from "@/components/HeaderDropdown";
import { notifyNotificationsChanged, useNotifications } from "@/lib/notifications";

export function NotificationMenu() {
  const { items, loading, load, refresh } = useNotifications();
  const [actionId, setActionId] = useState<string | null>(null);

  async function accept(friendshipId: string) {
    setActionId(friendshipId);
    try {
      await api.acceptFriend(friendshipId, getToken());
      await load();
      notifyNotificationsChanged();
    } finally {
      setActionId(null);
    }
  }

  async function decline(friendshipId: string) {
    setActionId(friendshipId);
    try {
      await api.declineFriend(friendshipId, getToken());
      await load();
      notifyNotificationsChanged();
    } finally {
      setActionId(null);
    }
  }

  async function acknowledge(friendshipId: string) {
    setActionId(friendshipId);
    try {
      await api.acknowledgeNotification(friendshipId, getToken());
      await load();
      notifyNotificationsChanged();
    } finally {
      setActionId(null);
    }
  }

  const unreadCount = items.length;

  return (
    <HeaderDropdown
      title="Notifications"
      onOpen={refresh}
      bodyClassName="sm-notification-dropdown__body"
      rootClassName="sm-notification-root"
      renderTrigger={({ open, toggle, panelId }) => (
        <ActionButton
          isQuiet
          aria-label={unreadCount > 0 ? `${unreadCount} notifications` : "Notifications"}
          aria-expanded={open}
          aria-controls={open ? panelId : undefined}
          UNSAFE_className={`sm-header-icon-link sm-notification-trigger${open ? " sm-notification-trigger--open" : ""}`}
          onPress={toggle}
        >
          <Icon
            icon="ion:notifications-outline"
            className="sm-header-icon sm-header-icon--default"
            aria-hidden
          />
          <Icon icon="ion:notifications" className="sm-header-icon sm-header-icon--hover" aria-hidden />
          {unreadCount > 0 && (
            <span className="sm-notification-badge" aria-hidden>
              {unreadCount > 9 ? "9+" : unreadCount}
            </span>
          )}
        </ActionButton>
      )}
    >
      {loading && items.length === 0 ? (
        <Text>Loading...</Text>
      ) : items.length === 0 ? (
        <Text>No new notifications.</Text>
      ) : (
        <Flex direction="column" gap="size-0">
          {items.map((item) => {
            const user = item.friendship.user;
            if (!user) return null;
            const busy = actionId === item.friendship.id;

            return (
              <View key={`${item.type}-${item.id}`} paddingY="size-100" UNSAFE_className="sm-notification-item">
                <Flex gap="size-150" alignItems="start">
                  <Avatar
                    src={avatarImageSrc(user.avatar_url)}
                    alt={user.display_name}
                    size="avatar-size-300"
                  />
                  <Flex direction="column" gap="size-75" flex={1}>
                    <Text>
                      <Link
                        href={`/users/${user.username}`}
                        style={{ color: "inherit", fontWeight: 600, textDecoration: "none" }}
                        onClick={() => {
                          if (item.type === "snack_mate_accepted") {
                            void acknowledge(item.friendship.id);
                          }
                        }}
                      >
                        {user.display_name}
                      </Link>{" "}
                      {item.type === "snack_mate_accepted"
                        ? "accepted your snack mate request."
                        : "sent you a snack mate request."}
                    </Text>
                    {item.type === "snack_mate_request" ? (
                      <Flex gap="size-100" wrap>
                        <Button
                          variant="accent"
                          onPress={() => accept(item.friendship.id)}
                          isDisabled={busy}
                        >
                          Accept
                        </Button>
                        <Button
                          variant="secondary"
                          onPress={() => decline(item.friendship.id)}
                          isDisabled={busy}
                        >
                          Decline
                        </Button>
                      </Flex>
                    ) : (
                      <Flex gap="size-100" wrap>
                        <Button
                          variant="accent"
                          onPress={() => acknowledge(item.friendship.id)}
                          isDisabled={busy}
                        >
                          Got It
                        </Button>
                      </Flex>
                    )}
                  </Flex>
                </Flex>
              </View>
            );
          })}
        </Flex>
      )}
    </HeaderDropdown>
  );
}
