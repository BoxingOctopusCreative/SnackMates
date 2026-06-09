"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Button,
  Flex,
  Form,
  Heading,
  Text,
  TextField,
} from "@adobe/react-spectrum";
import { AuthPageShell } from "@/components/AuthPageShell";
import { DiscordOAuthButton } from "@/components/DiscordOAuthButton";
import { ApiError, api, discordUrl, saveToken } from "@/lib/api";
import type { UnsplashPhoto } from "@/lib/unsplash";

export function LoginForm({ background }: { background: UnsplashPhoto | null }) {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const [mfaRequired, setMfaRequired] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [accountDeactivated, setAccountDeactivated] = useState(false);
  const [reactivateMessage, setReactivateMessage] = useState("");
  const [reactivateLoading, setReactivateLoading] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const oauthError = params.get("error");
    if (oauthError) {
      setError(oauthError);
      if (oauthError.toLowerCase().includes("deactivated")) {
        setAccountDeactivated(true);
      }
      window.history.replaceState({}, "", window.location.pathname);
    }
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    setAccountDeactivated(false);
    setReactivateMessage("");
    try {
      const res = await api.login({ email, password, totp_code: totpCode || undefined });
      if (res.mfa_required) {
        setMfaRequired(true);
        return;
      }
      if (res.token) {
        saveToken(res.token);
        router.push("/dashboard");
      }
    } catch (err) {
      if (err instanceof ApiError && err.code === "account_deactivated") {
        setAccountDeactivated(true);
        setError(err.message);
      } else {
        setAccountDeactivated(false);
        setError(err instanceof Error ? err.message : "Login failed");
      }
    } finally {
      setLoading(false);
    }
  }

  async function requestReactivation() {
    if (!email) {
      setError("Enter your email address to request a reactivation link.");
      return;
    }
    setReactivateLoading(true);
    setReactivateMessage("");
    try {
      const res = await api.requestAccountReactivate(email);
      setReactivateMessage(res.message);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to send reactivation email");
    } finally {
      setReactivateLoading(false);
    }
  }

  return (
    <AuthPageShell background={background}>
      <Heading level={2}>Sign In to SnackMates</Heading>
      <Form maxWidth="100%" onSubmit={handleSubmit}>
        <Flex direction="column" gap="size-200">
          <TextField label="Email" type="email" value={email} onChange={setEmail} isRequired />
          <TextField label="Password" type="password" value={password} onChange={setPassword} isRequired />
          {mfaRequired && (
            <TextField
              label="Authenticator code"
              value={totpCode}
              onChange={setTotpCode}
              description="Enter the 6-digit code from your authenticator app."
            />
          )}
          {error && <Text UNSAFE_style={{ color: "var(--sm-error)" }}>{error}</Text>}
          {reactivateMessage && <Text>{reactivateMessage}</Text>}
          {accountDeactivated && (
            <Button variant="secondary" onPress={requestReactivation} isDisabled={reactivateLoading}>
              {reactivateLoading ? "Sending..." : "Send Reactivation Email"}
            </Button>
          )}
          <Button type="submit" variant="accent" isDisabled={loading}>
            {loading ? "Signing in..." : "Sign in"}
          </Button>
        </Flex>
      </Form>
      <Flex direction="column" gap="size-150" marginTop="size-300">
        <DiscordOAuthButton href={discordUrl()}>Continue with Discord</DiscordOAuthButton>
        <Text>
          No account? <Link href="/register">Register</Link>
        </Text>
        <Text>
          <Link href="/forgot-password">Forgot password?</Link>
        </Text>
      </Flex>
    </AuthPageShell>
  );
}
