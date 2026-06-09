"use client";

import { Avatar, Flex, Text, View } from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { Conversation } from "@/lib/api";

type ConversationListProps = {
  conversations: Conversation[];
  activeId?: string | null;
  onSelect: (conversationId: string) => void;
  compact?: boolean;
};

export function ConversationList({
  conversations,
  activeId,
  onSelect,
  compact = false,
}: ConversationListProps) {
  if (conversations.length === 0) {
    return (
      <View padding="size-200">
        <Text>No conversations yet. Message a snack mate to get started.</Text>
      </View>
    );
  }

  return (
    <View UNSAFE_className="sm-conversation-list">
      {conversations.map((conv) => {
        const user = conv.other_user;
        const name = user?.display_name || user?.username || "Snack Mate";
        const preview = conv.last_message?.body ?? "No messages yet";
        const isActive = conv.id === activeId;

        return (
          <button
            key={conv.id}
            type="button"
            className={`sm-conversation-list__item${isActive ? " sm-conversation-list__item--active" : ""}`}
            onClick={() => onSelect(conv.id)}
          >
            <Flex alignItems="center" gap="size-150" width="100%">
              <Avatar src={avatarImageSrc(user?.avatar_url)} alt="" size={compact ? "avatar-size-300" : "avatar-size-400"} />
              <Flex direction="column" flex={1} minWidth={0}>
                <Flex justifyContent="space-between" alignItems="center" width="100%">
                  <Text UNSAFE_style={{ fontWeight: 600 }}>{name}</Text>
                  {conv.unread_count > 0 && (
                    <span className="sm-conversation-list__unread">{conv.unread_count}</span>
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
