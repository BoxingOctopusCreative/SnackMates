import type {
  Friendship,
  ProductSearchHit,
  PublicUser,
  SnackMatch,
  User,
  Wishlist,
  WishlistItem,
} from "@/lib/api";

export const mockUser: User = {
  id: "user-1",
  username: "snackfan",
  email: "snacker@example.com",
  email_verified: true,
  display_name: "Snack Fan",
  bio: "Loves chips",
  country: "US",
  avatar_url: "https://example.com/avatar.jpg",
  banner_url: "https://example.com/banner.jpg",
  discord_linked: false,
  totp_enabled: false,
  has_webauthn: false,
  created_at: "2026-01-01T00:00:00Z",
};

export const mockPublicUser: PublicUser = {
  id: "user-1",
  username: "snackfan",
  display_name: "Snack Fan",
  bio: "Loves chips",
  country: "US",
  avatar_url: "https://example.com/avatar.jpg",
  banner_url: "https://example.com/banner.jpg",
};

export const mockWishlist: Wishlist = {
  id: "wishlist-1",
  user_id: "user-1",
  slug: "sweet-treats",
  title: "Sweet treats",
  description: "Candy and cookies",
  is_public: true,
  item_count: 2,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

export const mockWishlistItem: WishlistItem = {
  id: "item-1",
  wishlist_id: "wishlist-1",
  name: "Pocky",
  type: "Candy",
  brand: "Glico",
  notes: "Strawberry",
  image_url: "https://images.openfoodfacts.org/images/products/123/front_en.jpg",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

export const mockSnackMate: Friendship = {
  id: "friendship-1",
  requester_id: "user-1",
  addressee_id: "user-2",
  status: "accepted",
  created_at: "2026-01-15T00:00:00Z",
  updated_at: "2026-01-15T00:00:00Z",
  user: {
    id: "user-2",
    username: "matesnacker",
    display_name: "Mate Snacker",
    bio: "From abroad",
    country: "JP",
  },
};

export const mockMatch: SnackMatch = {
  id: "match-1",
  user_a_id: "user-1",
  user_b_id: "user-2",
  status: "active",
  matched_at: "2026-01-15T00:00:00Z",
  mate: {
    id: "user-2",
    username: "matesnacker",
    email: "mate@example.com",
    email_verified: true,
    display_name: "Mate Snacker",
    bio: "From abroad",
    country: "JP",
    totp_enabled: false,
    has_webauthn: false,
    created_at: "2026-01-01T00:00:00Z",
  },
};

export const mockProductSearchHit: ProductSearchHit = {
  code: "0059800000215",
  name: "Coffee Crisp",
  brand: "Nestlé",
  type: "Candy",
  categories: "Chocolate stuffed wafers",
  image_url: "https://example.com/coffee-crisp.jpg",
  quantity: "50 g",
  relevance: 0.95,
};

export const mockUserSearchHit = {
  id: "user-2",
  username: "snackfan",
  display_name: "Snack Fan",
  bio: "Loves Japanese snacks",
  country: "JP",
};

export const mockWishlistItemSearchHit = {
  id: "item-9",
  name: "Pocky",
  type: "Candy",
  brand: "Glico",
  notes: "",
  user_id: "user-2",
  user_name: "Snack Fan",
  username: "snackfan",
  wishlist_slug: "tokyo-treats",
  wishlist_title: "Tokyo Treats",
};

export const mockSearchResponse = {
  query: "coffee crisp",
  search_terms: "nestle coffee crisp chocolate bar",
  ai_assisted: true,
  users: [mockUserSearchHit],
  wishlist_items: [mockWishlistItemSearchHit],
  products: [mockProductSearchHit],
};
