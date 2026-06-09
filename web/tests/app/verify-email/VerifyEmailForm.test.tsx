import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import VerifyEmailForm from "@/app/verify-email/VerifyEmailForm";
import * as apiModule from "@/lib/api";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders } from "@test/utils";

describe("VerifyEmailForm", () => {
  it("verifies email tokens from the URL", async () => {
    navigationMocks.searchParams = new URLSearchParams("token=verify-token");
    vi.spyOn(apiModule.api, "verifyEmail").mockResolvedValue({ ok: true });

    renderWithProviders(<VerifyEmailForm />);

    expect(
      await screen.findByText("Email verified! You can sign in and start building wishlists."),
    ).toBeInTheDocument();
    expect(apiModule.api.verifyEmail).toHaveBeenCalledWith("verify-token");
  });

  it("shows an error when the token is missing", async () => {
    navigationMocks.searchParams = new URLSearchParams();
    renderWithProviders(<VerifyEmailForm />);

    expect(await screen.findByText("Missing verification token.")).toBeInTheDocument();
  });

  it("shows verification failures", async () => {
    navigationMocks.searchParams = new URLSearchParams("token=bad-token");
    vi.spyOn(apiModule.api, "verifyEmail").mockRejectedValue(new Error("Token expired"));

    renderWithProviders(<VerifyEmailForm />);

    await waitFor(() => {
      expect(screen.getByText("Token expired")).toBeInTheDocument();
    });
  });
});
