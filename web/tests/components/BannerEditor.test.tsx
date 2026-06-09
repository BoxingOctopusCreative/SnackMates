import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { BannerEditor } from "@/components/BannerEditor";
import * as apiModule from "@/lib/api";
import { renderWithProviders, getField } from "@test/utils";

describe("BannerEditor", () => {
  it("searches Unsplash and selects a photo", async () => {
    const user = userEvent.setup();
    const onBannerChange = vi.fn();
    const onMessage = vi.fn();

    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: () =>
          Promise.resolve({
            results: [
              {
                url: "https://images.unsplash.com/photo-1",
                photographer: "Alex",
                photographerUrl: "https://unsplash.com/@alex",
                unsplashUrl: "https://unsplash.com/photos/1",
              },
            ],
          }),
      }),
    );
    vi.spyOn(apiModule.api, "setBannerUrl").mockResolvedValue({
      banner_url: "https://images.unsplash.com/photo-1",
    });

    renderWithProviders(
      <BannerEditor bannerUrl="" onBannerChange={onBannerChange} onMessage={onMessage} />,
    );

    await user.type(getField(/^Search Unsplash/), "snacks");
    await user.click(screen.getByRole("button", { name: "Search" }));

    await screen.findByRole("button", { name: /use photo by alex/i });
    await user.click(screen.getByRole("button", { name: /use photo by alex/i }));

    await waitFor(() => {
      expect(onBannerChange).toHaveBeenCalledWith("https://images.unsplash.com/photo-1");
      expect(onMessage).toHaveBeenCalledWith("Profile banner updated from Unsplash.");
    });
  });

  it("uploads a banner file", async () => {
    const user = userEvent.setup();
    const onBannerChange = vi.fn();
    const onMessage = vi.fn();

    vi.spyOn(apiModule.api, "uploadBanner").mockResolvedValue({
      banner_url: "https://cdn/banner.jpg",
    });

    renderWithProviders(
      <BannerEditor bannerUrl="" onBannerChange={onBannerChange} onMessage={onMessage} />,
    );

    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    const file = new File(["banner"], "banner.png", { type: "image/png" });
    await user.upload(input, file);

    await waitFor(() => {
      expect(onBannerChange).toHaveBeenCalledWith("https://cdn/banner.jpg");
      expect(onMessage).toHaveBeenCalledWith("Profile banner updated.");
    });
  });
});
