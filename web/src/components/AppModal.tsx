"use client";

import { Icon } from "@iconify/react";
import { Provider, defaultTheme } from "@adobe/react-spectrum";
import { useEffect, useId, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { useTheme } from "@/components/ThemeProvider";

type AppModalSize = "narrow" | "medium" | "wide";
type AppModalAlign = "center" | "below-header";

type AppModalProps = {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  titleId?: string;
  children: ReactNode;
  size?: AppModalSize;
  align?: AppModalAlign;
  showCloseButton?: boolean;
  scrollableBody?: boolean;
  bodyClassName?: string;
};

function updateModalTop() {
  const header = document.querySelector(".sm-app-header");
  const top = header?.getBoundingClientRect().bottom ?? 88;
  document.documentElement.style.setProperty("--sm-modal-top", `${top}px`);
}

export function AppModal({
  isOpen,
  onClose,
  title,
  titleId,
  children,
  size = "narrow",
  align = "center",
  showCloseButton = false,
  scrollableBody = false,
  bodyClassName,
}: AppModalProps) {
  const { colorScheme } = useTheme();
  const generatedTitleId = useId();
  const resolvedTitleId = titleId ?? generatedTitleId;

  useEffect(() => {
    if (!isOpen) return;

    if (align === "below-header") {
      updateModalTop();
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") onClose();
    }

    const onResize = align === "below-header" ? updateModalTop : undefined;
    if (onResize) window.addEventListener("resize", onResize);
    document.addEventListener("keydown", onKeyDown);

    return () => {
      if (onResize) window.removeEventListener("resize", onResize);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [align, isOpen, onClose]);

  if (!isOpen || typeof document === "undefined") return null;

  const overlayClass = [
    "sm-modal",
    align === "below-header" ? "sm-modal--below-header" : "sm-modal--center",
  ].join(" ");

  const panelClass = ["sm-modal__panel", `sm-modal__panel--${size}`].join(" ");

  const resolvedBodyClass = [
    "sm-modal__body",
    scrollableBody ? "sm-modal__body--scrollable" : undefined,
    bodyClassName,
  ]
    .filter(Boolean)
    .join(" ");

  return createPortal(
    <div className={overlayClass} onClick={onClose}>
      <div
        className={panelClass}
        role="dialog"
        aria-labelledby={resolvedTitleId}
        aria-modal="true"
        onClick={(event) => event.stopPropagation()}
      >
        <Provider
          theme={defaultTheme}
          colorScheme={colorScheme}
          locale="en-US"
          UNSAFE_className="sm-theme sm-modal__shell"
        >
          <div className="sm-modal__header">
            <h2 id={resolvedTitleId} className="sm-modal__title">
              {title}
            </h2>
            {showCloseButton && (
              <button
                type="button"
                className="sm-modal__close"
                aria-label={`Close ${title.toLowerCase()}`}
                onClick={onClose}
              >
                <Icon icon="ion:close" className="sm-modal__close-icon" aria-hidden />
              </button>
            )}
          </div>
          <div className={resolvedBodyClass}>{children}</div>
        </Provider>
      </div>
    </div>,
    document.body,
  );
}
