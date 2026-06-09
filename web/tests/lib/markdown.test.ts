import { afterEach, describe, expect, it, vi } from "vitest";
import { fetchLegalMarkdown } from "@/lib/markdown";
import { getStaticAssetsBaseUrl, staticAssetUrl } from "@/lib/static-assets";

describe("staticAssetUrl", () => {
  afterEach(() => {
    delete process.env.STATIC_ASSETS_BASE_URL;
  });

  it("defaults to the production static assets CDN", () => {
    expect(getStaticAssetsBaseUrl()).toBe("https://assets.snackmates.food");
    expect(staticAssetUrl("legal/terms-of-use.md")).toBe(
      "https://assets.snackmates.food/legal/terms-of-use.md",
    );
  });

  it("uses STATIC_ASSETS_BASE_URL when configured", () => {
    process.env.STATIC_ASSETS_BASE_URL = "http://localhost:9000/static-assets/";
    expect(staticAssetUrl("/legal/privacy-policy.md")).toBe(
      "http://localhost:9000/static-assets/legal/privacy-policy.md",
    );
  });
});

describe("fetchLegalMarkdown", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("fetches markdown from the legal directory on static assets", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue(
      new Response("# Terms\n\nBody copy", {
        status: 200,
        headers: { "Content-Type": "text/markdown; charset=utf-8" },
      }),
    );

    await expect(fetchLegalMarkdown("terms-of-use.md")).resolves.toBe("# Terms\n\nBody copy");
    expect(global.fetch).toHaveBeenCalledWith(
      "https://assets.snackmates.food/legal/terms-of-use.md",
      { next: { revalidate: 3600 } },
    );
  });

  it("throws when the markdown file cannot be loaded", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue(new Response("Not found", { status: 404 }));

    await expect(fetchLegalMarkdown("missing.md")).rejects.toThrow(
      "Failed to load legal markdown from https://assets.snackmates.food/legal/missing.md: 404",
    );
  });
});
