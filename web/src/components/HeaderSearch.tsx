"use client";

import { Icon } from "@iconify/react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import {
  createContext,
  Suspense,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { SearchField } from "@adobe/react-spectrum";

type HeaderSearchContextValue = {
  open: boolean;
  toggle: () => void;
  close: () => void;
};

const HeaderSearchContext = createContext<HeaderSearchContextValue | null>(null);

function useHeaderSearch() {
  const context = useContext(HeaderSearchContext);
  if (!context) {
    throw new Error("Header search components must be used within HeaderSearchProvider");
  }
  return context;
}

export function HeaderSearchProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const [openPath, setOpenPath] = useState<string | null>(null);
  const open = openPath === pathname;

  const value: HeaderSearchContextValue = {
    open,
    toggle: () =>
      setOpenPath((current) => (current === pathname ? null : pathname)),
    close: () => setOpenPath(null),
  };

  return (
    <HeaderSearchContext.Provider value={value}>{children}</HeaderSearchContext.Provider>
  );
}

function HeaderSearchField({ onSubmitComplete }: { onSubmitComplete?: () => void }) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const urlQuery = pathname === "/search" ? (searchParams.get("q") ?? "") : "";
  const [draft, setDraft] = useState<{ forUrlQuery: string; value: string } | null>(
    null,
  );
  const query = draft?.forUrlQuery === urlQuery ? draft.value : urlQuery;
  const setQuery = (value: string) => setDraft({ forUrlQuery: urlQuery, value });

  function submit() {
    const trimmed = query.trim();
    if (!trimmed) return;
    router.push(`/search?q=${encodeURIComponent(trimmed)}`);
    onSubmitComplete?.();
  }

  return (
    <SearchField
      aria-label="Search SnackMates"
      placeholder="Search SnackMates..."
      value={query}
      onChange={setQuery}
      onSubmit={submit}
      width="100%"
    />
  );
}

export function HeaderSearchInline() {
  return (
    <div className="sm-header-search-inline">
      <Suspense fallback={null}>
        <HeaderSearchField />
      </Suspense>
    </div>
  );
}

export function HeaderSearchToggle() {
  const { open, toggle } = useHeaderSearch();
  const [hovered, setHovered] = useState(false);

  return (
    <button
      type="button"
      className="sm-header-search-toggle sm-header-icon-link"
      aria-label={open ? "Close search" : "Open search"}
      aria-expanded={open}
      onClick={toggle}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocus={() => setHovered(true)}
      onBlur={() => setHovered(false)}
    >
      <Icon
        icon={
          open
            ? "ion:close"
            : hovered
              ? "ion:search"
              : "ion:search-outline"
        }
        className="sm-header-icon"
        aria-hidden
      />
    </button>
  );
}

export function HeaderSearchPanel() {
  const { open, close } = useHeaderSearch();
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const input = panelRef.current?.querySelector("input");
    input?.focus();
  }, [open]);

  if (!open) return null;

  return (
    <div ref={panelRef} className="sm-header-search-panel">
      <Suspense fallback={null}>
        <HeaderSearchField onSubmitComplete={close} />
      </Suspense>
    </div>
  );
}
