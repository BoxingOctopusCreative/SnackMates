import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import UserProfilePage from "@/app/(app)/users/[username]/page";
import * as apiModule from "@/lib/api";
import { mockPublicUser, mockUser, mockWishlist } from "@test/fixtures";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders } from "@test/utils";

vi.mock("@/components/AuthGate", () => ({
  useAuth: () => ({
    user: mockUser,
    updateUser: vi.fn(),
    refreshUser: vi.fn(),
  }),
}));

describe("UserProfilePage", () => {
  it("renders a public profile and wishlists", async () => {
    navigationMocks.params = { username: "matesnacker" };
    vi.spyOn(apiModule.api, "getUserProfile").mockResolvedValue({
      user: {
        ...mockPublicUser,
        id: "user-2",
        username: "matesnacker",
        display_name: "Mate Snacker",
        country: "JP",
      },
      wishlists: [mockWishlist],
    });

    renderWithProviders(<UserProfilePage />);

    expect(await screen.findByRole("heading", { name: "Mate Snacker" })).toBeInTheDocument();
    expect(screen.getByText("Japan")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Sweet treats" })).toBeInTheDocument();
  });

  it("shows an error state when the profile cannot be loaded", async () => {
    navigationMocks.params = { username: "missing" };
    vi.spyOn(apiModule.api, "getUserProfile").mockRejectedValue(
      new apiModule.ApiError(404, "User not found"),
    );

    renderWithProviders(<UserProfilePage />);

    expect(await screen.findByRole("heading", { name: "Profile Not Found" })).toBeInTheDocument();
    expect(screen.getByText("User not found")).toBeInTheDocument();
  });
});
