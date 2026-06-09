import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { DiscordOAuthButton } from "@/components/DiscordOAuthButton";
import { renderWithProviders } from "@test/utils";

describe("DiscordOAuthButton", () => {
  it("renders Discord branding and navigates on press", async () => {
    const user = userEvent.setup();
    Object.defineProperty(window, "location", {
      configurable: true,
      value: { ...window.location, href: "" },
    });

    renderWithProviders(
      <DiscordOAuthButton href="https://api.example.com/api/v1/auth/discord">
        Continue with Discord
      </DiscordOAuthButton>,
    );

    const button = screen.getByRole("button", { name: "Continue with Discord" });
    expect(button).toHaveClass("sm-discord-oauth-button");

    await user.click(button);
    expect(window.location.href).toBe("https://api.example.com/api/v1/auth/discord");
  });
});
