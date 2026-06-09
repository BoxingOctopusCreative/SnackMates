import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import MatchesPage from "@/app/(app)/matches/page";
import * as apiModule from "@/lib/api";
import { mockSnackMate } from "@test/fixtures";
import { renderWithProviders } from "@test/utils";

describe("MatchesPage", () => {
  beforeEach(() => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue({
      ok: true,
      json: async () => ({
        photo: {
          url: "https://images.unsplash.com/photo-1",
          photographer: "Alex",
          photographerUrl: "https://unsplash.com/@alex",
          unsplashUrl: "https://unsplash.com/photos/1",
        },
      }),
    } as Response);
  });

  it("loads snack mates and shows mate details", async () => {
    vi.spyOn(apiModule.api, "friends").mockResolvedValue([mockSnackMate]);
    vi.spyOn(apiModule.api, "matches").mockResolvedValue([]);

    renderWithProviders(<MatchesPage />);

    expect(await screen.findByRole("heading", { name: "Mate Snacker" })).toBeInTheDocument();
    expect(screen.getByText("Japan")).toBeInTheDocument();
  });

  it("shows an empty state when there are no snack mates", async () => {
    vi.spyOn(apiModule.api, "friends").mockResolvedValue([]);
    vi.spyOn(apiModule.api, "matches").mockResolvedValue([]);

    renderWithProviders(<MatchesPage />);

    expect(await screen.findByRole("heading", { name: "No Snack Mates Yet" })).toBeInTheDocument();
  });
});
