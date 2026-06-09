"use client";

import Link from "next/link";
import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  Avatar,
  Button,
  Content,
  Flex,
  Heading,
  IllustratedMessage,
  Text,
  View,
} from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { api, Friendship, getToken, SnackMatch } from "@/lib/api";
import { countryName } from "@/lib/countries";
import { PageHero } from "@/components/PageHero";
import { NOTIFICATIONS_CHANGED, notifyNotificationsChanged } from "@/lib/notifications";

export default function SnackMatesPage() {
  return (
    <Suspense fallback={<Text>Loading...</Text>}>
      <SnackMatesContent />
    </Suspense>
  );
}

function SnackMatesContent() {
  const searchParams = useSearchParams();
  const [mates, setMates] = useState<Friendship[]>([]);
  const [randomMatches, setRandomMatches] = useState<SnackMatch[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionId, setActionId] = useState<string | null>(null);
  const [message, setMessage] = useState("");

  async function load() {
    const token = getToken();
    const [friends, matches] = await Promise.all([api.friends(token), api.matches(token)]);
    setMates(friends);
    setRandomMatches(matches);
    setLoading(false);
  }

  useEffect(() => {
    load();
    const onChange = () => load();
    window.addEventListener(NOTIFICATIONS_CHANGED, onChange);
    return () => window.removeEventListener(NOTIFICATIONS_CHANGED, onChange);
  }, []);

  useEffect(() => {
    const error = searchParams.get("error");
    if (error) {
      setMessage(error);
      return;
    }

    const paired = searchParams.get("random");
    if (paired === null) return;
    const count = Number(paired);
    if (Number.isNaN(count)) return;
    setMessage(
      count === 0
        ? "No new random snack mates were paired this round."
        : `Paired ${count} new random snack mate${count === 1 ? "" : "s"}.`,
    );
  }, [searchParams]);

  async function remove(id: string) {
    setActionId(id);
    try {
      await api.removeFriend(id, getToken());
      await load();
      notifyNotificationsChanged();
    } finally {
      setActionId(null);
    }
  }

  if (loading) {
    return <Text>Loading...</Text>;
  }

  const hasMates = mates.length > 0 || randomMatches.length > 0;

  return (
    <Flex direction="column" gap="size-400">
      <PageHero
        title="Snack Mates"
        description="Add snack mates from search or get randomly paired with someone from another country. Once connected, you can snag items from each other's wishlists."
      />

      {message && <Text>{message}</Text>}

      {!hasMates ? (
        <IllustratedMessage>
          <Heading>No Snack Mates Yet</Heading>
          <Content>
            <Text>
              Use the snack mates menu in the header to add a snack mate or try a random match.
            </Text>
          </Content>
        </IllustratedMessage>
      ) : (
        <Flex direction="column" gap="size-400">
          {mates.length > 0 && (
            <Flex direction="column" gap="size-200">
              <Heading level={2}>Your Snack Mates</Heading>
              {mates.map((mate) => (
                <SnackMateCard
                  key={mate.id}
                  friendship={mate}
                  actionId={actionId}
                  onRemove={remove}
                />
              ))}
            </Flex>
          )}

          {randomMatches.length > 0 && (
            <Flex direction="column" gap="size-200">
              <Heading level={2}>Random Matches</Heading>
              {randomMatches.map((match) => (
                <RandomMatchCard key={match.id} match={match} />
              ))}
            </Flex>
          )}
        </Flex>
      )}
    </Flex>
  );
}

function SnackMateCard({
  friendship,
  actionId,
  onRemove,
}: {
  friendship: Friendship;
  actionId: string | null;
  onRemove: (id: string) => void;
}) {
  const user = friendship.user;
  if (!user) return null;

  const busy = actionId === friendship.id;

  return (
    <View
      backgroundColor="gray-50"
      padding="size-300"
      borderRadius="medium"
      borderWidth="thin"
      borderColor="gray-300"
    >
      <Flex gap="size-200" alignItems="center" justifyContent="space-between" wrap>
        <Flex gap="size-200" alignItems="center">
          <Avatar src={avatarImageSrc(user.avatar_url)} alt={user.display_name} size="avatar-size-400" />
          <Flex direction="column">
            <Heading level={3} margin={0}>
              <Link
                href={`/users/${user.username}`}
                style={{ color: "inherit", textDecoration: "none" }}
              >
                {user.display_name}
              </Link>
            </Heading>
            {user.country && <Text>{countryName(user.country)}</Text>}
            {user.bio && <Text>{user.bio}</Text>}
          </Flex>
        </Flex>
        <Button
          variant="negative"
          onPress={() => onRemove(friendship.id)}
          isDisabled={busy}
        >
          Remove
        </Button>
      </Flex>
    </View>
  );
}

function RandomMatchCard({ match }: { match: SnackMatch }) {
  const mate = match.mate;

  return (
    <View
      backgroundColor="gray-50"
      padding="size-300"
      borderRadius="medium"
      borderWidth="thin"
      borderColor="gray-300"
    >
      <Flex gap="size-200" alignItems="center">
        <Avatar
          src={avatarImageSrc(mate?.avatar_url)}
          alt={mate?.display_name ?? "Snack Mate"}
          size="avatar-size-400"
        />
        <Flex direction="column">
          <Heading level={3} margin={0}>
            {mate?.username ? (
              <Link
                href={`/users/${mate.username}`}
                style={{ color: "inherit", textDecoration: "none" }}
              >
                {mate.display_name}
              </Link>
            ) : (
              (mate?.display_name ?? "Snack Mate")
            )}
          </Heading>
          {mate?.country && <Text>{countryName(mate.country)}</Text>}
          <Text>Status: {match.status}</Text>
          <Text>Matched: {new Date(match.matched_at).toLocaleDateString()}</Text>
          {mate?.bio && <Text>{mate.bio}</Text>}
        </Flex>
      </Flex>
    </View>
  );
}
