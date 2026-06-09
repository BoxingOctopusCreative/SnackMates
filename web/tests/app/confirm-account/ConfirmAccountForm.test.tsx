import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ConfirmAccountForm from "@/app/confirm-account/ConfirmAccountForm";
import * as apiModule from "@/lib/api";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders } from "@test/utils";

describe("ConfirmAccountForm", () => {
  it("confirms account actions from the URL", async () => {
    navigationMocks.searchParams = new URLSearchParams("token=action-token");
    vi.spyOn(apiModule.api, "confirmAccountAction").mockResolvedValue({ ok: true, action: "deactivate" });
    const clearToken = vi.spyOn(apiModule, "clearToken");

    renderWithProviders(<ConfirmAccountForm />);

    expect(
      await screen.findByText(
        "Your account has been deactivated. You can reactivate it anytime from the sign-in page.",
      ),
    ).toBeInTheDocument();
    expect(apiModule.api.confirmAccountAction).toHaveBeenCalledWith("action-token");
    expect(clearToken).toHaveBeenCalled();
  });

  it("shows an error when the token is missing", async () => {
    navigationMocks.searchParams = new URLSearchParams();
    renderWithProviders(<ConfirmAccountForm />);

    expect(await screen.findByText("Missing confirmation token.")).toBeInTheDocument();
  });

  it("shows confirmation failures", async () => {
    navigationMocks.searchParams = new URLSearchParams("token=bad-token");
    vi.spyOn(apiModule.api, "confirmAccountAction").mockRejectedValue(new Error("Token expired"));

    renderWithProviders(<ConfirmAccountForm />);

    await waitFor(() => {
      expect(screen.getByText("Token expired")).toBeInTheDocument();
    });
  });
});
