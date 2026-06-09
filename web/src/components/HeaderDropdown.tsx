"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useId,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { createPortal } from "react-dom";
import { Provider, defaultTheme } from "@adobe/react-spectrum";
import { useTheme } from "@/components/ThemeProvider";

type PanelPosition = {
  top: number;
  right: number;
};

export type HeaderDropdownTriggerState = {
  open: boolean;
  toggle: () => void;
  panelId: string;
};

const HeaderDropdownCloseContext = createContext<() => void>(() => {});

export function useHeaderDropdownClose() {
  return useContext(HeaderDropdownCloseContext);
}

export type HeaderDropdownProps = {
  title: string;
  children: ReactNode;
  renderTrigger: (state: HeaderDropdownTriggerState) => ReactNode;
  onOpen?: () => void | Promise<void>;
  rootClassName?: string;
  bodyClassName?: string;
  panelWidth?: "narrow" | "medium";
};

export function HeaderDropdown({
  title,
  children,
  renderTrigger,
  onOpen,
  rootClassName,
  bodyClassName,
  panelWidth = "medium",
}: HeaderDropdownProps) {
  const { colorScheme } = useTheme();
  const [open, setOpen] = useState(false);
  const [panelPos, setPanelPos] = useState<PanelPosition>({ top: 0, right: 0 });
  const [mounted, setMounted] = useState(false);
  const panelId = useId();
  const titleId = useId();
  const rootRef = useRef<HTMLDivElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  const updatePanelPosition = useCallback(() => {
    if (!rootRef.current) return;
    const rect = rootRef.current.getBoundingClientRect();
    setPanelPos({
      top: rect.bottom + 8,
      right: Math.max(8, window.innerWidth - rect.right),
    });
  }, []);

  const close = useCallback(() => setOpen(false), []);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!open) return;

    updatePanelPosition();
    window.addEventListener("resize", updatePanelPosition);
    window.addEventListener("scroll", updatePanelPosition, true);

    function handlePointerDown(event: MouseEvent) {
      const target = event.target as Node;
      if (rootRef.current?.contains(target) || panelRef.current?.contains(target)) {
        return;
      }
      close();
    }

    function handleEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        close();
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleEscape);
    return () => {
      window.removeEventListener("resize", updatePanelPosition);
      window.removeEventListener("scroll", updatePanelPosition, true);
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [close, open, updatePanelPosition]);

  async function toggle() {
    const nextOpen = !open;
    setOpen(nextOpen);
    if (!nextOpen) return;

    updatePanelPosition();
    await onOpen?.();
  }

  const panel =
    open && mounted ? (
      <div
        ref={panelRef}
        id={panelId}
        className="sm-header-dropdown"
        style={{ top: panelPos.top, right: panelPos.right }}
        role="dialog"
        aria-labelledby={titleId}
        aria-modal="false"
        onClick={(event) => event.stopPropagation()}
      >
        <div className={`sm-modal__panel sm-modal__panel--${panelWidth}`}>
          <Provider
            theme={defaultTheme}
            colorScheme={colorScheme}
            locale="en-US"
            UNSAFE_className="sm-theme sm-modal__shell"
          >
            <div className="sm-modal__header">
              <h2 id={titleId} className="sm-modal__title">
                {title}
              </h2>
            </div>
            <HeaderDropdownCloseContext.Provider value={close}>
              <div
                className={`sm-modal__body sm-modal__body--scrollable sm-header-dropdown__body sm-user-menu${bodyClassName ? ` ${bodyClassName}` : ""}`}
              >
                {children}
              </div>
            </HeaderDropdownCloseContext.Provider>
          </Provider>
        </div>
      </div>
    ) : null;

  return (
    <>
      <div ref={rootRef} className={rootClassName}>
        {renderTrigger({ open, toggle, panelId })}
      </div>
      {mounted && panel ? createPortal(panel, document.body) : null}
    </>
  );
}
