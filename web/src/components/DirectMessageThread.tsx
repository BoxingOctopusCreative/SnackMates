"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { Avatar, Button, Flex, Text, TextArea, TextField, View } from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { Conversation, Message } from "@/lib/api";

type DirectMessageThreadProps = {
  otherUser: Conversation["other_user"];
  messages: Message[];
  currentUserId?: string;
  onSend: (subject: string, body: string) => Promise<void>;
};

export function DirectMessageThread({
  otherUser,
  messages,
  currentUserId,
  onSend,
}: DirectMessageThreadProps) {
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [sending, setSending] = useState(false);

  const threadSubject = messages[0]?.subject || "(no subject)";
  const displayName = otherUser?.display_name || otherUser?.username || "Snack Mate";

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    const trimmedSubject = subject.trim() || threadSubject;
    const trimmedBody = body.trim();
    if (!trimmedSubject || !trimmedBody || sending) return;
    setSending(true);
    try {
      await onSend(trimmedSubject, trimmedBody);
      setBody("");
      if (messages.length === 0) {
        setSubject("");
      }
    } finally {
      setSending(false);
    }
  }

  return (
    <View height="100%" UNSAFE_className="sm-direct-message-thread">
      <View padding="size-200" UNSAFE_className="sm-direct-message-thread__header">
        <Flex alignItems="center" gap="size-150">
          {otherUser && (
            <Link href={`/users/${otherUser.username}`} className="sm-conversation-thread__profile">
              <Flex alignItems="center" gap="size-150">
                <Avatar src={avatarImageSrc(otherUser.avatar_url)} alt="" size="avatar-size-400" />
                <Flex direction="column">
                  <Text UNSAFE_style={{ fontWeight: 600 }}>{displayName}</Text>
                  <Text UNSAFE_className="sm-direct-message-thread__thread-subject">{threadSubject}</Text>
                </Flex>
              </Flex>
            </Link>
          )}
        </Flex>
      </View>

      <View
        flex
        overflow="auto"
        padding="size-200"
        UNSAFE_className="sm-direct-message-thread__messages"
        UNSAFE_style={{ flex: 1, minHeight: 0 }}
      >
        {messages.length === 0 ? (
          <Text>Write the first note to {displayName}.</Text>
        ) : (
          messages.map((msg) => {
            const mine = currentUserId ? msg.sender_id === currentUserId : false;
            return (
              <View key={msg.id} marginBottom="size-200" UNSAFE_className="sm-direct-message-mail">
                <Flex justifyContent="space-between" alignItems="baseline" wrap gap="size-100">
                  <Text UNSAFE_className="sm-direct-message-mail__from">
                    {mine ? "You" : displayName}
                  </Text>
                  <Text UNSAFE_className="sm-direct-message-mail__date">
                    {formatMailDate(msg.created_at)}
                  </Text>
                </Flex>
                {msg.subject && (
                  <Text UNSAFE_className="sm-direct-message-mail__subject">{msg.subject}</Text>
                )}
                <Text UNSAFE_className="sm-direct-message-mail__body">{msg.body}</Text>
              </View>
            );
          })
        )}
      </View>

      <form onSubmit={handleSubmit} className="sm-direct-message-thread__composer">
        <Flex direction="column" gap="size-150" width="100%">
          {messages.length === 0 && (
            <TextField
              label="Subject"
              value={subject}
              onChange={setSubject}
              width="100%"
              isRequired
              isDisabled={sending}
            />
          )}
          <TextArea
            label="Message"
            value={body}
            onChange={setBody}
            width="100%"
            isRequired
            isDisabled={sending}
          />
          <Button type="submit" variant="cta" isDisabled={sending || !body.trim() || (messages.length === 0 && !subject.trim())}>
            Send
          </Button>
        </Flex>
      </form>
    </View>
  );
}

function formatMailDate(iso: string) {
  const date = new Date(iso);
  return date.toLocaleString(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}
