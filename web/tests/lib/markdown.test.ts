import { describe, expect, it } from "vitest";
import { fetchLegalMarkdown } from "@/lib/markdown";
import { getStaticAssetsBaseUrl, staticAssetUrl } from "@/lib/static-assets";

describe("staticAssetUrl", () => {
  it("defaults to the production static assets CDN", () => {
    expect(getStaticAssetsBaseUrl()).toBe("https://assets.snackmates.food");
    expect(staticAssetUrl("legal/terms-of-use.md")).toBe(
      "https://assets.snackmates.food/legal/terms-of-use.md",
    );
  });

  it("uses STATIC_ASSETS_BASE_URL when configured", () => {
    const previous = process.env.STATIC_ASSETS_BASE_URL;
    process.env.STATIC_ASSETS_BASE_URL = "http://localhost:9000/static-assets/";
    expect(staticAssetUrl("/legal/privacy-policy.md")).toBe(
      "http://localhost:9000/static-assets/legal/privacy-policy.md",
    );
    if (previous === undefined) {
      delete process.env.STATIC_ASSETS_BASE_URL;
    } else {
      process.env.STATIC_ASSETS_BASE_URL = previous;
    }
  });
});

describe("fetchLegalMarkdown", () => {
  it("reads markdown from content/legal", async () => {
    const content = await fetchLegalMarkdown("terms-of-use.md");
    expect(content).toContain("SnackMates");
  });

  it("rejects path traversal filenames", async () => {
    await expect(fetchLegalMarkdown("../terms-of-use.md")).rejects.toThrow(
      "Invalid legal markdown filename",
    );
  });

  it("throws when the markdown file does not exist", async () => {
    await expect(fetchLegalMarkdown("missing.md")).rejects.toThrow(
      "Failed to load legal markdown",
    );
  });
});
