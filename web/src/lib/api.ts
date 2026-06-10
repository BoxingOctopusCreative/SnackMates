const DEFAULT_API_URL = "http://localhost:8080";

function resolveApiUrl(): string {
  const configured = process.env.NEXT_PUBLIC_API_URL ?? DEFAULT_API_URL;

  // NEXT_PUBLIC_* values are baked at build time. When a production image was built
  // with localhost defaults, use same-origin relative API paths in the browser.
  if (
    typeof window !== "undefined" &&
    configured.includes("localhost") &&
    window.location.hostname !== "localhost"
  ) {
    return "";
  }

  return configured;
}

export const API_URL = resolveApiUrl();

export type PublicUser = {
  id: string;
  username: string;
  display_name: string;
  bio: string;
  country: string;
  avatar_url?: string;
  banner_url?: string;
};

export type FriendshipView = {
  id?: string;
  status: "none" | "friends" | "pending_outgoing" | "pending_incoming" | "declined";
};

export type Friendship = {
  id: string;
  requester_id: string;
  addressee_id: string;
  status: string;
  created_at: string;
  updated_at: string;
  user?: PublicUser;
};

export type UserProfile = {
  user: PublicUser;
  wishlists: Wishlist[];
  friendship?: FriendshipView;
};

export type User = {
  id: string;
  username: string;
  email: string;
  email_verified: boolean;
  display_name: string;
  bio: string;
  country: string;
  avatar_key?: string;
  avatar_url?: string;
  banner_key?: string;
  banner_url?: string;
  discord_id?: string;
  discord_linked?: boolean;
  totp_enabled: boolean;
  has_webauthn: boolean;
  created_at: string;
};

export type Wishlist = {
  id: string;
  user_id: string;
  slug: string;
  title: string;
  description: string;
  is_public: boolean;
  item_count: number;
  banner_url?: string;
  created_at: string;
  updated_at: string;
};

export type SnagDeliveryMethod = "in_person" | "mail";

export type SnaggedBy = {
  id: string;
  display_name: string;
  delivery_method: SnagDeliveryMethod;
  tracking_number?: string;
};

export type WishlistItem = {
  id: string;
  wishlist_id: string;
  name: string;
  type: string;
  brand: string;
  notes: string;
  image_url?: string;
  snagged_by?: SnaggedBy;
  created_at: string;
  updated_at: string;
};

export type WishlistDetail = {
  wishlist: Wishlist;
  items: WishlistItem[];
  viewer_can_snag: boolean;
};

export type FriendWishlist = Wishlist & {
  owner: {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
  };
};

export type Message = {
  id: string;
  conversation_id: string;
  sender_id: string;
  subject: string;
  body: string;
  created_at: string;
  read_at?: string;
};

export type ChatMessage = {
  id: string;
  chat_id: string;
  sender_id: string;
  body: string;
  created_at: string;
  read_at?: string;
};

export type Chat = {
  id: string;
  user_a_id: string;
  user_b_id: string;
  created_at: string;
  updated_at: string;
  other_user?: PublicUser;
  last_message?: ChatMessage;
  unread_count: number;
};

export type Conversation = {
  id: string;
  user_a_id: string;
  user_b_id: string;
  created_at: string;
  updated_at: string;
  other_user?: PublicUser;
  last_message?: Message;
  unread_count: number;
};

export type SnackMatch = {
  id: string;
  user_a_id: string;
  user_b_id: string;
  status: string;
  matched_at: string;
  mate?: User;
};

export type UserSearchHit = {
  id: string;
  username: string;
  display_name: string;
  bio: string;
  country: string;
  avatar_url?: string;
};

export type WishlistItemSearchHit = {
  id: string;
  name: string;
  type: string;
  brand: string;
  notes: string;
  score?: number;
  user_id: string;
  user_name: string;
  username: string;
  wishlist_slug: string;
  wishlist_title: string;
};

export type ProductSearchHit = {
  code: string;
  name: string;
  brand: string;
  type: string;
  categories?: string;
  image_url?: string;
  quantity?: string;
  relevance?: number;
};

export type SearchResponse = {
  query: string;
  search_terms: string;
  ai_assisted: boolean;
  users: UserSearchHit[];
  wishlist_items: WishlistItemSearchHit[];
  products: ProductSearchHit[];
};

