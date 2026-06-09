import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { AuthProvider } from "@/components/AuthGate";
import * as apiModule from "@/lib/api";
import { mockUser } from "@test/fixtures";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders } from "@test/utils";

vi.mock("@/components/AppShell", () => ({
  AppShell: ({
    children,
    user,
  }: {
    children: React.ReactNode;
    user: { display_name: string };
  }) => (
    <div data-testid="app-shell" data-user={user.display_name}>
      {children}
    </div>
  ),
}));

describe("AuthProvider", () => {
  it("shows a loading state while fetching the current user", () => {
    vi.spyOn(apiModule.api, "me").mockReturnValue(new Promise(() => {}));

    renderWithProviders(
      <AuthProvider>
        <div>Protected content</div>
      </AuthProvider>,
    );

    expect(screen.getByLabelText("Loading")).toBeInTheDocument();
  });

  it("renders children inside the app shell when authenticated", async () => {
    vi.spyOn(apiModule.api, "me").mockResolvedValue(mockUser);

    renderWithProviders(
      <AuthProvider>
        <div>Protected content</div>
      </AuthProvider>,
    );

    expect(await screen.findByTestId("app-shell")).toHaveAttribute("data-user", "Snack Fan");
    expect(screen.getByText("Protected content")).toBeInTheDocument();
  });

  it("redirects to login when auth fails", async () => {
    vi.spyOn(apiModule.api, "me").mockRejectedValue(new Error("Unauthorized"));

    renderWithProviders(
      <AuthProvider>
        <div>Protected content</div>
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(navigationMocks.replace).toHaveBeenCalledWith("/login");
    });
    expect(screen.queryByText("Protected content")).not.toBeInTheDocument();
  });

  it("consumes OAuth tokens from the URL", async () => {
    window.history.replaceState({}, "", "/dashboard?token=oauth-token");
    const meSpy = vi.spyOn(apiModule.api, "me").mockResolvedValue(mockUser);

    renderWithProviders(
      <AuthProvider>
        <div>Protected content</div>
      </AuthProvider>,
    );

    await screen.findByText("Protected content");
    expect(apiModule.getToken()).toBe("oauth-token");
    expect(window.location.search).not.toContain("token=");
    meSpy.mockRestore();
  });
});

describe("useAuth", () => {
  it("throws outside AuthProvider", async () => {
    const { useAuth } = await import("@/components/AuthGate");

    function BrokenConsumer() {
      useAuth();
      return null;
    }

    expect(() => render(<BrokenConsumer />)).toThrow(
      "useAuth must be used within AuthProvider",
    );
  });
});
