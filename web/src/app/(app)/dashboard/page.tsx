"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Flex, Text, View } from "@adobe/react-spectrum";
import { useAuth } from "@/components/AuthGate";
import { FriendWishlistsSection } from "@/components/FriendWishlistsSection";
import { PageHero } from "@/components/PageHero";
import { api, Friendship, getToken, Wishlist } from "@/lib/api";

export default function DashboardPage() {
  const { user } = useAuth();
  return <Dashboard userId={user.id} verified={user.email_verified} name={user.display_name} />;
}

function Dashboard({
  verified,
  name,
}: {
  userId: string;
  verified: boolean;
  name: string;
}) {
  const [wishlists, setWishlists] = useState<Wishlist[]>([]);
  const [mates, setMates] = useState<Friendship[]>([]);

  useEffect(() => {
    const token = getToken();
    Promise.all([api.wishlists(token), api.friends(token)]).then(([w, m]) => {
      setWishlists(w);
      setMates(m);
    });
  }, []);

  return (
    <Flex direction="column" gap="size-300">
      <PageHero title={`Welcome Back, ${name}`} ariaLabel="Dashboard" />
      {!verified && (
        <View backgroundColor="notice" padding="size-200" borderRadius="medium">
          <Text>Verify your email to participate in snack matching. Check Mailpit at localhost:8025 in dev.</Text>
        </View>
      )}
      <Flex gap="size-200" wrap>
        <StatCard label="Wishlists" value={String(wishlists.length)} href="/wishlists" />
        <StatCard label="Snack Mates" value={String(mates.length)} href="/matches" />
      </Flex>
      <FriendWishlistsSection />
    </Flex>
  );
}

function StatCard({ label, value, href }: { label: string; value: string; href: string }) {
  return (
    <Link href={href} style={{ textDecoration: "none" }}>
      <View
        backgroundColor="gray-50"
        padding="size-300"
        borderRadius="medium"
        borderWidth="thin"
        borderColor="gray-300"
        minWidth="size-2000"
      >
        <Text UNSAFE_style={{ fontSize: "2rem", fontWeight: 700 }}>{value}</Text>
        <Text>{label}</Text>
      </View>
    </Link>
  );
}