export type SearchTab = "all" | "people" | "wishlists" | "products";

export class ApiError extends Error {
  status: number;
  code?: string;
  constructor(status: number, message: string, code?: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  token?: string | null,
): Promise<T> {
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type") && !(options.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new ApiError(res.status, data.error ?? res.statusText, data.code);
  }
  return data as T;
}

export const api = {
  register(body: { email: string; password: string; display_name: string; country?: string }) {
    return request<{ user_id: string; message: string }>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify(body),
    });
  },
  login(body: { email: string; password: string; totp_code?: string }) {
    return request<{ token?: string; mfa_required?: boolean; methods?: string[] }>(
      "/api/v1/auth/login",
      { method: "POST", body: JSON.stringify(body) },
    );
  },
  logout(token?: string | null) {
    return request<{ ok: boolean }>("/api/v1/auth/logout", { method: "POST" }, token);
  },
  me(token?: string | null) {
    return request<User>("/api/v1/auth/me", {}, token);
  },
  verifyEmail(token: string) {
    return request<{ ok: boolean }>("/api/v1/auth/verify-email", {
      method: "POST",
      body: JSON.stringify({ token }),
    });
  },
  forgotPassword(email: string) {
    return request<{ message: string }>("/api/v1/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
  },
  resetPassword(token: string, new_password: string) {
    return request<{ ok: boolean }>("/api/v1/auth/reset-password", {
      method: "POST",
      body: JSON.stringify({ token, new_password }),
    });
  },
  requestAccountDeactivate(token?: string | null) {
    return request<{ message: string }>(
      "/api/v1/auth/account/deactivate/request",
      { method: "POST" },
      token,
    );
  },
  requestAccountDelete(token?: string | null) {
    return request<{ message: string }>(
      "/api/v1/auth/account/delete/request",
      { method: "POST" },
      token,
    );
  },
  requestAccountReactivate(email: string) {
    return request<{ message: string }>("/api/v1/auth/account/reactivate/request", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
  },
  confirmAccountAction(token: string) {
    return request<{ ok: boolean; action: string }>("/api/v1/auth/account/confirm", {
      method: "POST",
      body: JSON.stringify({ token }),
    });
  },
  discordUrl() {
    return `${API_URL}/api/v1/auth/discord`;
  },
  discordConnect(token?: string | null) {
    return request<{ url: string }>("/api/v1/auth/discord/connect", {}, token);
  },
  wishlists(token?: string | null) {
    return request<Wishlist[]>("/api/v1/wishlists/", {}, token).then((data) => data ?? []);
  },
  friendWishlists(token?: string | null) {
    return request<FriendWishlist[]>("/api/v1/wishlists/friends", {}, token).then((data) => data ?? []);
  },
  createWishlist(
    body: { title: string; description: string; is_public: boolean },
    token?: string | null,
  ) {
    return request<Wishlist>("/api/v1/wishlists/", { method: "POST", body: JSON.stringify(body) }, token);
  },
  getWishlist(slug: string, token?: string | null) {
    return request<WishlistDetail>(`/api/v1/wishlists/${slug}`, {}, token).then((data) => ({
      ...data,
      items: data.items ?? [],
      viewer_can_snag: data.viewer_can_snag ?? false,
    }));
  },
  updateWishlist(
    slug: string,
    body: { title: string; description: string; is_public: boolean },
    token?: string | null,
  ) {
    return request<{ ok: boolean; slug: string }>(
      `/api/v1/wishlists/${slug}`,
      { method: "PUT", body: JSON.stringify(body) },
      token,
    );
  },
  addItem(
    wishlistId: string,
    body: {
      name: string;
      type?: string;
      brand?: string;
      notes?: string;
      image_url?: string;
    },
    token?: string | null,
  ) {
    return request<WishlistItem>(
      `/api/v1/wishlists/${wishlistId}/items`,
      { method: "POST", body: JSON.stringify(body) },
      token,
    );
  },
  deleteItem(wishlistId: string, itemId: string, token?: string | null) {
    return request<{ ok: boolean }>(
      `/api/v1/wishlists/${wishlistId}/items/${itemId}`,
      { method: "DELETE" },
      token,
    );
  },
  snagItem(
    wishlistSlug: string,
    itemId: string,
    delivery: { delivery_method: SnagDeliveryMethod; tracking_number?: string },
    token?: string | null,
  ) {
    return request<WishlistItem>(
      `/api/v1/wishlists/${wishlistSlug}/items/${itemId}/snag`,
      {
        method: "POST",
        body: JSON.stringify(delivery),
      },
      token,
    );
  },
  uploadWishlistBanner(wishlistId: string, file: File, token?: string | null) {
    const form = new FormData();
    form.append("banner", file);
    return request<{ banner_url: string }>(
      `/api/v1/wishlists/${wishlistId}/banner`,
      { method: "POST", body: form },
      token,
    );
  },
  setWishlistBannerUrl(wishlistId: string, banner_url: string, token?: string | null) {
    return request<{ banner_url: string | null }>(
      `/api/v1/wishlists/${wishlistId}/banner`,
      { method: "PUT", body: JSON.stringify({ banner_url }) },
      token,
    );
  },
  matches(token?: string | null) {
    return request<SnackMatch[]>("/api/v1/matches/", {}, token).then((data) => data ?? []);
  },
  runPairing(token?: string | null) {
    return request<{ paired: number; matches: SnackMatch[] }>(
      "/api/v1/matches/run",
      { method: "POST" },
      token,
    );
  },
  notifications(token?: string | null) {
    return request<{
      unread_count: number;
      items: {
        id: string;
        type: "snack_mate_request" | "snack_mate_accepted";
        created_at: string;
        friendship: Friendship;
      }[];
    }>("/api/v1/notifications", {}, token).then((data) => ({
      unread_count: data.unread_count ?? 0,
      items: data.items ?? [],
    }));
  },
  acknowledgeNotification(friendshipId: string, token?: string | null) {
    return request<{ ok: boolean }>(
      `/api/v1/notifications/${friendshipId}/acknowledge`,
      { method: "POST" },
      token,
    );
  },
  search(q: string) {
    return request<SearchResponse>(`/api/v1/search?q=${encodeURIComponent(q)}`).then((data) => ({
      ...data,
      users: data.users ?? [],
      wishlist_items: data.wishlist_items ?? [],
      products: data.products ?? [],
    }));
  },
  getUserProfile(username: string, token?: string | null) {
    return request<UserProfile>(`/api/v1/users/${username}`, {}, token).then((data) => ({
      ...data,
      wishlists: data.wishlists ?? [],
    }));
  },
  friends(token?: string | null) {
    return request<Friendship[]>("/api/v1/friends/", {}, token).then((data) => data ?? []);
  },
  friendRequests(token?: string | null) {
    return request<Friendship[]>("/api/v1/friends/requests", {}, token).then((data) => data ?? []);
  },
  requestFriend(username: string, token?: string | null) {
    return request<Friendship>(
      "/api/v1/friends/request",
      { method: "POST", body: JSON.stringify({ username }) },
      token,
    );
  },
  acceptFriend(friendshipId: string, token?: string | null) {
    return request<Friendship>(
      `/api/v1/friends/${friendshipId}/accept`,
      { method: "POST" },
      token,
    );
  },
  declineFriend(friendshipId: string, token?: string | null) {
    return request<{ ok: boolean }>(
      `/api/v1/friends/${friendshipId}/decline`,
      { method: "POST" },
      token,
    );
  },
  removeFriend(friendshipId: string, token?: string | null) {
    return request<{ ok: boolean }>(
      `/api/v1/friends/${friendshipId}`,
      { method: "DELETE" },
      token,
    );
  },
  conversations(token?: string | null) {
    return request<{
      unread_count: number;
      conversations: Conversation[];
    }>("/api/v1/messages/conversations", {}, token).then((data) => ({
      unread_count: data.unread_count ?? 0,
      conversations: data.conversations ?? [],
    }));
  },
  startConversation(username: string, token?: string | null) {
    return request<{ id: string }>(
      "/api/v1/messages/conversations",
      { method: "POST", body: JSON.stringify({ username }) },
      token,
    );
  },
  getConversation(conversationId: string, token?: string | null) {
    return request<{
      conversation: Conversation;
      messages: Message[];
    }>(`/api/v1/messages/conversations/${conversationId}`, {}, token).then((data) => ({
      conversation: data.conversation,
      messages: data.messages ?? [],
    }));
  },
  sendMessage(conversationId: string, subject: string, body: string, token?: string | null) {
    return request<Message>(
      `/api/v1/messages/conversations/${conversationId}/messages`,
      { method: "POST", body: JSON.stringify({ subject, body }) },
      token,
    );
  },
  chats(token?: string | null) {
    return request<{
      unread_count: number;
      chats: Chat[];
    }>("/api/v1/chats", {}, token).then((data) => ({
      unread_count: data.unread_count ?? 0,
      chats: data.chats ?? [],
    }));
  },
  startChat(username: string, token?: string | null) {
    return request<{ id: string }>(
      "/api/v1/chats",
      { method: "POST", body: JSON.stringify({ username }) },
      token,
    );
  },
  getChat(chatId: string, token?: string | null) {
    return request<{
      chat: Chat;
      messages: ChatMessage[];
    }>(`/api/v1/chats/${chatId}`, {}, token).then((data) => ({
      chat: data.chat,
      messages: data.messages ?? [],
    }));
  },
  sendChatMessage(chatId: string, body: string, token?: string | null) {
    return request<ChatMessage>(
      `/api/v1/chats/${chatId}/messages`,
      { method: "POST", body: JSON.stringify({ body }) },
      token,
    );
  },
  markChatRead(chatId: string, token?: string | null) {
    return request<{ ok: boolean }>(
      `/api/v1/chats/${chatId}/read`,
      { method: "POST" },
      token,
    );
  },
  markConversationRead(conversationId: string, token?: string | null) {
    return request<{ ok: boolean }>(
      `/api/v1/messages/conversations/${conversationId}/read`,
      { method: "POST" },
      token,
    );
  },
  updateProfile(
    body: { display_name: string; bio: string; country: string },
    token?: string | null,
  ) {
    return request<{ ok: boolean }>("/api/v1/users/me", { method: "PUT", body: JSON.stringify(body) }, token);
  },
  uploadAvatar(file: File, token?: string | null) {
    const form = new FormData();
    form.append("avatar", file);
    return request<{ avatar_url: string }>("/api/v1/users/me/avatar", { method: "POST", body: form }, token);
  },
  uploadBanner(file: File, token?: string | null) {
    const form = new FormData();
    form.append("banner", file);
    return request<{ banner_url: string }>("/api/v1/users/me/banner", { method: "POST", body: form }, token);
  },
  setBannerUrl(banner_url: string, token?: string | null) {
    return request<{ banner_url: string | null }>(
      "/api/v1/users/me/banner",
      { method: "PUT", body: JSON.stringify({ banner_url }) },
      token,
    );
  },
  totpSetup(token?: string | null) {
    return request<{ secret: string; otpauth_url: string }>(
      "/api/v1/auth/mfa/totp/setup",
      { method: "POST" },
      token,
    );
  },
  totpEnable(code: string, token?: string | null) {
    return request<{ ok: boolean }>(
      "/api/v1/auth/mfa/totp/enable",
      { method: "POST", body: JSON.stringify({ code }) },
      token,
    );
  },
  totpDisable(code: string, token?: string | null) {
    return request<{ ok: boolean }>(
      "/api/v1/auth/mfa/totp/disable",
      { method: "POST", body: JSON.stringify({ code }) },
      token,
    );
  },
  webauthnRegisterBegin(token?: string | null) {
    return request<{
      options: import("@/lib/webauthn").WebAuthnCreationOptionsPayload;
      session_data: string;
    }>("/api/v1/auth/mfa/webauthn/register/begin", { method: "POST" }, token);
  },
  webauthnRegisterFinish(credential: ArrayBuffer, sessionData: string, token?: string | null) {
    return fetch(`${API_URL}/api/v1/auth/mfa/webauthn/register/finish`, {
      method: "POST",
      headers: {
        Authorization: token ? `Bearer ${token}` : "",
        "Content-Type": "application/json",
        "X-WebAuthn-Session": sessionData,
      },
      credentials: "include",
      body: credential,
    });
  },
};

export function discordConnect(token?: string | null) {
  return api.discordConnect(token);
}

export function discordUrl() {
  return `${API_URL}/api/v1/auth/discord`;
}

export function saveToken(token: string) {
  if (typeof window !== "undefined") {
    localStorage.setItem("snackmates_token", token);
  }
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("snackmates_token");
}

export function clearToken() {
  if (typeof window !== "undefined") {
    localStorage.removeItem("snackmates_token");
  }
}
