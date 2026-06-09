"use client";

import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import { Icon } from "@iconify/react";
import { Flex, Heading, Provider, Text, View, defaultTheme } from "@adobe/react-spectrum";
import { useTheme } from "@/components/ThemeProvider";
import { useChatContext } from "@/components/ChatProvider";
import { ChatList } from "@/components/ChatList";
import { ChatThread } from "@/components/ChatThread";
import { useCurrentUserId } from "@/components/ConversationThread";

export function ChatWidget() {
  const { colorScheme } = useTheme();
  const {
    chats,
    unreadCount,
    chatOpen,
    setChatOpen,
    activeChatId,
    activeMessages,
    activeOtherUser,
    openChat,
    sendChatMessage,
    closeChat,
  } = useChatContext();
  const currentUserId = useCurrentUserId();
  const [mounted, setMounted] = useState(false);
  const [hovered, setHovered] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) return null;

  return createPortal(
    <Provider theme={defaultTheme} colorScheme={colorScheme}>
      <div className="sm-chat-widget">
        {chatOpen && (
          <View UNSAFE_className="sm-chat-widget__panel">
            <View UNSAFE_className="sm-chat-widget__header">
              <Flex alignItems="center" justifyContent="space-between">
                <Heading level={3} margin={0}>
                  {activeChatId ? "Chat" : "Live Chat"}
                </Heading>
                <button
                  type="button"
                  className="sm-chat-widget__close"
                  aria-label="Close chat"
                  onClick={() => {
                    setChatOpen(false);
                    closeChat();
                  }}
                >
                  <Icon icon="ion:close" aria-hidden />
                </button>
              </Flex>
            </View>

            <View UNSAFE_className="sm-chat-widget__body">
              {activeChatId && activeOtherUser ? (
                <View UNSAFE_className="sm-chat-widget__thread">
                  <ChatThread
                    otherUser={activeOtherUser}
                    messages={activeMessages}
                    currentUserId={currentUserId}
                    onSend={sendChatMessage}
                    onBack={closeChat}
                  />
                </View>
              ) : (
                <View UNSAFE_className="sm-chat-widget__list">
                  {chats.length === 0 ? (
                    <Text>Start a live chat from a snack mate&apos;s profile.</Text>
                  ) : (
                    <ChatList
                      chats={chats}
                      activeId={activeChatId}
                      onSelect={(id) => openChat(id)}
                    />
                  )}
                </View>
              )}
            </View>
          </View>
        )}

        <button
          type="button"
          className="sm-chat-widget__fab"
          aria-label={chatOpen ? "Close live chat" : unreadCount > 0 ? `Open live chat, ${unreadCount} unread` : "Open live chat"}
          onClick={() => setChatOpen(!chatOpen)}
          onMouseEnter={() => setHovered(true)}
          onMouseLeave={() => setHovered(false)}
        >
          <Icon
            icon={chatOpen || hovered ? "ion:chatbubble" : "ion:chatbubble-outline"}
            className="sm-chat-widget__fab-icon"
            aria-hidden
          />
          {!chatOpen && unreadCount > 0 && (
            <span className="sm-chat-widget__badge" aria-hidden>
              {unreadCount > 99 ? "99+" : unreadCount}
            </span>
          )}
        </button>
      </div>
    </Provider>,
    document.body,
  );
}
