"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import {
  Avatar,
  Button,
  Flex,
  Form,
  Heading,
  Item,
  Picker,
  Text,
  TextArea,
  TextField,
  View,
} from "@adobe/react-spectrum";
import { useAuth } from "@/components/AuthGate";
import { BannerEditor } from "@/components/BannerEditor";
import { avatarImageSrc } from "@/lib/avatar";
import { API_URL, api, discordConnect, getToken } from "@/lib/api";
import { bufferToBase64URL, parseCreationOptions, webAuthnErrorMessage } from "@/lib/webauthn";
import { COUNTRIES } from "@/lib/countries";

type SettingsPanelProps = {
  showHeading?: boolean;
  showViewProfileLink?: boolean;
  embedded?: boolean;
};

export function SettingsPanel({
  showHeading = true,
  showViewProfileLink = true,
  embedded = false,
}: SettingsPanelProps) {
  const { user, updateUser, refreshUser } = useAuth();
  const [displayName, setDisplayName] = useState(user.display_name);
  const [bio, setBio] = useState(user.bio ?? "");
  const [country, setCountry] = useState(user.country || "");

  useEffect(() => {
    setDisplayName(user.display_name);
    setBio(user.bio ?? "");
    setCountry(user.country || "");
  }, [user.display_name, user.bio, user.country]);
  const [message, setMessage] = useState("");
  const [totpSecret, setTotpSecret] = useState("");
  const [otpauthUrl, setOtpauthUrl] = useState("");
  const [totpCode, setTotpCode] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  const discordLinked = Boolean(user.discord_linked || user.discord_id);

  async function connectDiscord() {
    try {
      const res = await discordConnect(getToken());
      window.location.href = res.url;
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Failed to connect Discord");
    }
  }

  async function saveProfile(e: React.FormEvent) {
    e.preventDefault();
    try {
      await api.updateProfile({ display_name: displayName, bio, country }, getToken());
      await refreshUser();
      setMessage("Profile updated.");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Failed to update profile");
    }
  }

  async function onAvatarSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      const res = await api.uploadAvatar(file, getToken());
      updateUser({ avatar_url: res.avatar_url });
      setMessage("Profile picture updated.");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Upload failed");
    }
  }

  async function setupTotp() {
    const res = await api.totpSetup(getToken());
    setTotpSecret(res.secret);
    setOtpauthUrl(res.otpauth_url);
  }

  async function enableTotp() {
    await api.totpEnable(totpCode, getToken());
    setMessage("TOTP enabled.");
  }

  async function disableTotp() {
    await api.totpDisable(totpCode, getToken());
    setMessage("TOTP disabled.");
  }

  async function requestDeactivate() {
    try {
      const res = await api.requestAccountDeactivate(getToken());
      setMessage(res.message);
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Failed to request deactivation");
    }
  }

  async function requestDelete() {
    try {
      const res = await api.requestAccountDelete(getToken());
      setMessage(res.message);
    } catch (err) {
      setMessage(err instanceof Error ? err.message : "Failed to request deletion");
    }
  }

  async function registerWebAuthn() {
    try {
      const begin = await api.webauthnRegisterBegin(getToken());
      const publicKey = parseCreationOptions(begin.options);

      const credential = await navigator.credentials.create({ publicKey });
      if (!credential || credential.type !== "public-key") {
        setMessage("Security key registration was cancelled. No key was registered.");
        return;
      }

      const publicKeyCredential = credential as PublicKeyCredential;
      const attestation = publicKeyCredential.response as AuthenticatorAttestationResponse;
      const body = JSON.stringify({
        id: publicKeyCredential.id,
        rawId: bufferToBase64URL(publicKeyCredential.rawId),
        type: publicKeyCredential.type,
        response: {
          attestationObject: bufferToBase64URL(attestation.attestationObject),
          clientDataJSON: bufferToBase64URL(attestation.clientDataJSON),
        },
      });

      const res = await fetch(`${API_URL}/api/v1/auth/mfa/webauthn/register/finish`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${getToken()}`,
          "Content-Type": "application/json",
          "X-WebAuthn-Session": begin.session_data,
        },
        credentials: "include",
        body,
      });
      if (!res.ok) {
        const data = await res.json();
        setMessage(data.error ?? "WebAuthn registration failed");
        return;
      }
      setMessage("Security key registered.");
    } catch (err) {
      setMessage(webAuthnErrorMessage(err));
    }
  }

  const profileSection = (
    <>
      <Flex justifyContent="space-between" alignItems="center">
        <Heading level={embedded ? 4 : 3} margin={0}>Profile</Heading>
        {showViewProfileLink && <Link href={`/users/${user.username}`}>View public profile</Link>}
      </Flex>
      <BannerEditor
        bannerUrl={user.banner_url}
        onBannerChange={(url) => updateUser({ banner_url: url })}
        onMessage={setMessage}
        embedded={embedded}
      />
      <div className={embedded ? "sm-settings-modal__avatar-block" : undefined}>
        <Avatar
          src={avatarImageSrc(user.avatar_url)}
          alt={user.display_name}
          size={embedded ? 48 : "avatar-size-700"}
          UNSAFE_className={embedded ? "sm-settings-modal__avatar" : undefined}
        />
        <div className={embedded ? "sm-settings-modal__avatar-actions" : undefined}>
          {discordLinked ? (
            <Text>
              Display name and profile picture sync from Discord when you sign in or reconnect.
            </Text>
          ) : (
            <Flex direction="column" gap="size-150" alignItems="start">
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                hidden
                onChange={onAvatarSelected}
              />
              <Button variant="secondary" onPress={() => fileInputRef.current?.click()}>
                Upload profile picture
              </Button>
              <Button variant="accent" onPress={connectDiscord}>
                Connect Discord
              </Button>
            </Flex>
          )}
        </div>
      </div>
      <Form maxWidth="100%" onSubmit={saveProfile} width="100%">
        <Flex direction="column" gap="size-200" width="100%">
          <TextField
            label="Display name"
            value={displayName}
            onChange={setDisplayName}
            isReadOnly={discordLinked}
            width="100%"
          />
          <Picker
            label="Country of origin"
            selectedKey={country}
            onSelectionChange={(key) => setCountry(String(key))}
            width="100%"
          >
            {COUNTRIES.map((c) => (
              <Item key={c.id}>{c.name}</Item>
            ))}
          </Picker>
          <TextArea label="Bio" value={bio} onChange={setBio} width="100%" />
          <Button type="submit" variant="accent">Save profile</Button>
        </Flex>
      </Form>
    </>
  );

  const accountSection = (
    <>
      <Heading level={embedded ? 4 : 3} margin={0}>Account</Heading>
      <Flex direction="column" gap="size-200" width="100%">
        <Text>
          Deactivate your account to hide your profile and pause matching. You can reactivate later.
          Deletion permanently removes your account and all data.
        </Text>
        <Button variant="secondary" onPress={requestDeactivate}>
          Deactivate Account
        </Button>
        <Button variant="negative" onPress={requestDelete}>
          Delete Account
        </Button>
      </Flex>
    </>
  );

  const mfaSection = (
    <>
      <Heading level={embedded ? 4 : 3} margin={0}>Multi-Factor Authentication</Heading>
      <Flex direction="column" gap="size-200" width="100%">
        <Text>TOTP: {user.totp_enabled ? "Enabled" : "Disabled"}</Text>
        <Text>WebAuthn: {user.has_webauthn ? "Registered" : "Not registered"}</Text>
        {!user.totp_enabled && (
          <>
            <Button variant="secondary" onPress={setupTotp}>Set up authenticator app</Button>
            {totpSecret && (
              <>
                <Text>Secret: {totpSecret}</Text>
                <Text>URI: {otpauthUrl}</Text>
                <TextField label="Verification code" value={totpCode} onChange={setTotpCode} width="100%" />
                <Button variant="accent" onPress={enableTotp}>Enable TOTP</Button>
              </>
            )}
          </>
        )}
        {user.totp_enabled && (
          <>
            <TextField label="Verification code" value={totpCode} onChange={setTotpCode} width="100%" />
            <Button variant="negative" onPress={disableTotp}>Disable TOTP</Button>
          </>
        )}
        <Button variant="secondary" onPress={registerWebAuthn}>Register security key (WebAuthn)</Button>
      </Flex>
    </>
  );

  if (embedded) {
    return (
      <div className="sm-settings-modal__content">
        {message && <Text>{message}</Text>}
        <section className="sm-settings-modal__section">{profileSection}</section>
        <section className="sm-settings-modal__section">{mfaSection}</section>
        <section className="sm-settings-modal__section">{accountSection}</section>
      </div>
    );
  }

  return (
    <Flex direction="column" gap="size-400">
      {showHeading && <Heading level={1}>Account Settings</Heading>}
      {message && <Text>{message}</Text>}

      <View
        backgroundColor="gray-50"
        padding="size-300"
        borderRadius="medium"
        borderWidth="thin"
        borderColor="gray-300"
      >
        {profileSection}
      </View>

      <View
        backgroundColor="gray-50"
        padding="size-300"
        borderRadius="medium"
        borderWidth="thin"
        borderColor="gray-300"
      >
        {mfaSection}
      </View>

      <View
        backgroundColor="gray-50"
        padding="size-300"
        borderRadius="medium"
        borderWidth="thin"
        borderColor="gray-300"
      >
        {accountSection}
      </View>
    </Flex>
  );
}
