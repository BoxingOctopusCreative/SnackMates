import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import WishlistsPage from "@/app/(app)/wishlists/page";
import * as apiModule from "@/lib/api";
import { mockWishlist } from "@test/fixtures";
import { renderWithProviders, getField } from "@test/utils";

describe("WishlistsPage", () => {
  beforeEach(() => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({
        photo: {
          url: "https://images.unsplash.com/photo-1",
          photographer: "Alex",
          photographerUrl: "https://unsplash.com/@alex",
          unsplashUrl: "https://unsplash.com/photos/1",
        },
      }),
    } as Response);
  });

  it("loads and displays wishlists", async () => {
    vi.spyOn(apiModule.api, "wishlists").mockResolvedValue([mockWishlist]);

    renderWithProviders(<WishlistsPage />);

    expect(await screen.findByRole("link", { name: "Sweet treats" })).toHaveAttribute(
      "href",
      "/wishlists/sweet-treats",
    );
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("creates a new wishlist", async () => {
    const user = userEvent.setup();
    const wishlistsSpy = vi
      .spyOn(apiModule.api, "wishlists")
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([mockWishlist]);
    vi.spyOn(apiModule.api, "createWishlist").mockResolvedValue(mockWishlist);

    renderWithProviders(<WishlistsPage />);

    await user.click(screen.getByRole("button", { name: "Create Wishlist" }));
    await user.type(getField(/^Title/), "Sweet treats");
    await user.type(getField(/^Description/), "Candy and cookies");
    await user.click(screen.getByRole("button", { name: "Create wishlist" }));

    await waitFor(() => {
      expect(apiModule.api.createWishlist).toHaveBeenCalledWith(
        {
          title: "Sweet treats",
          description: "Candy and cookies",
          is_public: true,
        },
        null,
      );
      expect(wishlistsSpy).toHaveBeenCalledTimes(2);
    });
  });
});
