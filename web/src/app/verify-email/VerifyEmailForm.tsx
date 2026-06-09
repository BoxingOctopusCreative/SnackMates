"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  Flex,
  Heading,
  ProgressCircle,
  Text,
  View,
} from "@adobe/react-spectrum";
import { api } from "@/lib/api";

export default function VerifyEmailForm() {
  const params = useSearchParams();
  const token = params.get("token") ?? "";
  const [status, setStatus] = useState<"loading" | "ok" | "error">("loading");
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (!token) {
      setStatus("error");
      setMessage("Missing verification token.");
      return;
    }
    api
      .verifyEmail(token)
      .then(() => {
        setStatus("ok");
        setMessage("Email verified! You can sign in and start building wishlists.");
      })
      .catch((err) => {
        setStatus("error");
        setMessage(err instanceof Error ? err.message : "Verification failed");
      });
  }, [token]);

  return (
    <View minHeight="100vh" backgroundColor="gray-100" padding="size-400">
      <View maxWidth="size-3600" marginX="auto" backgroundColor="gray-50" padding="size-400" borderRadius="large">
        <Heading level={2}>Email Verification</Heading>
        <Flex direction="column" gap="size-200" marginTop="size-200">
          {status === "loading" && <ProgressCircle isIndeterminate aria-label="Verifying" />}
          <Text>{message}</Text>
        </Flex>
      </View>
    </View>
  );
}
