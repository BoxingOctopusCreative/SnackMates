"use client";

import { FormEvent, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Icon } from "@iconify/react";
import { Avatar, Button, Flex, Text, TextField, View } from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { Chat, ChatMessage } from "@/lib/api";

type ChatThreadProps = {
  otherUser: Chat["other_user"];
  messages: ChatMessage[];
  currentUserId?: string;
  onSend: (body: string) => Promise<void>;
  onBack?: () => void;
};

export function ChatThread({
  otherUser,
  messages,
  currentUserId,
  onSend,
  onBack,
}: ChatThreadProps) {
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
    setDraft("");
    setSending(true);
    try {
      await onSend(body);
    } catch {
      setDraft(body);
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
            <button
              type="button"
              className="sm-chat-widget__close"
              aria-label="Back to chats"
              onClick={onBack}
            >
              <Icon icon="ion:chevron-back" aria-hidden />
            </button>
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
                <Text UNSAFE_className="sm-message-bubble__time" UNSAFE_style={{ fontSize: "0.75rem", opacity: 0.7 }}>
                  {formatChatTime(msg.created_at)}
                </Text>
              </View>
            );
          })
        )}
        <div ref={bottomRef} />
      </View>

      <form onSubmit={handleSubmit} className="sm-conversation-thread__composer">
        <TextField
          aria-label="Chat message"
          value={draft}
          onChange={setDraft}
          width="100%"
          isDisabled={sending}
        />
        <Button type="submit" variant="cta" isDisabled={sending || !draft.trim()}>
          Send
        </Button>
      </form>
    </View>
  );
}

function formatChatTime(iso: string) {
  const date = new Date(iso);
  return date.toLocaleString(undefined, {
    hour: "numeric",
    minute: "2-digit",
  });
}
