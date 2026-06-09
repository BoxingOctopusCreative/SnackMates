import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { LoginForm } from "@/app/login/LoginForm";
import * as apiModule from "@/lib/api";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders, getField } from "@test/utils";

describe("LoginPage", () => {
  it("signs in and stores the auth token", async () => {
    const user = userEvent.setup();
    vi.spyOn(apiModule.api, "login").mockResolvedValue({ token: "jwt-token" });

    renderWithProviders(<LoginForm background={null} />);

    await user.type(getField(/^Email/), "snacker@example.com");
    await user.type(getField(/^Password/), "secret123");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(apiModule.getToken()).toBe("jwt-token");
      expect(navigationMocks.push).toHaveBeenCalledWith("/dashboard");
    });
  });

  it("shows MFA step when required", async () => {
    const user = userEvent.setup();
    vi.spyOn(apiModule.api, "login").mockResolvedValue({ mfa_required: true, methods: ["totp"] });

    renderWithProviders(<LoginForm background={null} />);

    await user.type(getField(/^Email/), "snacker@example.com");
    await user.type(getField(/^Password/), "secret123");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByRole("textbox", { name: /^Authenticator code/ })).toBeInTheDocument();
  });

  it("shows OAuth errors from the query string", async () => {
    window.history.replaceState({}, "", "/login?error=Discord%20failed");
    renderWithProviders(<LoginForm background={null} />);

    expect(await screen.findByText("Discord failed")).toBeInTheDocument();
    expect(window.location.search).toBe("");
  });

  it("shows API errors", async () => {
    const user = userEvent.setup();
    vi.spyOn(apiModule.api, "login").mockRejectedValue(new apiModule.ApiError(401, "Bad credentials"));

    renderWithProviders(<LoginForm background={null} />);

    await user.type(getField(/^Email/), "snacker@example.com");
    await user.type(getField(/^Password/), "wrong");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByText("Bad credentials")).toBeInTheDocument();
  });
});
