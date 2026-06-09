import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import SearchPage from "@/app/(app)/search/page";
import * as apiModule from "@/lib/api";
import { mockSearchResponse } from "@test/fixtures";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders } from "@test/utils";

describe("SearchPage", () => {
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

  it("loads universal results from the URL query", async () => {
    navigationMocks.searchParams = new URLSearchParams("q=coffee+crisp");
    vi.spyOn(apiModule.api, "search").mockResolvedValue(mockSearchResponse);

    renderWithProviders(<SearchPage />);

    expect(await screen.findByText("Snack Fan")).toBeInTheDocument();
    expect(screen.getByText("Pocky")).toBeInTheDocument();
    expect(screen.getByText("Coffee Crisp")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add to wishlist" })).toBeInTheDocument();
  });

  it("adds a product result to a wishlist", async () => {
    const user = userEvent.setup();
    navigationMocks.searchParams = new URLSearchParams("q=coffee+crisp&type=products");
    vi.spyOn(apiModule.api, "search").mockResolvedValue(mockSearchResponse);
    vi.spyOn(apiModule.api, "wishlists").mockResolvedValue([
      {
        id: "wishlist-1",
        user_id: "user-1",
        slug: "sweet-treats",
        title: "Sweet treats",
        description: "",
        is_public: true,
        item_count: 0,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      },
    ]);
    vi.spyOn(apiModule.api, "addItem").mockResolvedValue({
      id: "item-1",
      wishlist_id: "wishlist-1",
      name: "Coffee Crisp",
      type: "Candy",
      brand: "Nestlé",
      notes: "",
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
    });

    renderWithProviders(<SearchPage />);
    await screen.findByText("Coffee Crisp");

    await user.click(screen.getByRole("button", { name: "Add to wishlist" }));
    await user.click(await screen.findByRole("menuitem", { name: "Sweet treats" }));

    await waitFor(() => {
      expect(apiModule.api.addItem).toHaveBeenCalledWith(
        "sweet-treats",
        {
          name: "Coffee Crisp",
          type: "Candy",
          brand: "Nestlé",
          notes: "Barcode: 0059800000215 · Chocolate stuffed wafers · 50 g",
          image_url: "https://example.com/coffee-crisp.jpg",
        },
        null,
      );
    });
    expect(await screen.findByText("Added to Sweet treats")).toBeInTheDocument();
  });

  it("shows a no-results message", async () => {
    navigationMocks.searchParams = new URLSearchParams("q=unknown");
    vi.spyOn(apiModule.api, "search").mockResolvedValue({
      query: "unknown",
      search_terms: "unknown",
      ai_assisted: false,
      users: [],
      wishlist_items: [],
      products: [],
    });

    renderWithProviders(<SearchPage />);

    expect(await screen.findByText("No results found.")).toBeInTheDocument();
  });

  it("updates the URL when submitting a new search", async () => {
    const user = userEvent.setup();
    navigationMocks.searchParams = new URLSearchParams();
    vi.spyOn(apiModule.api, "search").mockResolvedValue({
      query: "",
      search_terms: "",
      ai_assisted: false,
      users: [],
      wishlist_items: [],
      products: [],
    });

    renderWithProviders(<SearchPage />);

    await user.type(screen.getByRole("searchbox", { name: "Search" }), "chips");
    await user.click(screen.getByRole("button", { name: "Search" }));

    expect(navigationMocks.replace).toHaveBeenCalledWith("/search?q=chips");
  });

  it("shows AI-assisted product search terms when they differ from the query", async () => {
    navigationMocks.searchParams = new URLSearchParams("q=coffee+crisp");
    vi.spyOn(apiModule.api, "search").mockResolvedValue(mockSearchResponse);

    renderWithProviders(<SearchPage />);

    expect(
      await screen.findByText(
        "Product matches used AI-assisted OpenFoodFacts search for “nestle coffee crisp chocolate bar”.",
      ),
    ).toBeInTheDocument();
  });

  it("filters results with tabs", async () => {
    const user = userEvent.setup();
    navigationMocks.searchParams = new URLSearchParams("q=pocky");
    vi.spyOn(apiModule.api, "search").mockResolvedValue(mockSearchResponse);

    renderWithProviders(<SearchPage />);
    await screen.findByText("Pocky");

    await user.click(screen.getByRole("button", { name: "People" }));

    expect(navigationMocks.replace).toHaveBeenCalledWith("/search?q=pocky&type=people");
  });
});
