"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button, Flex } from "@adobe/react-spectrum";
import { useChatContext } from "@/components/ChatProvider";

export function ProfileContactButtons({ username }: { username: string }) {
  const router = useRouter();
  const { startChat, setChatOpen } = useChatContext();
  const [opening, setOpening] = useState(false);

  async function openLiveChat() {
    setOpening(true);
    try {
      await startChat(username);
      setChatOpen(true);
    } finally {
      setOpening(false);
    }
  }

  return (
    <Flex gap="size-100" wrap>
      <Button variant="secondary" onPress={() => router.push("/messages")}>
        Send message
      </Button>
      <Button variant="accent" onPress={openLiveChat} isDisabled={opening}>
        {opening ? "Opening..." : "Live chat"}
      </Button>
    </Flex>
  );
}
