import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  clearRandomUnsplashPhotoCache,
  fetchRandomUnsplashPhoto,
  searchUnsplashPhotos,
  trackUnsplashDownload,
} from "@/lib/unsplash";
import { mockJsonResponse } from "@test/utils";

const unsplashPhoto = {
  urls: { regular: "https://images.unsplash.com/photo-1" },
  links: {
    html: "https://unsplash.com/photos/1",
    download_location: "https://api.unsplash.com/photos/1/download",
  },
  user: {
    name: "Alex",
    links: { html: "https://unsplash.com/@alex" },
  },
};

describe("searchUnsplashPhotos", () => {
  beforeEach(() => {
    vi.stubEnv("UNSPLASH_ACCESS_KEY", "test-key");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("returns an empty array without an access key", async () => {
    vi.unstubAllEnvs();
    await expect(searchUnsplashPhotos("snacks")).resolves.toEqual([]);
  });

  it("returns an empty array for blank queries", async () => {
    await expect(searchUnsplashPhotos("   ")).resolves.toEqual([]);
  });

  it("maps Unsplash search results", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        mockJsonResponse({
          results: [unsplashPhoto],
        }),
      ),
    );

    await expect(searchUnsplashPhotos("snacks", 6)).resolves.toEqual([
      {
        url: unsplashPhoto.urls.regular,
        photographer: "Alex",
        photographerUrl: "https://unsplash.com/@alex",
        unsplashUrl: "https://unsplash.com/photos/1",
      },
    ]);

    expect(fetch).toHaveBeenCalledWith(
      expect.stringContaining("https://api.unsplash.com/search/photos?"),
      expect.objectContaining({
        headers: { Authorization: "Client-ID test-key" },
      }),
    );
  });

  it("returns an empty array when the API fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(mockJsonResponse({}, 500)),
    );

    await expect(searchUnsplashPhotos("snacks")).resolves.toEqual([]);
  });
});

describe("trackUnsplashDownload", () => {
  beforeEach(() => {
    vi.stubEnv("UNSPLASH_ACCESS_KEY", "test-key");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("tracks downloads when configured", async () => {
    const fetchMock = vi.fn().mockResolvedValue(mockJsonResponse({}));
    vi.stubGlobal("fetch", fetchMock);

    await trackUnsplashDownload("https://api.unsplash.com/photos/1/download");

    expect(fetchMock).toHaveBeenCalledWith(
      "https://api.unsplash.com/photos/1/download?client_id=test-key",
      { cache: "no-store" },
    );
  });

  it("no-ops without an access key", async () => {
    vi.unstubAllEnvs();
    await trackUnsplashDownload("https://api.unsplash.com/photos/1/download");
    expect(fetch).not.toHaveBeenCalled();
  });
});

describe("fetchRandomUnsplashPhoto", () => {
  beforeEach(() => {
    clearRandomUnsplashPhotoCache();
    vi.stubEnv("UNSPLASH_ACCESS_KEY", "test-key");
  });

  afterEach(() => {
    clearRandomUnsplashPhotoCache();
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("returns null without an access key", async () => {
    vi.unstubAllEnvs();
    await expect(fetchRandomUnsplashPhoto("snacks")).resolves.toBeNull();
  });

  it("returns null when no photos are found", async () => {
    const fetchMock = vi.fn().mockResolvedValue(mockJsonResponse({ errors: ["Not Found"] }, 404));
    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchRandomUnsplashPhoto("snacks")).resolves.toBeNull();
    expect(fetchMock).toHaveBeenCalledTimes(4);
  });

  it("returns a mapped photo and tracks the download", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(mockJsonResponse(unsplashPhoto))
      .mockResolvedValueOnce(mockJsonResponse({}));

    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchRandomUnsplashPhoto("snacks")).resolves.toEqual({
      url: unsplashPhoto.urls.regular,
      photographer: "Alex",
      photographerUrl: "https://unsplash.com/@alex",
      unsplashUrl: "https://unsplash.com/photos/1",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("https://api.unsplash.com/photos/random?"),
      expect.objectContaining({
        headers: { Authorization: "Client-ID test-key" },
      }),
    );
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("falls back to broader queries when the primary query misses", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(mockJsonResponse({ errors: ["Not Found"] }, 404))
      .mockResolvedValueOnce(mockJsonResponse(unsplashPhoto))
      .mockResolvedValueOnce(mockJsonResponse({}));

    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchRandomUnsplashPhoto("cartoon snack food")).resolves.toEqual({
      url: unsplashPhoto.urls.regular,
      photographer: "Alex",
      photographerUrl: "https://unsplash.com/@alex",
      unsplashUrl: "https://unsplash.com/photos/1",
    });

    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it("reuses cached photos for repeated requests", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(mockJsonResponse(unsplashPhoto))
      .mockResolvedValueOnce(mockJsonResponse({}));

    vi.stubGlobal("fetch", fetchMock);

    await expect(fetchRandomUnsplashPhoto("snacks")).resolves.not.toBeNull();
    await expect(fetchRandomUnsplashPhoto("snacks")).resolves.not.toBeNull();

    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});
