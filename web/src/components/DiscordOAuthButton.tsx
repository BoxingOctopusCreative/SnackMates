"use client";

import { Icon } from "@iconify/react";
import { Button } from "@adobe/react-spectrum";

export function DiscordOAuthButton({
  children,
  href,
}: {
  children: React.ReactNode;
  href: string;
}) {
  return (
    <Button
      variant="accent"
      width="100%"
      UNSAFE_className="sm-discord-oauth-button"
      onPress={() => {
        window.location.href = href;
      }}
    >
      <span className="sm-discord-oauth-button__content">
        <Icon icon="ion:logo-discord" className="sm-discord-oauth-button__icon" aria-hidden />
        <span>{children}</span>
      </span>
    </Button>
  );
}
