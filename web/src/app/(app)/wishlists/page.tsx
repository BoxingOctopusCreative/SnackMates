"use client";

import { Icon } from "@iconify/react";
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
} from "@adobe/react-spectrum";
import { CreateWishlistModal } from "@/components/CreateWishlistModal";
import { PageHero } from "@/components/PageHero";
import { api, getToken, Wishlist } from "@/lib/api";

export default function WishlistsPage() {
  return <WishlistsContent />;
}

function WishlistsContent() {
  const [lists, setLists] = useState<Wishlist[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [addHovered, setAddHovered] = useState(false);

  async function load() {
    setLists(await api.wishlists(getToken()));
  }

  useEffect(() => {
    load();
  }, []);

  return (
    <Flex direction="column" gap="size-300">
      <PageHero title="Your Wishlists" ariaLabel="Your wishlists">
        <Flex direction="column" gap="size-75">
          <Flex
            alignItems="center"
            justifyContent="space-between"
            gap="size-100"
            wrap
            UNSAFE_className="sm-page-hero__title-row"
          >
            <Heading level={1} margin={0}>
              Your Wishlists
            </Heading>
            <button
              type="button"
              className="sm-page-hero__edit-button sm-page-hero__add-button"
              aria-label="Create Wishlist"
              title="Create Wishlist"
              onClick={() => setCreateOpen(true)}
              onMouseEnter={() => setAddHovered(true)}
              onMouseLeave={() => setAddHovered(false)}
              onFocus={() => setAddHovered(true)}
              onBlur={() => setAddHovered(false)}
            >
              <Icon
                icon={addHovered ? "ion:add" : "ion:add-outline"}
                className="sm-page-hero__edit-icon"
                aria-hidden
              />
            </button>
          </Flex>
        </Flex>
      </PageHero>

      <div className="sm-wishlists-table-panel">
        <TableView aria-label="Wishlists" selectionMode="none">
          <TableHeader>
            <Column key="title">Title</Column>
            <Column key="items">Items</Column>
            <Column key="public">Public</Column>
          </TableHeader>
          <TableBody items={lists}>
            {(item) => (
              <Row key={item.id}>
                <Cell>
                  <Link href={`/wishlists/${item.slug}`} className="sm-wishlist-title-link">
                    {item.title}
                  </Link>
                </Cell>
                <Cell>{item.item_count}</Cell>
                <Cell>{item.is_public ? "Yes" : "No"}</Cell>
              </Row>
            )}
          </TableBody>
        </TableView>
      </div>

      <CreateWishlistModal
        isOpen={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={load}
      />
    </Flex>
  );
}
