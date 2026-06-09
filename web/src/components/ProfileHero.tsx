"use client";

import { Icon } from "@iconify/react";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { Avatar, Flex, Heading, Text } from "@adobe/react-spectrum";
import { useSettingsModal } from "@/components/SettingsModalProvider";
import { ProfileContactButtons } from "@/components/ProfileContactButtons";
import { SnackMateActionButton } from "@/components/SnackMateActionButton";
import { avatarImageSrc } from "@/lib/avatar";
import { FriendshipView, PublicUser } from "@/lib/api";
import { countryName } from "@/lib/countries";

type ProfileHeroProps = {
  user: PublicUser;
  isOwnProfile?: boolean;
  friendship?: FriendshipView;
  onFriendshipChange?: (friendship?: FriendshipView) => void;
};

export function ProfileHero({
  user,
  isOwnProfile,
  friendship,
  onFriendshipChange,
}: ProfileHeroProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { openSettings } = useSettingsModal();
  const [settingsHovered, setSettingsHovered] = useState(false);
  const hasBanner = Boolean(user.banner_url);
  const settingsQuery = searchParams.get("settings");
  const discordQuery = searchParams.get("discord");

  useEffect(() => {
    if (!isOwnProfile) return;
    if (settingsQuery !== "open" && discordQuery !== "connected") return;

    openSettings();
    const params = new URLSearchParams(searchParams.toString());
    params.delete("settings");
    params.delete("discord");
    const query = params.toString();
    router.replace(`/users/${user.username}${query ? `?${query}` : ""}`, { scroll: false });
  }, [discordQuery, isOwnProfile, openSettings, router, searchParams, settingsQuery, user.username]);

  return (
    <>
    <section className="sm-profile-hero" aria-label={`${user.display_name}'s profile`}>
      <div
        className={`sm-profile-hero__banner${hasBanner ? "" : " sm-profile-hero__banner--default"}`}
        style={
          hasBanner
            ? {
                backgroundImage: `var(--sm-hero-banner-overlay-image), url("${user.banner_url}")`,
              }
            : undefined
        }
        role={hasBanner ? "img" : undefined}
        aria-label={hasBanner ? `${user.display_name}'s profile banner` : undefined}
      />
      <div className="sm-profile-hero__body">
        <Flex direction="column" gap="size-150">
          <Flex alignItems="end" justifyContent="space-between" wrap gap="size-200">
            <div className="sm-profile-hero__avatar-wrap">
              <Avatar
                src={avatarImageSrc(user.avatar_url)}
                alt={user.display_name}
                size="avatar-size-1000"
              />
            </div>
            {isOwnProfile ? (
              <button
                type="button"
                aria-label="Profile Settings"
                title="Profile Settings"
                className="sm-profile-hero__settings-link"
                onClick={openSettings}
                onMouseEnter={() => setSettingsHovered(true)}
                onMouseLeave={() => setSettingsHovered(false)}
                onFocus={() => setSettingsHovered(true)}
                onBlur={() => setSettingsHovered(false)}
              >
                <Icon
                  icon={settingsHovered ? "ion:settings" : "ion:settings-outline"}
                  className="sm-profile-hero__settings-icon"
                  aria-hidden
                />
              </button>
            ) : friendship?.status === "friends" ? (
              <ProfileContactButtons username={user.username} />
            ) : (
              <SnackMateActionButton
                username={user.username}
                friendship={friendship}
                onChange={onFriendshipChange}
              />
            )}
          </Flex>
          <Flex direction="column" gap="size-75">
            <Heading level={1} margin={0}>
              {user.display_name}
            </Heading>
            {user.country && (
              <Text UNSAFE_style={{ color: "var(--sm-text-muted)" }}>
                {countryName(user.country)}
              </Text>
            )}
            {user.bio && <Text>{user.bio}</Text>}
          </Flex>
        </Flex>
      </div>
    </section>

    </>
  );
}
