"use client";

import { useState } from "react";
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

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await api.forgotPassword(email);
      setMessage(res.message);
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Request failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <View minHeight="100vh" backgroundColor="gray-100" padding="size-400">
      <View maxWidth="size-3600" marginX="auto" backgroundColor="gray-50" padding="size-400" borderRadius="large">
        <Heading level={2}>Reset Password</Heading>
        <Form maxWidth="100%" onSubmit={handleSubmit}>
          <Flex direction="column" gap="size-200">
            <TextField label="Email" type="email" value={email} onChange={setEmail} isRequired />
            {message && <Text>{message}</Text>}
            <Button type="submit" variant="accent" isDisabled={loading}>
              Send reset link
            </Button>
          </Flex>
        </Form>
      </View>
    </View>
  );
}
