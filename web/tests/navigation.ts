import { vi } from "vitest";

export const navigationMocks = {
  push: vi.fn(),
  replace: vi.fn(),
  pathname: "/",
  searchParams: new URLSearchParams(),
  params: {} as Record<string, string>,
};

export function resetNavigationMocks() {
  navigationMocks.push.mockReset();
  navigationMocks.replace.mockReset();
  navigationMocks.pathname = "/";
  navigationMocks.searchParams = new URLSearchParams();
  navigationMocks.params = {};
}
