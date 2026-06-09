"use client";

import { Provider, defaultTheme, type ColorScheme } from "@adobe/react-spectrum";
import { ToastContainer } from "@react-spectrum/toast";
import { createContext, useContext, useSyncExternalStore } from "react";

const STORAGE_KEY = "sm-color-scheme";

type ThemeContextValue = {
  colorScheme: ColorScheme;
  setColorScheme: (scheme: ColorScheme) => void;
  toggleColorScheme: () => void;
};

const ThemeContext = createContext<ThemeContextValue | null>(null);

const themeListeners = new Set<() => void>();
let activeColorScheme: ColorScheme | null = null;

function notifyThemeListeners() {
  themeListeners.forEach((listener) => listener());
}

function subscribeToTheme(listener: () => void) {
  themeListeners.add(listener);
  if (typeof window === "undefined") {
    return () => themeListeners.delete(listener);
  }

  const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
  const onMediaChange = () => {
    if (activeColorScheme === null) notifyThemeListeners();
  };
  const onStorage = (event: StorageEvent) => {
    if (event.key === STORAGE_KEY) {
      activeColorScheme = null;
      notifyThemeListeners();
    }
  };

  mediaQuery.addEventListener("change", onMediaChange);
  window.addEventListener("storage", onStorage);

  return () => {
    themeListeners.delete(listener);
    mediaQuery.removeEventListener("change", onMediaChange);
    window.removeEventListener("storage", onStorage);
  };
}

function readStoredColorScheme(): ColorScheme {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "light" || stored === "dark") return stored;
  } catch {
    // ignore
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function readColorSchemeSnapshot(): ColorScheme {
  if (activeColorScheme !== null) return activeColorScheme;
  if (typeof window === "undefined") return "light";

  const fromDom = document.documentElement.dataset.theme;
  if (fromDom === "light" || fromDom === "dark") return fromDom;

  return readStoredColorScheme();
}

function applyDocumentTheme(scheme: ColorScheme) {
  document.documentElement.dataset.theme = scheme;
  document.documentElement.style.colorScheme = scheme;
  try {
    localStorage.setItem(STORAGE_KEY, scheme);
  } catch {
    // ignore
  }
}

function setColorSchemeValue(scheme: ColorScheme) {
  activeColorScheme = scheme;
  applyDocumentTheme(scheme);
  notifyThemeListeners();
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const colorScheme = useSyncExternalStore<ColorScheme>(
    subscribeToTheme,
    readColorSchemeSnapshot,
    (): ColorScheme => "light",
  );

  const setColorScheme = (scheme: ColorScheme) => setColorSchemeValue(scheme);
  const toggleColorScheme = () => {
    setColorSchemeValue(colorScheme === "light" ? "dark" : "light");
  };

  return (
    <ThemeContext.Provider value={{ colorScheme, setColorScheme, toggleColorScheme }}>
      <Provider
        theme={defaultTheme}
        colorScheme={colorScheme}
        locale="en-US"
        UNSAFE_className="sm-theme"
      >
        <ToastContainer placement="top end" />
        {children}
      </Provider>
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error("useTheme must be used within ThemeProvider");
  }
  return context;
}
