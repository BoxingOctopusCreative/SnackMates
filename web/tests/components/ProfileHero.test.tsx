import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ProfileHero } from "@/components/ProfileHero";
import { mockPublicUser, mockUser } from "@test/fixtures";
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

describe("ProfileHero", () => {
  it("renders profile details and country name", () => {
    renderWithProviders(<ProfileHero user={mockPublicUser} />);

    expect(screen.getByRole("heading", { name: "Snack Fan" })).toBeInTheDocument();
    expect(screen.getByText("United States")).toBeInTheDocument();
    expect(screen.getByText("Loves chips")).toBeInTheDocument();
  });

  it("shows a settings button on the user's own profile", () => {
    renderWithProviders(<ProfileHero user={mockPublicUser} isOwnProfile />);
    expect(screen.getByRole("button", { name: "Profile Settings" })).toBeInTheDocument();
  });

  it("opens a settings modal when the settings button is clicked", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProfileHero user={mockPublicUser} isOwnProfile />);

    await user.click(screen.getByRole("button", { name: "Profile Settings" }));

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Account Settings" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save profile" })).toBeInTheDocument();
  });

  it("omits the settings button for other profiles", () => {
    renderWithProviders(<ProfileHero user={mockPublicUser} />);
    expect(screen.queryByRole("button", { name: "Profile Settings" })).not.toBeInTheDocument();
  });

  it("opens the settings modal when settings=open is in the URL", async () => {
    navigationMocks.searchParams = new URLSearchParams("settings=open");
    renderWithProviders(<ProfileHero user={mockPublicUser} isOwnProfile />);

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(navigationMocks.replace).toHaveBeenCalledWith("/users/snackfan", { scroll: false });
  });
});
