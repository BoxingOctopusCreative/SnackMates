import "@testing-library/jest-dom/vitest";
import React from "react";
import { afterEach, beforeEach, vi } from "vitest";
import { navigationMocks, resetNavigationMocks } from "@test/navigation";

class LocalStorageMock {
  private store = new Map<string, string>();

  clear() {
    this.store.clear();
  }

  getItem(key: string) {
    return this.store.get(key) ?? null;
  }

  setItem(key: string, value: string) {
    this.store.set(key, value);
  }

  removeItem(key: string) {
    this.store.delete(key);
  }
}

const localStorageMock = new LocalStorageMock();
vi.stubGlobal("localStorage", localStorageMock);

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: query.includes("dark") ? false : false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: navigationMocks.push,
    replace: navigationMocks.replace,
  }),
  usePathname: () => navigationMocks.pathname,
  useSearchParams: () => navigationMocks.searchParams,
  useParams: () => navigationMocks.params,
}));

vi.mock("next/image", () => ({
  default: ({
    src,
    alt,
    width,
    height,
    priority: _priority,
    fill: _fill,
    ...rest
  }: {
    src: string;
    alt: string;
    width?: number;
    height?: number;
    priority?: boolean;
    fill?: boolean;
  }) =>
    React.createElement("img", {
      src,
      alt,
      width,
      height,
      ...rest,
    }),
}));

beforeEach(() => {
  resetNavigationMocks();
  localStorageMock.clear();
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.clearAllMocks();
});
