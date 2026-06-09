import { render, screen, type RenderOptions } from "@testing-library/react";
import type { ReactElement, ReactNode } from "react";
import { vi } from "vitest";
import { SettingsModalProvider } from "@/components/SettingsModalProvider";
import { ThemeProvider } from "@/components/ThemeProvider";

export function renderWithProviders(ui: ReactElement, options?: RenderOptions) {
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <ThemeProvider>
        <SettingsModalProvider>{children}</SettingsModalProvider>
      </ThemeProvider>
    );
  }

  return render(ui, { wrapper: Wrapper, ...options });
}

export function mockJsonResponse(data: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? "OK" : "Error",
    json: () => Promise.resolve(data),
  } as Response;
}

export function mockFetchJson(data: unknown, status = 200) {
  return vi.fn().mockResolvedValue(mockJsonResponse(data, status));
}

export function getField(label: string | RegExp) {
  return screen.getByLabelText(label);
}
