import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { WishlistHero } from "@/components/WishlistHero";
import * as apiModule from "@/lib/api";
import { renderWithProviders } from "@test/utils";

describe("WishlistHero", () => {
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

  it("renders the title and description in the hero", async () => {
    renderWithProviders(
      <WishlistHero
        wishlistSlug="sweet-treats"
        title="Sweet treats"
        description="Candy and cookies"
        isPublic
        onBannerChange={vi.fn()}
        onDetailsChange={vi.fn()}
      />,
    );

    expect(await screen.findByRole("heading", { name: "Sweet treats" })).toBeInTheDocument();
    expect(screen.getByText("Candy and cookies")).toBeInTheDocument();
  });

  it("uses a custom banner when provided", async () => {
    renderWithProviders(
      <WishlistHero
        wishlistSlug="sweet-treats"
        title="Sweet treats"
        isPublic
        bannerUrl="https://example.com/banner.jpg"
        onBannerChange={vi.fn()}
        onDetailsChange={vi.fn()}
      />,
    );

    await waitFor(() => {
      expect(screen.getByRole("img", { name: "Sweet treats banner" })).toHaveStyle({
        backgroundImage: "url(https://example.com/banner.jpg)",
      });
    });
  });

  it("shows an add snack button for owners when onAddSnack is provided", async () => {
    const onAddSnack = vi.fn();

    renderWithProviders(
      <WishlistHero
        wishlistSlug="sweet-treats"
        title="Sweet treats"
        description="Candy and cookies"
        isPublic
        isOwner
        onAddSnack={onAddSnack}
        onBannerChange={vi.fn()}
        onDetailsChange={vi.fn()}
      />,
    );

    await screen.findByRole("heading", { name: "Sweet treats" });

    await userEvent.click(screen.getByRole("button", { name: "Add Snack" }));

    expect(onAddSnack).toHaveBeenCalledTimes(1);
  });

  it("shows pencil edit buttons for owners", async () => {
    renderWithProviders(
      <WishlistHero
        wishlistSlug="sweet-treats"
        title="Sweet treats"
        description="Candy and cookies"
        isPublic
        isOwner
        onBannerChange={vi.fn()}
        onDetailsChange={vi.fn()}
      />,
    );

    await screen.findByRole("heading", { name: "Sweet treats" });

    expect(screen.getByRole("button", { name: "Edit banner" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Edit title" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Edit description" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Customize banner" })).not.toBeInTheDocument();
  });

  it("edits the title inline", async () => {
    const user = userEvent.setup();
    const onDetailsChange = vi.fn();
    vi.spyOn(apiModule.api, "updateWishlist").mockResolvedValue({ ok: true, slug: "holiday-snacks" });

    renderWithProviders(
      <WishlistHero
        wishlistSlug="sweet-treats"
        title="Sweet treats"
        description="Candy and cookies"
        isPublic
        isOwner
        onBannerChange={vi.fn()}
        onDetailsChange={onDetailsChange}
      />,
    );

    await user.click(await screen.findByRole("button", { name: "Edit title" }));
    const field = screen.getByLabelText(/^Title/);
    await user.clear(field);
    await user.type(field, "Holiday snacks");
    await user.keyboard("{Enter}");

    await waitFor(() => {
      expect(apiModule.api.updateWishlist).toHaveBeenCalledWith(
        "sweet-treats",
        {
          title: "Holiday snacks",
          description: "Candy and cookies",
          is_public: true,
        },
        null,
      );
      expect(onDetailsChange).toHaveBeenCalledWith({
        title: "Holiday snacks",
        description: "Candy and cookies",
        slug: "holiday-snacks",
      });
    });
  });
});
