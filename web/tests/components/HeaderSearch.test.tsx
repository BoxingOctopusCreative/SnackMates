import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import {
  HeaderSearchPanel,
  HeaderSearchProvider,
  HeaderSearchToggle,
} from "@/components/HeaderSearch";
import { navigationMocks } from "@test/navigation";
import { renderWithProviders } from "@test/utils";

function SearchHarness() {
  return (
    <HeaderSearchProvider>
      <HeaderSearchToggle />
      <HeaderSearchPanel />
    </HeaderSearchProvider>
  );
}

describe("HeaderSearch", () => {
  it("opens and closes the search panel", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SearchHarness />);

    expect(screen.queryByRole("searchbox", { name: "Search SnackMates" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Open search" }));
    expect(screen.getByRole("searchbox", { name: "Search SnackMates" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Close search" }));
    expect(screen.queryByRole("searchbox", { name: "Search SnackMates" })).not.toBeInTheDocument();
  });

  it("routes to the search page on submit", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SearchHarness />);

    await user.click(screen.getByRole("button", { name: "Open search" }));
    const field = screen.getByRole("searchbox", { name: "Search SnackMates" });
    await user.type(field, "pocky{enter}");

    expect(navigationMocks.push).toHaveBeenCalledWith("/search?q=pocky");
  });

  it("closes the panel after submitting from the panel field", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SearchHarness />);

    await user.click(screen.getByRole("button", { name: "Open search" }));
    await user.type(screen.getByRole("searchbox", { name: "Search SnackMates" }), "chips{enter}");

    expect(screen.queryByRole("searchbox", { name: "Search SnackMates" })).not.toBeInTheDocument();
  });
});
