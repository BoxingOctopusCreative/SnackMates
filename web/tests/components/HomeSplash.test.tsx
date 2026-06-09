import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { HomeSplash } from "@/components/HomeSplash";

describe("HomeSplash", () => {
  it("renders the landing page CTAs", () => {
    render(<HomeSplash background={null} />);

    expect(screen.getByRole("img", { name: "SnackMates" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /get started/i })).toHaveAttribute("href", "/register");
    expect(screen.getByRole("link", { name: /sign in/i })).toHaveAttribute("href", "/login");
    expect(screen.getByRole("link", { name: /terms of use/i })).toHaveAttribute("href", "/terms-of-use");
    expect(screen.getByRole("link", { name: /privacy policy/i })).toHaveAttribute(
      "href",
      "/privacy-policy",
    );
    expect(screen.getByRole("link", { name: /sitemap/i })).toHaveAttribute("href", "/sitemap");
    expect(screen.getByRole("link", { name: /boxing octopus creative/i })).toHaveAttribute(
      "href",
      "https://boxingoctop.us",
    );
    expect(screen.getByText(new RegExp(`Copyright ${new Date().getFullYear()} All rights reserved`))).toBeInTheDocument();
  });

  it("shows Unsplash attribution when a background is provided", () => {
    render(
      <HomeSplash
        background={{
          url: "https://images.unsplash.com/photo-1",
          photographer: "Alex",
          photographerUrl: "https://unsplash.com/@alex",
          unsplashUrl: "https://unsplash.com/photos/1",
        }}
      />,
    );

    expect(screen.getByRole("link", { name: "Alex" })).toHaveAttribute(
      "href",
      expect.stringContaining("unsplash.com/@alex"),
    );
    expect(screen.getByRole("link", { name: "Unsplash" })).toBeInTheDocument();
  });
});
