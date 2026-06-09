import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { RegisterForm } from "@/app/register/RegisterForm";
import * as apiModule from "@/lib/api";
import { renderWithProviders, getField } from "@test/utils";

describe("RegisterPage", () => {
  it("creates an account and shows the confirmation message", async () => {
    const user = userEvent.setup();
    vi.spyOn(apiModule.api, "register").mockResolvedValue({
      user_id: "user-1",
      message: "Check your email to verify your account.",
    });

    renderWithProviders(<RegisterForm background={null} />);

    await user.type(getField(/^Display name/), "Snack Fan");
    await user.type(getField(/^Email/), "snacker@example.com");
    await user.type(getField(/^Password/), "secret123");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    expect(await screen.findByText("Check your email to verify your account.")).toBeInTheDocument();
  });

  it("shows registration errors", async () => {
    const user = userEvent.setup();
    vi.spyOn(apiModule.api, "register").mockRejectedValue(
      new apiModule.ApiError(409, "Email already registered"),
    );

    renderWithProviders(<RegisterForm background={null} />);

    await user.type(getField(/^Display name/), "Snack Fan");
    await user.type(getField(/^Email/), "snacker@example.com");
    await user.type(getField(/^Password/), "secret123");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    expect(await screen.findByText("Email already registered")).toBeInTheDocument();
  });
});
