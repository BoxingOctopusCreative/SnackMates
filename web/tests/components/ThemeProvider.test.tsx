import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { ThemeProvider, useTheme } from "@/components/ThemeProvider";

function ThemeConsumer() {
  const { colorScheme, toggleColorScheme, setColorScheme } = useTheme();

  return (
    <div>
      <span data-testid="scheme">{colorScheme}</span>
      <button type="button" onClick={toggleColorScheme}>
        Toggle
      </button>
      <button type="button" onClick={() => setColorScheme("dark")}>
        Dark
      </button>
    </div>
  );
}

describe("ThemeProvider", () => {
  it("persists the selected color scheme", async () => {
    const user = userEvent.setup();
    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>,
    );

    expect(screen.getByTestId("scheme")).toHaveTextContent("light");
    await user.click(screen.getByRole("button", { name: "Dark" }));
    expect(screen.getByTestId("scheme")).toHaveTextContent("dark");
    expect(document.documentElement.dataset.theme).toBe("dark");
    expect(localStorage.getItem("sm-color-scheme")).toBe("dark");
  });

  it("reads a stored color scheme on mount", () => {
    localStorage.setItem("sm-color-scheme", "dark");

    render(
      <ThemeProvider>
        <ThemeConsumer />
      </ThemeProvider>,
    );

    expect(screen.getByTestId("scheme")).toHaveTextContent("dark");
  });
});

describe("useTheme", () => {
  it("throws outside ThemeProvider", () => {
    function BrokenConsumer() {
      useTheme();
      return null;
    }

    expect(() => render(<BrokenConsumer />)).toThrow(
      "useTheme must be used within ThemeProvider",
    );
  });
});
