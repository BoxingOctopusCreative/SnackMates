import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import WishlistDetailPage from "@/app/(app)/wishlists/[slug]/page";
import * as apiModule from "@/lib/api";
import { mockUser, mockWishlist, mockWishlistItem } from "@test/fixtures";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders, getField } from "@test/utils";

vi.mock("@/components/AuthGate", () => ({
  useAuth: () => ({
    user: mockUser,
    updateUser: vi.fn(),
    refreshUser: vi.fn(),
  }),
}));

describe("WishlistDetailPage", () => {
  beforeEach(() => {
    vi.spyOn(globalThis, "fetch").mockImplementation((input) => {
      if (String(input).includes("/api/unsplash/random")) {
        return Promise.resolve({
          ok: true,
          json: async () => ({ photo: null }),
        } as Response);
      }
      return Promise.reject(new Error(`Unexpected fetch: ${String(input)}`));
    });
  });

  it("loads wishlist details and items", async () => {
    navigationMocks.params = { slug: "sweet-treats" };
    vi.spyOn(apiModule.api, "getWishlist").mockResolvedValue({
      wishlist: mockWishlist,
      items: [mockWishlistItem],
      viewer_can_snag: false,
    });

    renderWithProviders(<WishlistDetailPage />);

    expect(await screen.findByRole("heading", { name: "Sweet treats" })).toBeInTheDocument();
    expect(screen.getByText("Pocky")).toBeInTheDocument();
    expect(screen.getByRole("img", { name: "Pocky" })).toHaveAttribute(
      "src",
      mockWishlistItem.image_url,
    );
  });

  it("shows all table columns when the wishlist is empty", async () => {
    navigationMocks.params = { slug: "sweet-treats" };
    vi.spyOn(apiModule.api, "getWishlist").mockResolvedValue({
      wishlist: mockWishlist,
      items: [],
      viewer_can_snag: false,
    });

    renderWithProviders(<WishlistDetailPage />);

    await screen.findByRole("heading", { name: "Sweet treats" });

    for (const heading of ["Image", "Snack Name", "Type", "Brand", "Notes", "Status", "Actions"]) {
      expect(screen.getByRole("columnheader", { name: heading })).toBeVisible();
    }

    expect(screen.getByText("No snacks on this wishlist yet.")).toBeInTheDocument();
  });

  it("sorts items when a column header is clicked", async () => {
    const user = userEvent.setup();
    navigationMocks.params = { slug: "sweet-treats" };
    vi.spyOn(apiModule.api, "getWishlist").mockResolvedValue({
      wishlist: mockWishlist,
      items: [
        {
          ...mockWishlistItem,
          id: "item-1",
          name: "Snack B",
          brand: "Brand A",
        },
        {
          ...mockWishlistItem,
          id: "item-2",
          name: "Snack A",
          brand: "Brand Z",
        },
      ],
      viewer_can_snag: false,
    });

    renderWithProviders(<WishlistDetailPage />);
    await screen.findByRole("heading", { name: "Sweet treats" });

    expect(screen.getAllByRole("row").slice(1).map((row) => row.textContent)).toEqual([
      expect.stringContaining("Snack A"),
      expect.stringContaining("Snack B"),
    ]);

    await user.click(screen.getByRole("button", { name: /^Brand/ }));

    expect(screen.getAllByRole("row").slice(1).map((row) => row.textContent)).toEqual([
      expect.stringContaining("Snack B"),
      expect.stringContaining("Snack A"),
    ]);

    await user.click(screen.getByRole("button", { name: /^Brand/ }));

    expect(screen.getAllByRole("row").slice(1).map((row) => row.textContent)).toEqual([
      expect.stringContaining("Snack A"),
      expect.stringContaining("Snack B"),
    ]);
  });

  it("adds and removes items", async () => {
    const user = userEvent.setup();
    navigationMocks.params = { slug: "sweet-treats" };
    const getWishlistSpy = vi
      .spyOn(apiModule.api, "getWishlist")
      .mockResolvedValueOnce({
        wishlist: mockWishlist,
        items: [],
        viewer_can_snag: false,
      })
      .mockResolvedValueOnce({
        wishlist: mockWishlist,
        items: [mockWishlistItem],
        viewer_can_snag: false,
      })
      .mockResolvedValueOnce({
        wishlist: mockWishlist,
        items: [],
        viewer_can_snag: false,
      });
    vi.spyOn(apiModule.api, "addItem").mockResolvedValue(mockWishlistItem);
    vi.spyOn(apiModule.api, "deleteItem").mockResolvedValue({ ok: true });

    renderWithProviders(<WishlistDetailPage />);
    await screen.findByRole("heading", { name: "Sweet treats" });

    await user.click(screen.getByRole("button", { name: "Add Snack" }));
    await user.type(getField(/^Snack Name/), "Pocky");
    await user.click(screen.getByRole("button", { name: "Add to wishlist" }));

    await waitFor(() => {
      expect(apiModule.api.addItem).toHaveBeenCalledWith(
        "sweet-treats",
        expect.objectContaining({
          name: "Pocky",
          type: "Candy",
        }),
        null,
      );
    });

    await user.click(screen.getByRole("button", { name: "Remove" }));

    await waitFor(() => {
      expect(apiModule.api.deleteItem).toHaveBeenCalledWith("sweet-treats", "item-1", null);
      expect(getWishlistSpy).toHaveBeenCalledTimes(3);
    });
  });

  it("shows snagged status and lets matched mates mark items", async () => {
    const user = userEvent.setup();
    navigationMocks.params = { slug: "mate-treats" };
    const mateWishlist = { ...mockWishlist, id: "wishlist-2", user_id: "user-2", slug: "mate-treats" };
    const snagSpy = vi.spyOn(apiModule.api, "snagItem").mockResolvedValue({
      ...mockWishlistItem,
      snagged_by: {
        id: "user-1",
        display_name: "Snack Fan",
        delivery_method: "in_person",
      },
    });
    const getWishlistSpy = vi
      .spyOn(apiModule.api, "getWishlist")
      .mockResolvedValueOnce({
        wishlist: mateWishlist,
        items: [
          mockWishlistItem,
          {
            ...mockWishlistItem,
            id: "item-2",
            name: "KitKat",
            snagged_by: { id: "user-1", display_name: "Snack Fan", delivery_method: "in_person" },
          },
        ],
        viewer_can_snag: true,
      })
      .mockResolvedValueOnce({
        wishlist: mateWishlist,
        items: [
          {
            ...mockWishlistItem,
            snagged_by: { id: "user-1", display_name: "Snack Fan", delivery_method: "in_person" },
          },
          {
            ...mockWishlistItem,
            id: "item-2",
            name: "KitKat",
            snagged_by: { id: "user-1", display_name: "Snack Fan", delivery_method: "in_person" },
          },
        ],
        viewer_can_snag: true,
      });

    renderWithProviders(<WishlistDetailPage />);
    await screen.findByRole("heading", { name: "Sweet treats" });

    expect(screen.getByText("Snagged by Snack Fan (delivery in-person)")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Mark snagged" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Remove" })).not.toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: "Add Snack" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Mark snagged" }));
    await screen.findByRole("heading", { name: "Mark Snagged" });

    await user.click(screen.getByRole("radio", { name: "By mail" }));
    await user.type(screen.getByLabelText("Tracking number"), "1Z999AA10123456784");
    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Mark snagged" }));

    await waitFor(() => {
      expect(snagSpy).toHaveBeenCalledWith(
        "mate-treats",
        "item-1",
        { delivery_method: "mail", tracking_number: "1Z999AA10123456784" },
        null,
      );
      expect(getWishlistSpy).toHaveBeenCalledTimes(2);
    });
  });
});
