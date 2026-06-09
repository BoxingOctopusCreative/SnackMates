"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  Avatar,
  Button,
  Content,
  Flex,
  Heading,
  IllustratedMessage,
  Item,
  Picker,
  Text,
  TextArea,
  TextField,
  View,
} from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { api, Conversation, Friendship, getToken, Message } from "@/lib/api";
import { PageHero } from "@/components/PageHero";
import { DirectMessageList } from "@/components/DirectMessageList";
import { DirectMessageThread } from "@/components/DirectMessageThread";
import { useCurrentUserId } from "@/components/ConversationThread";
import {
  fetchConversationMessages,
  notifyMessagesChanged,
  openConversationWith,
  postMessage,
  useMessages,
} from "@/lib/messages";

export default function MessagesPage() {
  return (
    <Suspense fallback={<Text>Loading...</Text>}>
      <MessagesContent />
    </Suspense>
  );
}

function MessagesContent() {
  const searchParams = useSearchParams();
  const { conversations, loading, refresh } = useMessages();
  const currentUserId = useCurrentUserId();
  const [mates, setMates] = useState<Friendship[]>([]);
  const [selectedMate, setSelectedMate] = useState<string | null>(null);
  const [composeSubject, setComposeSubject] = useState("");
  const [composeBody, setComposeBody] = useState("");
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState("");
  const [activeConversationId, setActiveConversationId] = useState<string | null>(null);
  const [activeMessages, setActiveMessages] = useState<Message[]>([]);
  const [activeOtherUser, setActiveOtherUser] = useState<Conversation["other_user"] | null>(null);

  const openConversation = useCallback(
    async (conversationId: string) => {
      const data = await fetchConversationMessages(conversationId);
      const fromList = conversations.find((c) => c.id === conversationId)?.other_user;
      setActiveConversationId(conversationId);
      setActiveMessages(data.messages);
      setActiveOtherUser(data.conversation.other_user ?? fromList ?? null);
      notifyMessagesChanged();
    },
    [conversations],
  );

  useEffect(() => {
    api.friends(getToken()).then(setMates).catch(() => setMates([]));
  }, []);

  useEffect(() => {
    const conversationId = searchParams.get("c");
    if (conversationId) {
      openConversation(conversationId).catch(() => {});
    }
  }, [searchParams, openConversation]);

  async function handleCompose() {
    if (!selectedMate) return;
    const mate = mates.find((m) => m.id === selectedMate);
    const username = mate?.user?.username;
    if (!username) return;

    const subject = composeSubject.trim();
    const body = composeBody.trim();
    if (!subject || !body) return;

    setStarting(true);
    setError("");
    try {
      const conversationId = await openConversationWith(username);
      await postMessage(conversationId, subject, body);
      setComposeSubject("");
      setComposeBody("");
      await openConversation(conversationId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not send message");
    } finally {
      setStarting(false);
    }
  }

  async function handleReply(subject: string, body: string) {
    if (!activeConversationId) return;
    const msg = await postMessage(activeConversationId, subject, body);
    setActiveMessages((prev) => [...prev, msg]);
  }

  const showThread = Boolean(activeConversationId && activeOtherUser);

  return (
    <View>
      <PageHero
        title="Messages"
        description="Asynchronous notes with your snack mates — like email, not live chat."
      />

      <Flex
        gap="size-200"
        marginTop="size-300"
        direction={{ base: "column", M: "row" }}
        UNSAFE_className="sm-messages-layout"
      >
        <View width={{ base: "100%", M: "18rem" }} UNSAFE_className="sm-messages-sidebar">
          <View padding="size-200" UNSAFE_className="sm-messages-new">
            <Heading level={4} marginBottom="size-150">
              Compose
            </Heading>
            {mates.length === 0 ? (
              <Text>No snack mates yet. Add mates to send messages.</Text>
            ) : (
              <Flex direction="column" gap="size-150">
                <Picker
                  label="To"
                  selectedKey={selectedMate}
                  onSelectionChange={(key) => setSelectedMate(key as string)}
                  width="100%"
                  UNSAFE_className="sm-messages-mate-picker"
                >
                  {mates.map((mate) => {
                    const name = mate.user?.display_name ?? mate.user?.username ?? "";
                    return (
                      <Item key={mate.id} textValue={name}>
                        <Avatar src={avatarImageSrc(mate.user?.avatar_url)} alt="" size="avatar-size-200" />
                        <Text>{name}</Text>
                      </Item>
                    );
                  })}
                </Picker>
                <TextField
                  label="Subject"
                  value={composeSubject}
                  onChange={setComposeSubject}
                  width="100%"
                  isRequired
                />
                <TextArea
                  label="Message"
                  value={composeBody}
                  onChange={setComposeBody}
                  width="100%"
                  isRequired
                />
                <Button
                  variant="cta"
                  onPress={handleCompose}
                  isDisabled={!selectedMate || starting || !composeSubject.trim() || !composeBody.trim()}
                >
                  Send message
                </Button>
                {error && (
                  <Text UNSAFE_style={{ color: "var(--spectrum-global-color-red-600)" }}>{error}</Text>
                )}
              </Flex>
            )}
          </View>

          <Heading level={4} marginStart="size-200" marginTop="size-200" marginBottom="size-100">
            Inbox
          </Heading>
          {loading && conversations.length === 0 ? (
            <Text marginStart="size-200">Loading...</Text>
          ) : (
            <DirectMessageList
              conversations={conversations}
              activeId={activeConversationId}
              onSelect={(id) => openConversation(id)}
            />
          )}
        </View>

        <View flex={1} minHeight="24rem" UNSAFE_className="sm-messages-thread-panel">
          {showThread && activeOtherUser ? (
            <DirectMessageThread
              otherUser={activeOtherUser}
              messages={activeMessages}
              currentUserId={currentUserId}
              onSend={handleReply}
            />
          ) : (
            <IllustratedMessage>
              <Heading>Select a Message</Heading>
              <Content>
                <Text>
                  Choose a message from your inbox or compose a new note. For real-time conversation,
                  use Live Chat.
                </Text>
              </Content>
              <Button variant="secondary" onPress={() => refresh()}>
                Refresh
              </Button>
            </IllustratedMessage>
          )}
        </View>
      </Flex>
    </View>
  );
}
