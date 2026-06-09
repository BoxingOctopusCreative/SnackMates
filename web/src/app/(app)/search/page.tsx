"use client";

import Image from "next/image";
import Link from "next/link";
import { Suspense, useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  Avatar,
  Button,
  Flex,
  Form,
  Heading,
  ProgressCircle,
  SearchField,
  Text,
  View,
} from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { PageHero } from "@/components/PageHero";
import { AddToWishlistButton } from "@/components/AddToWishlistButton";
import {
  api,
  ProductSearchHit,
  SearchResponse,
  SearchTab,
  UserSearchHit,
  WishlistItemSearchHit,
} from "@/lib/api";

const TABS: { id: SearchTab; label: string }[] = [
  { id: "all", label: "All" },
  { id: "people", label: "People" },
  { id: "wishlists", label: "Wishlists" },
  { id: "products", label: "Products" },
];

const ALL_PREVIEW = {
  people: 3,
  wishlists: 4,
  products: 6,
} as const;

function parseTab(value: string | null): SearchTab {
  if (value === "people" || value === "wishlists" || value === "products") {
    return value;
  }
  return "all";
}

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <View UNSAFE_style={{ display: "grid", placeItems: "center", minHeight: 200 }}>
          <ProgressCircle isIndeterminate aria-label="Loading search" />
        </View>
      }
    >
      <SearchContent />
    </Suspense>
  );
}

function SearchContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const urlQuery = searchParams.get("q")?.trim() ?? "";
  const activeTab = parseTab(searchParams.get("type"));
  const [query, setQuery] = useState(urlQuery);
  const [data, setData] = useState<SearchResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searched, setSearched] = useState(false);

  async function runSearch(q: string) {
    setLoading(true);
    setError(null);
    try {
      const response = await api.search(q);
      setData(response);
      setSearched(true);
    } catch (err) {
      setData(null);
      setSearched(true);
      setError(err instanceof Error ? err.message : "Search failed");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    setQuery(urlQuery);
    if (!urlQuery) {
      setData(null);
      setSearched(false);
      setError(null);
      return;
    }
    runSearch(urlQuery);
  }, [urlQuery]);

  function buildSearchUrl(nextQuery: string, tab: SearchTab = activeTab) {
    const params = new URLSearchParams();
    params.set("q", nextQuery);
    if (tab !== "all") {
      params.set("type", tab);
    }
    return `/search?${params.toString()}`;
  }

  async function search(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = query.trim();
    if (!trimmed) return;
    router.replace(buildSearchUrl(trimmed));
  }

  function setTab(tab: SearchTab) {
    if (!urlQuery) return;
    router.replace(buildSearchUrl(urlQuery, tab));
  }

  const totals = useMemo(
    () => ({
      people: data?.users.length ?? 0,
      wishlists: data?.wishlist_items.length ?? 0,
      products: data?.products.length ?? 0,
    }),
    [data],
  );

  const hasResults = totals.people + totals.wishlists + totals.products > 0;
  const emptyTabMessage =
    activeTab === "people"
      ? "No people found."
      : activeTab === "wishlists"
        ? "No wishlist snacks found."
        : activeTab === "products"
          ? "No products found."
          : null;

  return (
    <Flex direction="column" gap="size-300">
      <PageHero
        title="Search"
        description="Find people, public wishlist snacks, and packaged products across SnackMates."
      />
      <Form maxWidth="100%" onSubmit={search}>
        <Flex gap="size-200" alignItems="end">
          <SearchField label="Search" value={query} onChange={setQuery} width="size-4600" />
          <Button type="submit" variant="accent">Search</Button>
        </Flex>
      </Form>

      {(urlQuery || activeTab !== "all") && (
        <nav className="sm-search-tabs" aria-label="Search filters">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              type="button"
              className={`sm-search-tabs__tab${activeTab === tab.id ? " sm-search-tabs__tab--active" : ""}`}
              aria-current={activeTab === tab.id ? "page" : undefined}
              onClick={() => setTab(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      )}

      {!urlQuery && activeTab === "people" && (
        <Text>Search for people to send a snack mate request.</Text>
      )}

      {loading && (
        <View UNSAFE_style={{ display: "grid", placeItems: "center", minHeight: 120 }}>
          <ProgressCircle isIndeterminate aria-label="Searching" />
        </View>
      )}

      {error && !loading && <Text UNSAFE_style={{ color: "var(--spectrum-global-color-red-600)" }}>{error}</Text>}

      {searched && !loading && !error && !hasResults && activeTab === "all" && (
        <Text>No results found.</Text>
      )}

      {searched && !loading && !error && emptyTabMessage && !hasResultsForTab(activeTab, totals) && (
        <Text>{emptyTabMessage}</Text>
      )}

      {searched && !loading && !error && data && data.ai_assisted && data.search_terms && data.search_terms !== urlQuery && (
        <Text>
          Product matches used AI-assisted OpenFoodFacts search for &ldquo;{data.search_terms}&rdquo;.
        </Text>
      )}

      {!loading && data && (
        <>
          {(activeTab === "all" || activeTab === "people") && totals.people > 0 && (
            <SearchSection
              title="People"
              showViewAll={activeTab === "all" && totals.people > ALL_PREVIEW.people}
              onViewAll={() => setTab("people")}
            >
              {sliceForTab(data.users, activeTab, ALL_PREVIEW.people).map((user) => (
                <UserResult key={user.id} user={user} />
              ))}
            </SearchSection>
          )}

          {(activeTab === "all" || activeTab === "wishlists") && totals.wishlists > 0 && (
            <SearchSection
              title="Wishlists"
              showViewAll={activeTab === "all" && totals.wishlists > ALL_PREVIEW.wishlists}
              onViewAll={() => setTab("wishlists")}
            >
              {sliceForTab(data.wishlist_items, activeTab, ALL_PREVIEW.wishlists).map((item) => (
                <WishlistItemResult key={item.id} item={item} />
              ))}
            </SearchSection>
          )}

          {(activeTab === "all" || activeTab === "products") && totals.products > 0 && (
            <SearchSection
              title="Products"
              showViewAll={activeTab === "all" && totals.products > ALL_PREVIEW.products}
              onViewAll={() => setTab("products")}
            >
              <View
                UNSAFE_className="sm-search-results"
                UNSAFE_style={{
                  display: "grid",
                  gap: "var(--spectrum-global-dimension-size-200)",
                  gridTemplateColumns: activeTab === "all"
                    ? "repeat(auto-fill, minmax(240px, 1fr))"
                    : "repeat(auto-fill, minmax(240px, 1fr))",
                }}
              >
                {sliceForTab(data.products, activeTab, ALL_PREVIEW.products).map((item) => (
                  <ProductCard key={item.code} item={item} />
                ))}
              </View>
            </SearchSection>
          )}
        </>
      )}
    </Flex>
  );
}

function hasResultsForTab(
  tab: SearchTab,
  totals: { people: number; wishlists: number; products: number },
) {
  if (tab === "people") return totals.people > 0;
  if (tab === "wishlists") return totals.wishlists > 0;
  if (tab === "products") return totals.products > 0;
  return totals.people + totals.wishlists + totals.products > 0;
}

function sliceForTab<T>(items: T[], tab: SearchTab, previewLimit: number): T[] {
  if (tab === "all") {
    return items.slice(0, previewLimit);
  }
  return items;
}

function SearchSection({
  title,
  children,
  showViewAll,
  onViewAll,
}: {
  title: string;
  children: React.ReactNode;
  showViewAll?: boolean;
  onViewAll?: () => void;
}) {
  return (
    <section className="sm-search-section">
      <Flex justifyContent="space-between" alignItems="center" marginBottom="size-150">
        <Heading level={3} margin={0}>{title}</Heading>
        {showViewAll && onViewAll && (
          <Button variant="secondary" onPress={onViewAll}>View all</Button>
        )}
      </Flex>
      <Flex direction="column" gap="size-150">{children}</Flex>
    </section>
  );
}

function UserResult({ user }: { user: UserSearchHit }) {
  return (
    <Link href={`/users/${user.username}`} className="sm-search-result-link">
      <View
        backgroundColor="gray-50"
        borderWidth="thin"
        borderColor="gray-300"
        borderRadius="medium"
        padding="size-200"
      >
        <Flex gap="size-200" alignItems="center">
          <Avatar src={avatarImageSrc(user.avatar_url)} alt={user.display_name} size={48} />
          <Flex direction="column" gap="size-75">
            <Text UNSAFE_style={{ fontWeight: 700 }}>{user.display_name}</Text>
            <Text>@{user.username}</Text>
            {user.country && <Text>{user.country}</Text>}
            {user.bio && (
              <Text UNSAFE_style={{ color: "var(--spectrum-global-color-gray-700)" }}>
                {user.bio}
              </Text>
            )}
          </Flex>
        </Flex>
      </View>
    </Link>
  );
}

function WishlistItemResult({ item }: { item: WishlistItemSearchHit }) {
  return (
    <Link href={`/wishlists/${item.wishlist_slug}`} className="sm-search-result-link">
      <View
        backgroundColor="gray-50"
        borderWidth="thin"
        borderColor="gray-300"
        borderRadius="medium"
        padding="size-200"
      >
        <Flex direction="column" gap="size-75">
          <Text UNSAFE_style={{ fontWeight: 700 }}>{item.name}</Text>
          <Text>
            {item.type}
            {item.brand ? ` · ${item.brand}` : ""}
          </Text>
          <Text UNSAFE_style={{ color: "var(--spectrum-global-color-gray-700)" }}>
            On {item.wishlist_title || "wishlist"} by {item.user_name}
          </Text>
        </Flex>
      </View>
    </Link>
  );
}

function ProductCard({ item }: { item: ProductSearchHit }) {
  return (
    <View
      backgroundColor="gray-50"
      borderWidth="thin"
      borderColor="gray-300"
      borderRadius="medium"
      padding="size-200"
      UNSAFE_style={{ display: "flex", flexDirection: "column", gap: "0.75rem", height: "100%" }}
    >
      <View
        UNSAFE_style={{
          position: "relative",
          width: "100%",
          aspectRatio: "4 / 3",
          backgroundColor: "var(--sm-bg)",
          borderRadius: "8px",
          overflow: "hidden",
        }}
      >
        {item.image_url ? (
          <Image
            src={item.image_url}
            alt={item.name}
            fill
            sizes="240px"
            style={{ objectFit: "contain" }}
          />
        ) : (
          <View UNSAFE_style={{ display: "grid", placeItems: "center", height: "100%" }}>
            <Text>No image</Text>
          </View>
        )}
      </View>
      <Flex direction="column" gap="size-75">
        <Text UNSAFE_style={{ fontWeight: 700 }}>{item.name}</Text>
        {item.brand && <Text>{item.brand}</Text>}
        <Text>Type: {item.type || "-"}</Text>
        {item.quantity && <Text>{item.quantity}</Text>}
        {item.categories && (
          <Text UNSAFE_style={{ color: "var(--spectrum-global-color-gray-700)" }}>
            {item.categories}
          </Text>
        )}
      </Flex>
      <AddToWishlistButton item={item} />
    </View>
  );
}
