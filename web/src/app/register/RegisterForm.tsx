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
  ProgressCircle,
  Text,
  TextField,
  View,
} from "@adobe/react-spectrum";
import { AuthPageShell } from "@/components/AuthPageShell";
import { DiscordOAuthButton } from "@/components/DiscordOAuthButton";
import { TurnstileWidget } from "@/components/TurnstileWidget";
import { useExistingSessionRedirect } from "@/components/useExistingSessionRedirect";
import { api, discordUrl } from "@/lib/api";
import { turnstileEnabled } from "@/lib/turnstile";
import { COUNTRIES } from "@/lib/countries";
import type { UnsplashPhoto } from "@/lib/unsplash";

export function RegisterForm({ background }: { background: UnsplashPhoto | null }) {
  const checkingSession = useExistingSessionRedirect();
  const [displayName, setDisplayName] = useState("");
  const [country, setCountry] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [turnstileToken, setTurnstileToken] = useState("");
  const [turnstileResetKey, setTurnstileResetKey] = useState(0);
  const turnstileOn = turnstileEnabled();

  function resetTurnstile() {
    setTurnstileToken("");
    setTurnstileResetKey((key) => key + 1);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (turnstileOn && !turnstileToken) {
      setError("Please complete the security check.");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const res = await api.register({
        email,
        password,
        display_name: displayName,
        country,
        turnstile_token: turnstileToken || undefined,
      });
      setMessage(res.message);
      resetTurnstile();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
      resetTurnstile();
    } finally {
      setLoading(false);
    }
  }

  if (checkingSession) {
    return (
      <View minHeight="100vh" UNSAFE_style={{ display: "grid", placeItems: "center" }}>
        <ProgressCircle isIndeterminate aria-label="Loading" />
      </View>
    );
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
          {turnstileOn && (
            <TurnstileWidget
              resetKey={turnstileResetKey}
              onToken={setTurnstileToken}
              onExpire={resetTurnstile}
              onError={resetTurnstile}
            />
          )}
          <Button
            type="submit"
            variant="accent"
            isDisabled={loading || (turnstileOn && !turnstileToken)}
          >
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
