"use client";

import { Avatar, Flex, Text, View } from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { Chat } from "@/lib/api";

type ChatListProps = {
  chats: Chat[];
  activeId?: string | null;
  onSelect: (chatId: string) => void;
};

export function ChatList({ chats, activeId, onSelect }: ChatListProps) {
  if (chats.length === 0) {
    return (
      <View padding="size-200">
        <Text>No live chats yet. Open a chat with a snack mate online.</Text>
      </View>
    );
  }

  return (
    <View UNSAFE_className="sm-conversation-list">
      {chats.map((chat) => {
        const user = chat.other_user;
        const name = user?.display_name || user?.username || "Snack Mate";
        const preview = chat.last_message?.body ?? "No messages yet";
        const isActive = chat.id === activeId;

        return (
          <button
            key={chat.id}
            type="button"
            className={`sm-conversation-list__item${isActive ? " sm-conversation-list__item--active" : ""}`}
            onClick={() => onSelect(chat.id)}
          >
            <Flex alignItems="center" gap="size-150" width="100%">
              <Avatar src={avatarImageSrc(user?.avatar_url)} alt="" size="avatar-size-300" />
              <Flex direction="column" flex={1} minWidth={0}>
                <Flex justifyContent="space-between" alignItems="center" width="100%">
                  <Text UNSAFE_style={{ fontWeight: 600 }}>{name}</Text>
                  {chat.unread_count > 0 && (
                    <span className="sm-conversation-list__unread">{chat.unread_count}</span>
                  )}
                </Flex>
                <Text UNSAFE_className="sm-conversation-list__preview">{preview}</Text>
              </Flex>
            </Flex>
          </button>
        );
      })}
    </View>
  );
}
