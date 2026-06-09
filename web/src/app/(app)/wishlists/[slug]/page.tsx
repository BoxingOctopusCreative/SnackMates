"use client";

import { Icon } from "@iconify/react";
import Image from "next/image";
import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button, Flex, Text } from "@adobe/react-spectrum";
import { AddWishlistSnackModal } from "@/components/AddWishlistSnackModal";
import { useAuth } from "@/components/AuthGate";
import { SnagItemModal } from "@/components/SnagItemModal";
import { WishlistHero } from "@/components/WishlistHero";
import { api, getToken, Wishlist, WishlistItem } from "@/lib/api";
import { snagStatusLabel } from "@/lib/snag-delivery";

type SortColumn = "name" | "type" | "brand" | "notes";
type SortDirection = "asc" | "desc";

const SORTABLE_COLUMNS: { key: SortColumn; label: string }[] = [
  { key: "name", label: "Snack Name" },
  { key: "type", label: "Type" },
  { key: "brand", label: "Brand" },
  { key: "notes", label: "Notes" },
];

function RemoveItemButton({ onPress }: { onPress: () => void }) {
  const [hovered, setHovered] = useState(false);

  return (
    <button
      type="button"
      className="sm-wishlist-item-remove-button"
      aria-label="Remove"
      title="Remove"
      onClick={onPress}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocus={() => setHovered(true)}
      onBlur={() => setHovered(false)}
    >
      <Icon
        icon={hovered ? "ion:trash" : "ion:trash-outline"}
        className="sm-wishlist-item-remove-icon"
        aria-hidden
      />
    </button>
  );
}

function sortWishlistItems(
  items: WishlistItem[],
  column: SortColumn,
  direction: SortDirection,
): WishlistItem[] {
  return [...items].sort((a, b) => {
    const result = a[column].localeCompare(b[column], undefined, { sensitivity: "base" });
    return direction === "asc" ? result : -result;
  });
}

export default function WishlistDetailPage() {
  return <WishlistDetail />;
}

