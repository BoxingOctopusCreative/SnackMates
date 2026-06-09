"use client";

import { useState } from "react";
import { Button, Flex, Item, Menu, MenuTrigger, Text } from "@adobe/react-spectrum";
import { Key } from "@react-types/shared";
import { api, getToken, ProductSearchHit, Wishlist } from "@/lib/api";
import { normalizeSnackType } from "@/lib/snack-types";

function buildNotes(item: ProductSearchHit): string {
  const parts: string[] = [];
  if (item.code) parts.push(`Barcode: ${item.code}`);
  if (item.categories) parts.push(item.categories);
  if (item.quantity) parts.push(item.quantity);
  return parts.join(" · ");
}

export function AddToWishlistButton({ item }: { item: ProductSearchHit }) {
  const [wishlists, setWishlists] = useState<Wishlist[] | null>(null);
  const [loadingLists, setLoadingLists] = useState(false);
  const [addingTo, setAddingTo] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  function handleOpenChange(isOpen: boolean) {
    if (!isOpen || wishlists !== null || loadingLists) return;

    setLoadingLists(true);
    setError(null);
    api
      .wishlists(getToken())
      .then(setWishlists)
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Could not load wishlists");
        setWishlists([]);
      })
      .finally(() => setLoadingLists(false));
  }

  async function handleAction(key: Key) {
    const slug = String(key);
    setAddingTo(slug);
    setError(null);
    setMessage(null);
    try {
      await api.addItem(
        slug,
        {
          name: item.name,
          type: normalizeSnackType(item.type),
          brand: item.brand,
          notes: buildNotes(item),
          image_url: item.image_url,
        },
        getToken(),
      );
      const list = wishlists?.find((w) => w.slug === slug);
      setMessage(`Added to ${list?.title ?? "wishlist"}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not add item");
    } finally {
      setAddingTo(null);
    }
  }

  return (
    <Flex direction="column" gap="size-75">
      <MenuTrigger onOpenChange={handleOpenChange}>
        <Button variant="accent" isDisabled={addingTo !== null}>
          Add to wishlist
        </Button>
        <Menu
          onAction={handleAction}
          disabledKeys={[
            ...(addingTo ? [addingTo] : []),
            ...(loadingLists || wishlists === null ? ["loading"] : []),
            ...(wishlists?.length === 0 ? ["empty"] : []),
          ]}
        >
          {loadingLists || wishlists === null ? (
            <Item key="loading">Loading…</Item>
          ) : wishlists.length === 0 ? (
            <Item key="empty">No wishlists yet</Item>
          ) : (
            wishlists.map((list) => <Item key={list.slug}>{list.title}</Item>)
          )}
        </Menu>
      </MenuTrigger>
      {message && (
        <Text UNSAFE_style={{ color: "var(--sm-highlight)" }}>{message}</Text>
      )}
      {error && (
        <Text UNSAFE_style={{ color: "var(--spectrum-global-color-red-600)" }}>{error}</Text>
      )}
    </Flex>
  );
}
