"use client";

import { FormEvent, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Avatar, Button, Flex, Text, TextField, View } from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { api, Conversation, getToken, Message } from "@/lib/api";

type ConversationThreadProps = {
  otherUser: Conversation["other_user"];
  messages: Message[];
  currentUserId?: string;
  onSend: (body: string) => Promise<void>;
  onBack?: () => void;
  compact?: boolean;
};

export function ConversationThread({
  otherUser,
  messages,
  currentUserId,
  onSend,
  onBack,
  compact = false,
}: ConversationThreadProps) {
  const [draft, setDraft] = useState("");
  const [sending, setSending] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const body = draft.trim();
    if (!body || sending) return;
    setSending(true);
    try {
      await onSend(body);
      setDraft("");
    } finally {
      setSending(false);
    }
  }

  const displayName = otherUser?.display_name || otherUser?.username || "Snack Mate";

  return (
    <View height="100%" UNSAFE_className="sm-conversation-thread">
      <View padding="size-200" UNSAFE_className="sm-conversation-thread__header">
      <Flex alignItems="center" gap="size-150">
        {onBack && (
          <Button variant="secondary" onPress={onBack} aria-label="Back to conversations">
            Back
          </Button>
        )}
        {otherUser && (
          <Link href={`/users/${otherUser.username}`} className="sm-conversation-thread__profile">
            <Flex alignItems="center" gap="size-150">
              <Avatar src={avatarImageSrc(otherUser.avatar_url)} alt="" size="avatar-size-400" />
              <Text UNSAFE_style={{ fontWeight: 600 }}>{displayName}</Text>
            </Flex>
          </Link>
        )}
      </Flex>
      </View>

      <View
        flex
        overflow="auto"
        padding="size-200"
        UNSAFE_className="sm-conversation-thread__messages"
        UNSAFE_style={{ flex: 1, minHeight: 0 }}
      >
        {messages.length === 0 ? (
          <Text marginTop="size-200">Say hello to {displayName}.</Text>
        ) : (
          messages.map((msg) => {
            const mine = currentUserId ? msg.sender_id === currentUserId : false;
            return (
              <View
                key={msg.id}
                marginBottom="size-150"
                UNSAFE_className={`sm-message-bubble ${mine ? "sm-message-bubble--mine" : "sm-message-bubble--theirs"}`}
              >
                <Text>{msg.body}</Text>
                <Text
                  UNSAFE_className="sm-message-bubble__time"
                  UNSAFE_style={{ fontSize: "0.75rem", opacity: 0.7 }}
                >
                  {formatMessageTime(msg.created_at)}
                </Text>
              </View>
            );
          })
        )}
        <div ref={bottomRef} />
      </View>

      <form onSubmit={handleSubmit} className="sm-conversation-thread__composer">
        <TextField
          aria-label="Message"
          value={draft}
          onChange={setDraft}
          width={compact ? "100%" : "100%"}
          isDisabled={sending}
        />
        <Button type="submit" variant="cta" isDisabled={sending || !draft.trim()}>
          Send
        </Button>
      </form>
    </View>
  );
}

function formatMessageTime(iso: string) {
  const date = new Date(iso);
  return date.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

export function useCurrentUserId() {
  const [userId, setUserId] = useState<string | undefined>();
  useEffect(() => {
    api.me(getToken()).then((user) => setUserId(user.id)).catch(() => {});
  }, []);
  return userId;
}