function WishlistDetail() {
  const { user } = useAuth();
  const router = useRouter();
  const params = useParams<{ slug: string }>();
  const slug = params.slug;
  const [wishlist, setWishlist] = useState<Wishlist | null>(null);
  const [items, setItems] = useState<WishlistItem[]>([]);
  const [viewerCanSnag, setViewerCanSnag] = useState(false);
  const [addSnackOpen, setAddSnackOpen] = useState(false);
  const [snagItemTarget, setSnagItemTarget] = useState<WishlistItem | null>(null);
  const [sortColumn, setSortColumn] = useState<SortColumn>("name");
  const [sortDirection, setSortDirection] = useState<SortDirection>("asc");

  const sortedItems = useMemo(
    () => sortWishlistItems(items, sortColumn, sortDirection),
    [items, sortColumn, sortDirection],
  );

  function toggleSort(column: SortColumn) {
    if (sortColumn === column) {
      setSortDirection((direction) => (direction === "asc" ? "desc" : "asc"));
      return;
    }

    setSortColumn(column);
    setSortDirection("asc");
  }

  function sortAriaValue(column: SortColumn): "ascending" | "descending" | "none" {
    if (sortColumn !== column) return "none";
    return sortDirection === "asc" ? "ascending" : "descending";
  }

  async function load(wishlistSlug: string) {
    const data = await api.getWishlist(wishlistSlug, getToken());
    setWishlist(data.wishlist);
    setItems(data.items);
    setViewerCanSnag(data.viewer_can_snag);
  }

  useEffect(() => {
    load(slug);
  }, [slug]);

  async function removeItem(itemId: string) {
    await api.deleteItem(slug, itemId, getToken());
    await load(slug);
  }

  async function confirmSnagItem(delivery: {
    delivery_method: "in_person" | "mail";
    tracking_number?: string;
  }) {
    if (!snagItemTarget) return;
    await api.snagItem(slug, snagItemTarget.id, delivery, getToken());
    setSnagItemTarget(null);
    await load(slug);
  }

  function handleDetailsChange(patch: { title?: string; description?: string; slug?: string }) {
    setWishlist((current) => (current ? { ...current, ...patch } : current));
    if (patch.slug && patch.slug !== slug) {
      router.replace(`/wishlists/${patch.slug}`);
    }
  }

  if (!wishlist) return <Text>Loading...</Text>;

  const isOwner = user.id === wishlist.user_id;

  return (
    <Flex direction="column" gap="size-300">
      <WishlistHero
        wishlistSlug={slug}
        title={wishlist.title}
        description={wishlist.description}
        isPublic={wishlist.is_public}
        bannerUrl={wishlist.banner_url}
        isOwner={isOwner}
        onAddSnack={isOwner ? () => setAddSnackOpen(true) : undefined}
        onBannerChange={(url) => setWishlist({ ...wishlist, banner_url: url })}
        onDetailsChange={handleDetailsChange}
      />

      {isOwner && (
        <AddWishlistSnackModal
          isOpen={addSnackOpen}
          onClose={() => setAddSnackOpen(false)}
          wishlistSlug={slug}
          onAdded={() => load(slug)}
        />
      )}

      {viewerCanSnag && (
        <SnagItemModal
          isOpen={snagItemTarget !== null}
          snackName={snagItemTarget?.name}
          onClose={() => setSnagItemTarget(null)}
          onConfirm={confirmSnagItem}
        />
      )}

      <div className="sm-wishlist-items-table-panel">
        <div className="sm-wishlist-items-table" role="table" aria-label="Wishlist items">
          <div role="rowgroup">
            <div
              role="row"
              className="sm-wishlist-items-table__row sm-wishlist-items-table__row--header"
            >
              <span role="columnheader">Image</span>
              {SORTABLE_COLUMNS.map(({ key, label }) => (
                <span key={key} role="columnheader" aria-sort={sortAriaValue(key)}>
                  <button
                    type="button"
                    className="sm-wishlist-items-table__sort-button"
                    onClick={() => toggleSort(key)}
                  >
                    {label}
                    <span className="sm-wishlist-items-table__sort-indicator" aria-hidden="true">
                      {sortColumn === key ? (sortDirection === "asc" ? "↑" : "↓") : "↕"}
                    </span>
                  </button>
                </span>
              ))}
              <span role="columnheader">Status</span>
              <span role="columnheader">Actions</span>
            </div>
          </div>
          <div role="rowgroup">
            {items.length === 0 ? (
              <div role="row" className="sm-wishlist-items-table__row">
                <span role="cell" className="sm-wishlist-items-table__empty">
                  No snacks on this wishlist yet.
                </span>
              </div>
            ) : (
              sortedItems.map((item) => (
                <div role="row" className="sm-wishlist-items-table__row" key={item.id}>
                  <span role="cell" className="sm-wishlist-items-table__image-cell">
                    {item.image_url ? (
                      <Image
                        src={item.image_url}
                        alt={item.name}
                        width={80}
                        height={80}
                        className="sm-wishlist-item-image"
                        unoptimized
                      />
                    ) : (
                      "—"
                    )}
                  </span>
                  <span role="cell">{item.name}</span>
                  <span role="cell">{item.type || "-"}</span>
                  <span role="cell">{item.brand || "-"}</span>
                  <span role="cell">{item.notes || "-"}</span>
                  <span role="cell" className="sm-wishlist-items-table__status">
                    {item.snagged_by ? (
                      <span className="sm-wishlist-item-snagged">
                        {snagStatusLabel(item.snagged_by)}
                      </span>
                    ) : (
                      "—"
                    )}
                  </span>
                  <span role="cell" className="sm-wishlist-items-table__actions">
                    {isOwner ? (
                      <RemoveItemButton onPress={() => removeItem(item.id)} />
                    ) : viewerCanSnag && !item.snagged_by ? (
                      <Button variant="accent" onPress={() => setSnagItemTarget(item)}>
                        Mark snagged
                      </Button>
                    ) : null}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </Flex>
  );
}
