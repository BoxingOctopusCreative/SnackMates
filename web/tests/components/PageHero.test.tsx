import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { PageHero } from "@/components/PageHero";
import { renderWithProviders } from "@test/utils";

describe("PageHero", () => {
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

  it("renders the title with the hero banner", async () => {
    renderWithProviders(<PageHero title="Your Wishlists" />);

    expect(await screen.findByRole("heading", { name: "Your Wishlists" })).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "Your Wishlists banner" })).toHaveStyle({
      backgroundImage: "url(https://images.unsplash.com/photo-1)",
    });
  });

  it("renders an optional description", async () => {
    renderWithProviders(
      <PageHero title="Snack Mates" description="Find your next snack pen pal." />,
    );

    expect(await screen.findByRole("heading", { name: "Snack Mates" })).toBeInTheDocument();
    expect(screen.getByText("Find your next snack pen pal.")).toBeInTheDocument();
  });
});
