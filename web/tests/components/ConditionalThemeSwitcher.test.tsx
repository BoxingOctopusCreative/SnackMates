import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ConditionalThemeSwitcher } from "@/components/ConditionalThemeSwitcher";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders } from "@test/utils";

describe("ConditionalThemeSwitcher", () => {
  it("hides the switcher on the home page", () => {
    navigationMocks.pathname = "/";
    renderWithProviders(<ConditionalThemeSwitcher />);
    expect(screen.queryByRole("button", { name: /switch to/i })).not.toBeInTheDocument();
  });

  it("hides the switcher on authenticated app routes", () => {
    navigationMocks.pathname = "/dashboard";
    renderWithProviders(<ConditionalThemeSwitcher />);
    expect(screen.queryByRole("button", { name: /switch to/i })).not.toBeInTheDocument();
  });

  it("hides the switcher on sign-in and registration pages", () => {
    navigationMocks.pathname = "/login";
    renderWithProviders(<ConditionalThemeSwitcher />);
    expect(screen.queryByRole("button", { name: /switch to/i })).not.toBeInTheDocument();

    navigationMocks.pathname = "/register";
    renderWithProviders(<ConditionalThemeSwitcher />);
    expect(screen.queryByRole("button", { name: /switch to/i })).not.toBeInTheDocument();
  });

  it("hides the switcher on authenticated app routes such as messages", () => {
    navigationMocks.pathname = "/messages";
    renderWithProviders(<ConditionalThemeSwitcher />);
    expect(screen.queryByRole("button", { name: /switch to/i })).not.toBeInTheDocument();
  });

  it("shows the switcher on public utility routes without a profile menu", () => {
    navigationMocks.pathname = "/forgot-password";
    renderWithProviders(<ConditionalThemeSwitcher />);
    expect(screen.getByRole("button", { name: /switch to/i })).toBeInTheDocument();
  });
});
