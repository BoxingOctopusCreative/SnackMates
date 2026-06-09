"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import {
  Content,
  Flex,
  Heading,
  IllustratedMessage,
  Text,
  View,
} from "@adobe/react-spectrum";
import { useAuth } from "@/components/AuthGate";
import { ProfileHero } from "@/components/ProfileHero";
import { api, ApiError, FriendshipView, getToken, User, UserProfile } from "@/lib/api";

export default function UserProfilePage() {
  const { user } = useAuth();
  return <UserProfileContent currentUser={user} />;
}

function UserProfileContent({ currentUser }: { currentUser: User }) {
  const params = useParams<{ username: string }>();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [friendship, setFriendship] = useState<FriendshipView | undefined>();
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .getUserProfile(params.username, getToken())
      .then((data) => {
        setProfile(data);
        setFriendship(data.friendship);
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "Failed to load profile");
      });
  }, [params.username, currentUser.bio, currentUser.display_name, currentUser.country]);

  if (error) {
    return (
      <IllustratedMessage>
        <Heading>Profile Not Found</Heading>
        <Content>
          <Text>{error}</Text>
        </Content>
      </IllustratedMessage>
    );
  }

  if (!profile) {
    return <Text>Loading...</Text>;
  }

  const { user, wishlists } = profile;
  const isOwnProfile = currentUser.id === user.id;

  return (
    <Flex direction="column" gap="size-400">
      <ProfileHero
        user={user}
        isOwnProfile={isOwnProfile}
        friendship={friendship}
        onFriendshipChange={setFriendship}
      />

      <Flex direction="column" gap="size-200">
        <Heading level={2}>Public Wishlists</Heading>
        {wishlists.length === 0 ? (
          <Text>No public wishlists yet.</Text>
        ) : (
          wishlists.map((list) => (
            <View
              key={list.id}
              backgroundColor="gray-50"
              padding="size-300"
              borderRadius="medium"
              borderWidth="thin"
              borderColor="gray-300"
            >
              <Flex direction="column" gap="size-100">
                <Link href={`/wishlists/${list.slug}`} style={{ textDecoration: "none" }}>
                  <Heading level={3} margin={0}>
                    {list.title}
                  </Heading>
                </Link>
                {list.description && <Text>{list.description}</Text>}
                <Text UNSAFE_style={{ color: "var(--sm-text-muted)" }}>
                  {list.item_count} {list.item_count === 1 ? "item" : "items"}
                </Text>
              </Flex>
            </View>
          ))
        )}
      </Flex>
    </Flex>
  );
}
