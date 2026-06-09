import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import ForgotPasswordPage from "@/app/forgot-password/page";
import * as apiModule from "@/lib/api";
import { renderWithProviders, getField } from "@test/utils";

describe("ForgotPasswordPage", () => {
  it("requests a password reset email", async () => {
    const user = userEvent.setup();
    vi.spyOn(apiModule.api, "forgotPassword").mockResolvedValue({
      message: "If that email exists, a reset link was sent.",
    });

    renderWithProviders(<ForgotPasswordPage />);

    await user.type(getField(/^Email/), "snacker@example.com");
    await user.click(screen.getByRole("button", { name: "Send reset link" }));

    expect(
      await screen.findByText("If that email exists, a reset link was sent."),
    ).toBeInTheDocument();
  });

  it("shows request failures", async () => {
    const user = userEvent.setup();
    vi.spyOn(apiModule.api, "forgotPassword").mockRejectedValue(new Error("Network down"));

    renderWithProviders(<ForgotPasswordPage />);

    await user.type(getField(/^Email/), "snacker@example.com");
    await user.click(screen.getByRole("button", { name: "Send reset link" }));

    expect(await screen.findByText("Network down")).toBeInTheDocument();
  });
});
