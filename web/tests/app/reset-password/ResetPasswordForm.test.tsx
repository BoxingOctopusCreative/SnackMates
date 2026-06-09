import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import ResetPasswordForm from "@/app/reset-password/ResetPasswordForm";
import * as apiModule from "@/lib/api";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders, getField } from "@test/utils";

describe("ResetPasswordForm", () => {
  it("updates the password and redirects to login", async () => {
    const user = userEvent.setup();
    navigationMocks.searchParams = new URLSearchParams("token=reset-token");
    vi.spyOn(apiModule.api, "resetPassword").mockResolvedValue({ ok: true });

    renderWithProviders(<ResetPasswordForm />);

    await user.type(getField(/^New password/), "new-secret");
    await user.click(screen.getByRole("button", { name: "Update password" }));

    await waitFor(() => {
      expect(apiModule.api.resetPassword).toHaveBeenCalledWith("reset-token", "new-secret");
      expect(navigationMocks.push).toHaveBeenCalledWith("/login");
    });
  });

  it("disables submit when the token is missing", () => {
    navigationMocks.searchParams = new URLSearchParams();
    renderWithProviders(<ResetPasswordForm />);
    expect(screen.getByRole("button", { name: "Update password" })).toBeDisabled();
  });
});
