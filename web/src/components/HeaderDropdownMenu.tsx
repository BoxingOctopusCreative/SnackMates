"use client";

import { type ReactNode } from "react";
import { useHeaderDropdownClose } from "@/components/HeaderDropdown";

export function HeaderDropdownMenu({ children }: { children: ReactNode }) {
  return (
    <div className="sm-header-dropdown__menu" role="menu">
      {children}
    </div>
  );
}

type HeaderDropdownItemProps = {
  children: ReactNode;
  onPress: () => void;
  variant?: "default" | "danger";
  disabled?: boolean;
};

export function HeaderDropdownItem({
  children,
  onPress,
  variant = "default",
  disabled = false,
}: HeaderDropdownItemProps) {
  const close = useHeaderDropdownClose();

  return (
    <button
      type="button"
      role="menuitem"
      className={`sm-header-dropdown__item${variant === "danger" ? " sm-header-dropdown__item--danger" : ""}`}
      onClick={() => {
        close();
        onPress();
      }}
      disabled={disabled}
    >
      {children}
    </button>
  );
}

export function HeaderDropdownSection({ title, children }: { title?: string; children: ReactNode }) {
  return (
    <div className="sm-header-dropdown__section" role="group" aria-label={title}>
      {title ? <div className="sm-header-dropdown__section-heading">{title}</div> : null}
      {children}
    </div>
  );
}

export function HeaderDropdownDivider() {
  return <div className="sm-header-dropdown__divider" role="separator" />;
}
