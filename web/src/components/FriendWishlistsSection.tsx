"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import {
  Cell,
  Column,
  Flex,
  Heading,
  Row,
  TableBody,
  TableHeader,
  TableView,
  Text,
} from "@adobe/react-spectrum";
import { api, FriendWishlist, getToken } from "@/lib/api";

export function FriendWishlistsSection() {
  const [friendLists, setFriendLists] = useState<FriendWishlist[]>([]);

  useEffect(() => {
    api.friendWishlists(getToken()).then(setFriendLists);
  }, []);

  return (
    <Flex direction="column" gap="size-150">
      <Heading level={2} margin={0}>
        Snack Mates&apos; Wishlists
      </Heading>
      <Text>Latest updated public wishlists from your snack mates.</Text>
      <div className="sm-wishlists-table-panel">
        {friendLists.length === 0 ? (
          <Text margin="size-200">No public wishlists from your snack mates yet.</Text>
        ) : (
          <TableView aria-label="Snack mates' wishlists" selectionMode="none">
            <TableHeader>
              <Column key="mate">Snack Mate</Column>
              <Column key="title">Title</Column>
              <Column key="items">Items</Column>
              <Column key="updated">Updated</Column>
            </TableHeader>
            <TableBody items={friendLists}>
              {(item) => (
                <Row key={item.id}>
                  <Cell>
                    <Link href={`/users/${item.owner.username}`} className="sm-wishlist-title-link">
                      {item.owner.display_name || item.owner.username}
                    </Link>
                  </Cell>
                  <Cell>
                    <Link href={`/wishlists/${item.slug}`} className="sm-wishlist-title-link">
                      {item.title}
                    </Link>
                  </Cell>
                  <Cell>{item.item_count}</Cell>
                  <Cell>{formatWishlistUpdated(item.updated_at)}</Cell>
                </Row>
              )}
            </TableBody>
          </TableView>
        )}
      </div>
    </Flex>
  );
}

function formatWishlistUpdated(iso: string) {
  const date = new Date(iso);
  return date.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}
