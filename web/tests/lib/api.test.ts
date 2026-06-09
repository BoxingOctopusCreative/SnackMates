import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  API_URL,
  ApiError,
  api,
  clearToken,
  discordUrl,
  getToken,
  saveToken,
} from "@/lib/api";
import { mockFetchJson } from "@test/utils";

describe("api token helpers", () => {
  it("stores and retrieves the auth token", () => {
    saveToken("abc123");
    expect(getToken()).toBe("abc123");
  });

  it("clears the auth token", () => {
    saveToken("abc123");
    clearToken();
    expect(getToken()).toBeNull();
  });
});

describe("ApiError", () => {
  it("captures status and message", () => {
    const error = new ApiError(401, "Unauthorized");
    expect(error.status).toBe(401);
    expect(error.message).toBe("Unauthorized");
  });
});

describe("discordUrl", () => {
  it("returns the Discord OAuth endpoint", () => {
    expect(discordUrl()).toBe(`${API_URL}/api/v1/auth/discord`);
  });
});

describe("api client", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", mockFetchJson({ ok: true }));
  });

  it("sends JSON requests with credentials", async () => {
    const fetchMock = mockFetchJson({ user_id: "1", message: "ok" });
    vi.stubGlobal("fetch", fetchMock);

    await api.register({
      email: "a@b.com",
      password: "secret",
      display_name: "Snack",
      country: "US",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      `${API_URL}/api/v1/auth/register`,
      expect.objectContaining({
        method: "POST",
        credentials: "include",
        body: JSON.stringify({
          email: "a@b.com",
          password: "secret",
          display_name: "Snack",
          country: "US",
        }),
      }),
    );
  });

  it("attaches bearer token when provided", async () => {
    const fetchMock = mockFetchJson({ id: "user-1" });
    vi.stubGlobal("fetch", fetchMock);

    await api.me("token-123");

    const [, init] = fetchMock.mock.calls[0];
    const headers = (init as RequestInit).headers as Headers;
    expect(headers.get("Authorization")).toBe("Bearer token-123");
  });

  it("throws ApiError on failed responses", async () => {
    vi.stubGlobal("fetch", mockFetchJson({ error: "Invalid credentials" }, 401));

    await expect(api.login({ email: "a@b.com", password: "bad" })).rejects.toEqual(
      expect.objectContaining<ApiError>({
        status: 401,
        message: "Invalid credentials",
      }),
    );
  });

  it("defaults wishlists to an empty array when null", async () => {
    vi.stubGlobal("fetch", mockFetchJson(null));

    await expect(api.wishlists("token")).resolves.toEqual([]);
  });

  it("defaults wishlist items to an empty array when missing", async () => {
    vi.stubGlobal(
      "fetch",
      mockFetchJson({
        wishlist: { id: "w1", title: "Snacks" },
      }),
    );

    await expect(api.getWishlist("w1", "token")).resolves.toEqual({
      wishlist: { id: "w1", title: "Snacks" },
      items: [],
      viewer_can_snag: false,
    });
  });

  it("defaults matches to an empty array when null", async () => {
    vi.stubGlobal("fetch", mockFetchJson(null));
    await expect(api.matches("token")).resolves.toEqual([]);
  });

  it("defaults notifications items to an empty array when missing", async () => {
    vi.stubGlobal("fetch", mockFetchJson({}));
    await expect(api.notifications("token")).resolves.toEqual({
      unread_count: 0,
      items: [],
    });
  });

  it("encodes search queries", async () => {
    const fetchMock = mockFetchJson([]);
    vi.stubGlobal("fetch", fetchMock);

    await api.search("hot & spicy");

    expect(fetchMock).toHaveBeenCalledWith(
      `${API_URL}/api/v1/search?q=hot%20%26%20spicy`,
      expect.any(Object),
    );
  });

  it("defaults profile wishlists to an empty array", async () => {
    vi.stubGlobal(
      "fetch",
      mockFetchJson({
        user: { id: "u1", display_name: "Snack" },
      }),
    );

    await expect(api.getUserProfile("u1", "token")).resolves.toEqual({
      user: { id: "u1", display_name: "Snack" },
      wishlists: [],
    });
  });

  it("uploads avatar as multipart form data without json content type", async () => {
    const fetchMock = mockFetchJson({ avatar_url: "https://cdn/avatar.jpg" });
    vi.stubGlobal("fetch", fetchMock);

    const file = new File(["avatar"], "avatar.png", { type: "image/png" });
    await api.uploadAvatar(file, "token");

    const [, init] = fetchMock.mock.calls[0];
    expect((init as RequestInit).body).toBeInstanceOf(FormData);
    expect((init as RequestInit).headers).not.toEqual(
      expect.objectContaining({ "Content-Type": "application/json" }),
    );
  });

  it("calls webauthnRegisterFinish with raw credential body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(mockJsonResponse({ ok: true }));
    vi.stubGlobal("fetch", fetchMock);

    const credential = new ArrayBuffer(8);
    await api.webauthnRegisterFinish(credential, "session-1", "token");

    expect(fetchMock).toHaveBeenCalledWith(
      `${API_URL}/api/v1/auth/mfa/webauthn/register/finish`,
      expect.objectContaining({
        method: "POST",
        body: credential,
        headers: expect.objectContaining({
          Authorization: "Bearer token",
          "X-WebAuthn-Session": "session-1",
        }),
      }),
    );
  });
});

function mockJsonResponse(data: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? "OK" : "Error",
    json: () => Promise.resolve(data),
  } as Response;
}
