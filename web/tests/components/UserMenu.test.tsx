import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { UserMenu } from "@/components/UserMenu";
import { clearToken, getToken } from "@/lib/api";
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

describe("UserMenu", () => {
  it("opens profile settings in a modal", async () => {
    const user = userEvent.setup();
    renderWithProviders(<UserMenu user={mockUser} />);

    await user.click(screen.getByRole("button", { name: "Account menu" }));
    await user.click(screen.getByRole("menuitem", { name: "Profile Settings" }));

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Account Settings" })).toBeInTheDocument();
    expect(navigationMocks.push).not.toHaveBeenCalledWith("/settings");
  });

  it("navigates to the profile page", async () => {
    const user = userEvent.setup();
    renderWithProviders(<UserMenu user={mockUser} />);

    await user.click(screen.getByRole("button", { name: "Account menu" }));
    await user.click(screen.getByRole("menuitem", { name: "View Profile" }));
    expect(navigationMocks.push).toHaveBeenCalledWith("/users/snackfan");
  });

  it("logs out and clears the auth token", async () => {
    const user = userEvent.setup();
    localStorage.setItem("snackmates_token", "token-123");

    renderWithProviders(<UserMenu user={mockUser} />);

    await user.click(screen.getByRole("button", { name: "Account menu" }));
    await user.click(screen.getByRole("menuitem", { name: "Log Out" }));

    expect(getToken()).toBeNull();
    expect(navigationMocks.push).toHaveBeenCalledWith("/login");
    clearToken();
  });

  it("switches color scheme from the menu", async () => {
    const user = userEvent.setup();
    renderWithProviders(<UserMenu user={mockUser} />);

    await user.click(screen.getByRole("button", { name: "Account menu" }));
    await user.click(screen.getByRole("menuitem", { name: "Dark mode" }));

    expect(document.documentElement.dataset.theme).toBe("dark");
  });
});
