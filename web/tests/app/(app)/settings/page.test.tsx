import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import SettingsPage from "@/app/(app)/settings/page";
import { mockUser } from "@test/fixtures";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders } from "@test/utils";

vi.mock("@/components/AuthGate", () => ({
  useAuth: () => ({
    user: mockUser,
    updateUser: vi.fn(),
    refreshUser: vi.fn(),
  }),
}));

vi.mock("@/components/BannerEditor", () => ({
  BannerEditor: () => <div data-testid="banner-editor" />,
}));

describe("SettingsPage", () => {
  it("redirects to the profile page and opens settings in the modal", async () => {
    renderWithProviders(<SettingsPage />);

    await waitFor(() => {
      expect(navigationMocks.replace).toHaveBeenCalledWith("/users/snackfan");
    });
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Account Settings" })).toBeInTheDocument();
  });
});
