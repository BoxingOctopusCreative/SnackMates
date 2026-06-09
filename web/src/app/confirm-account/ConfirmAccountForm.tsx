"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  Flex,
  Heading,
  ProgressCircle,
  Text,
  View,
} from "@adobe/react-spectrum";
import { api, clearToken } from "@/lib/api";

export default function ConfirmAccountForm() {
  const params = useSearchParams();
  const router = useRouter();
  const token = params.get("token") ?? "";
  const [status, setStatus] = useState<"loading" | "ok" | "error">("loading");
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (!token) {
      setStatus("error");
      setMessage("Missing confirmation token.");
      return;
    }
    api
      .confirmAccountAction(token)
      .then((res) => {
        setStatus("ok");
        if (res.action === "delete") {
          clearToken();
          setMessage("Your account has been permanently deleted.");
          return;
        }
        if (res.action === "deactivate") {
          clearToken();
          setMessage("Your account has been deactivated. You can reactivate it anytime from the sign-in page.");
          return;
        }
        setMessage("Your account has been reactivated. You can sign in again.");
        setTimeout(() => router.push("/login"), 2000);
      })
      .catch((err) => {
        setStatus("error");
        setMessage(err instanceof Error ? err.message : "Confirmation failed");
      });
  }, [token, router]);

  return (
    <View minHeight="100vh" backgroundColor="gray-100" padding="size-400">
      <View maxWidth="size-3600" marginX="auto" backgroundColor="gray-50" padding="size-400" borderRadius="large">
        <Heading level={2}>Confirm Account Action</Heading>
        <Flex direction="column" gap="size-200" marginTop="size-200">
          {status === "loading" && <ProgressCircle isIndeterminate aria-label="Confirming" />}
          <Text>{message}</Text>
          {status === "ok" && (
            <Link href="/login">Go to sign in</Link>
          )}
        </Flex>
      </View>
    </View>
  );
}
