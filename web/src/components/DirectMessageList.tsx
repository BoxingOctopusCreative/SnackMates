"use client";

import { Flex, Text, View } from "@adobe/react-spectrum";
import { Conversation } from "@/lib/api";

type DirectMessageListProps = {
  conversations: Conversation[];
  activeId?: string | null;
  onSelect: (conversationId: string) => void;
};

export function DirectMessageList({
  conversations,
  activeId,
  onSelect,
}: DirectMessageListProps) {
  if (conversations.length === 0) {
    return (
      <View padding="size-200">
        <Text>No messages yet. Compose a note to a snack mate.</Text>
      </View>
    );
  }

  return (
    <View UNSAFE_className="sm-direct-message-list">
      {conversations.map((conv) => {
        const user = conv.other_user;
        const from = user?.display_name || user?.username || "Snack Mate";
        const subject = conv.last_message?.subject || "(no subject)";
        const preview = conv.last_message?.body ?? "No messages yet";
        const isActive = conv.id === activeId;

        return (
          <button
            key={conv.id}
            type="button"
            className={`sm-direct-message-list__item${isActive ? " sm-direct-message-list__item--active" : ""}`}
            onClick={() => onSelect(conv.id)}
          >
            <Flex direction="column" gap="size-50" width="100%" alignItems="start">
              <Flex justifyContent="space-between" alignItems="center" width="100%">
                <Text UNSAFE_className="sm-direct-message-list__subject">{subject}</Text>
                {conv.unread_count > 0 && (
                  <span className="sm-conversation-list__unread">{conv.unread_count}</span>
                )}
              </Flex>
              <Text UNSAFE_className="sm-direct-message-list__from">From {from}</Text>
              <Text UNSAFE_className="sm-direct-message-list__preview">{preview}</Text>
            </Flex>
          </button>
        );
      })}
    </View>
  );
}
