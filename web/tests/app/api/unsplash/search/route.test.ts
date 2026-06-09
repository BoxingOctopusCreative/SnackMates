import { NextRequest } from "next/server";
import { describe, expect, it, vi } from "vitest";
import { GET } from "@/app/api/unsplash/search/route";

vi.mock("@/lib/unsplash", () => ({
  searchUnsplashPhotos: vi.fn(),
}));

import { searchUnsplashPhotos } from "@/lib/unsplash";

describe("GET /api/unsplash/search", () => {
  it("returns an empty result set for blank queries", async () => {
    const response = await GET(new NextRequest("http://localhost/api/unsplash/search?q="));
    const body = await response.json();

    expect(body).toEqual({ results: [] });
    expect(searchUnsplashPhotos).not.toHaveBeenCalled();
  });

  it("proxies trimmed search queries", async () => {
    vi.mocked(searchUnsplashPhotos).mockResolvedValue([
      {
        url: "https://images.unsplash.com/photo-1",
        photographer: "Alex",
        photographerUrl: "https://unsplash.com/@alex",
        unsplashUrl: "https://unsplash.com/photos/1",
      },
    ]);

    const response = await GET(
      new NextRequest("http://localhost/api/unsplash/search?q=  snacks  "),
    );
    const body = await response.json();

    expect(searchUnsplashPhotos).toHaveBeenCalledWith("snacks", 12);
    expect(body.results).toHaveLength(1);
  });
});
