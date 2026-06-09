"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Button,
  Flex,
  Form,
  Heading,
  Item,
  Picker,
  Text,
  TextField,
} from "@adobe/react-spectrum";
import { AuthPageShell } from "@/components/AuthPageShell";
import { DiscordOAuthButton } from "@/components/DiscordOAuthButton";
import { api, discordUrl } from "@/lib/api";
import { COUNTRIES } from "@/lib/countries";
import type { UnsplashPhoto } from "@/lib/unsplash";

export function RegisterForm({ background }: { background: UnsplashPhoto | null }) {
  const [displayName, setDisplayName] = useState("");
  const [country, setCountry] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    try {
      const res = await api.register({ email, password, display_name: displayName, country });
      setMessage(res.message);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthPageShell background={background}>
      <Heading level={2}>Join SnackMates</Heading>
      <Form maxWidth="100%" onSubmit={handleSubmit}>
        <Flex direction="column" gap="size-200">
          <TextField label="Display name" value={displayName} onChange={setDisplayName} isRequired />
          <Picker
            label="Country of origin"
            selectedKey={country}
            onSelectionChange={(key) => setCountry(String(key))}
          >
            {COUNTRIES.map((c) => (
              <Item key={c.id}>{c.name}</Item>
            ))}
          </Picker>
          <TextField label="Email" type="email" value={email} onChange={setEmail} isRequired />
          <TextField label="Password" type="password" value={password} onChange={setPassword} isRequired />
          {error && <Text UNSAFE_style={{ color: "var(--sm-error)" }}>{error}</Text>}
          {message && <Text UNSAFE_style={{ color: "#12805c" }}>{message}</Text>}
          <Button type="submit" variant="accent" isDisabled={loading}>
            {loading ? "Creating account..." : "Create account"}
          </Button>
        </Flex>
      </Form>
      <Flex direction="column" gap="size-150" marginTop="size-300">
        <DiscordOAuthButton href={discordUrl()}>Sign up with Discord</DiscordOAuthButton>
        <Text>
          Already have an account? <Link href="/login">Sign in</Link>
        </Text>
      </Flex>
    </AuthPageShell>
  );
}
