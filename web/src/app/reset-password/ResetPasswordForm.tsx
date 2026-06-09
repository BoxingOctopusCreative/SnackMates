"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  Button,
  Flex,
  Form,
  Heading,
  Text,
  TextField,
  View,
} from "@adobe/react-spectrum";
import { api } from "@/lib/api";

export default function ResetPasswordForm() {
  const params = useSearchParams();
  const router = useRouter();
  const token = params.get("token") ?? "";
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      await api.resetPassword(token, password);
      router.push("/login");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Reset failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <View minHeight="100vh" backgroundColor="gray-100" padding="size-400">
      <View maxWidth="size-3600" marginX="auto" backgroundColor="gray-50" padding="size-400" borderRadius="large">
        <Heading level={2}>Choose a New Password</Heading>
        <Form maxWidth="100%" onSubmit={handleSubmit}>
          <Flex direction="column" gap="size-200">
            <TextField label="New password" type="password" value={password} onChange={setPassword} isRequired />
            {error && <Text UNSAFE_style={{ color: "var(--sm-error)" }}>{error}</Text>}
            <Button type="submit" variant="accent" isDisabled={loading || !token}>
              Update password
            </Button>
          </Flex>
        </Form>
      </View>
    </View>
  );
}
