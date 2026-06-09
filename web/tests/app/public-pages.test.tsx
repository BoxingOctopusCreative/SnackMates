import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import PrivacyPolicyPage from "@/app/privacy-policy/page";
import SitemapPage from "@/app/sitemap/page";
import TermsOfUsePage from "@/app/terms-of-use/page";

const termsMarkdown = `_Last updated: June 9, 2026_

Welcome to SnackMates.

## 1. Eligibility

You must be at least 13 years old to use SnackMates.`;

const privacyMarkdown = `_Last updated: June 9, 2026_

Boxing Octopus Creative respects your privacy.

## 1. Information we collect

We collect information you provide directly.`;

vi.mock("@/lib/unsplash", () => ({
  DEFAULT_UNSPLASH_QUERY: "cartoon snack food",
  fetchRandomUnsplashPhoto: vi.fn().mockResolvedValue({
    url: "https://images.unsplash.com/photo-1",
    photographer: "Alex",
    photographerUrl: "https://unsplash.com/@alex",
    unsplashUrl: "https://unsplash.com/photos/1",
  }),
}));

vi.mock("@/lib/markdown", () => ({
  fetchLegalMarkdown: vi.fn(async (filename: string) => {
    if (filename === "terms-of-use.md") return termsMarkdown;
    if (filename === "privacy-policy.md") return privacyMarkdown;
    throw new Error(`Unexpected legal markdown file: ${filename}`);
  }),
}));

describe("TermsOfUsePage", () => {
  it("renders markdown content with typography styles", async () => {
    const ui = await TermsOfUsePage();
    render(ui);

    expect(screen.getByRole("heading", { name: "Terms of Use", level: 1 })).toHaveClass("sm-hero-title");
    expect(screen.getByRole("heading", { name: /eligibility/i, level: 2 })).toBeInTheDocument();
    expect(screen.getByText(/Welcome to SnackMates/)).toBeInTheDocument();
    expect(document.querySelector(".sm-prose")).not.toBeNull();
    expect(document.querySelector(".sm-public-document-page__panel")).not.toBeNull();
  });
});

describe("PrivacyPolicyPage", () => {
  it("renders markdown content with typography styles", async () => {
    const ui = await PrivacyPolicyPage();
    render(ui);

    expect(screen.getByRole("heading", { name: "Privacy Policy", level: 1 })).toHaveClass(
      "sm-hero-title",
    );
    expect(screen.getByRole("heading", { name: /information we collect/i, level: 2 })).toBeInTheDocument();
    expect(screen.getByText(/respects your privacy/i)).toBeInTheDocument();
    expect(document.querySelector(".sm-prose")).not.toBeNull();
    expect(document.querySelector(".sm-public-document-page__panel")).not.toBeNull();
  });
});

describe("SitemapPage", () => {
  it("lists public pages", async () => {
    const ui = await SitemapPage();
    render(ui);

    expect(screen.getByRole("heading", { name: "Sitemap", level: 1 })).toHaveClass("sm-hero-title");
    expect(screen.getByRole("link", { name: "Home" })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "Terms of use" })).toHaveAttribute("href", "/terms-of-use");
    expect(screen.getByRole("link", { name: "Privacy policy" })).toHaveAttribute(
      "href",
      "/privacy-policy",
    );
  });
});
