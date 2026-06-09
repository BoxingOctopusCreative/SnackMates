import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import DashboardPage from "@/app/(app)/dashboard/page";
import * as apiModule from "@/lib/api";
import { mockPublicUser, mockSnackMate, mockUser, mockWishlist } from "@test/fixtures";
import { renderWithProviders } from "@test/utils";

vi.mock("@/components/AuthGate", () => ({
  useAuth: () => ({
    user: { ...mockUser, email_verified: false },
    updateUser: vi.fn(),
    refreshUser: vi.fn(),
  }),
}));

describe("DashboardPage", () => {
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

  it("shows welcome stats and verification notice", async () => {
    vi.spyOn(apiModule.api, "wishlists").mockResolvedValue([mockWishlist]);
    vi.spyOn(apiModule.api, "friends").mockResolvedValue([mockSnackMate]);
    vi.spyOn(apiModule.api, "friendWishlists").mockResolvedValue([]);

    renderWithProviders(<DashboardPage />);

    expect(await screen.findByRole("heading", { name: /welcome back, snack fan/i })).toBeInTheDocument();
    expect(screen.getByText(/verify your email/i)).toBeInTheDocument();
    expect(await screen.findByText("Wishlists")).toBeInTheDocument();
    expect(screen.getByText("Snack Mates")).toBeInTheDocument();
  });

  it("shows snack mates' wishlists on the dashboard", async () => {
    vi.spyOn(apiModule.api, "wishlists").mockResolvedValue([mockWishlist]);
    vi.spyOn(apiModule.api, "friends").mockResolvedValue([mockSnackMate]);
    vi.spyOn(apiModule.api, "friendWishlists").mockResolvedValue([
      {
        ...mockWishlist,
        id: "wishlist-2",
        slug: "bruno-chocolate",
        title: "Bruno's Chocolate Box",
        user_id: "user-2",
        owner: {
          id: mockPublicUser.id,
          username: mockPublicUser.username,
          display_name: mockPublicUser.display_name,
        },
      },
    ]);

    renderWithProviders(<DashboardPage />);

    expect(await screen.findByRole("heading", { name: "Snack Mates' Wishlists" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Bruno's Chocolate Box" })).toHaveAttribute(
      "href",
      "/wishlists/bruno-chocolate",
    );
    expect(screen.getByRole("link", { name: "Snack Fan" })).toHaveAttribute("href", "/users/snackfan");
  });
});
