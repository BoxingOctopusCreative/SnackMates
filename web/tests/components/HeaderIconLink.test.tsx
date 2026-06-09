import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { HeaderIconLink } from "@/components/HeaderIconLink";
import { renderWithProviders } from "@test/utils";

describe("HeaderIconLink", () => {
  it("renders a labeled navigation link", () => {
    renderWithProviders(
      <HeaderIconLink
        href="/wishlists"
        label="Wishlist"
        outlineIcon="ion:heart-outline"
        filledIcon="ion:heart"
      />,
    );

    expect(screen.getByRole("link", { name: "Wishlist" })).toHaveAttribute("href", "/wishlists");
  });

  it("uses the filled icon on hover", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <HeaderIconLink
        href="/matches"
        label="Match Me"
        outlineIcon="ion:people-outline"
        filledIcon="ion:people"
      />,
    );

    const link = screen.getByRole("link", { name: "Match Me" });
    await user.hover(link);
    expect(link).toHaveAttribute("href", "/matches");
  });
});
