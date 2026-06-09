"use client";

import { Icon } from "@iconify/react";
import { ActionButton, Tooltip, TooltipTrigger } from "@adobe/react-spectrum";
import { useTheme } from "@/components/ThemeProvider";

type ThemeSwitcherProps = {
  /** Use on brand header bars; defaults to floating placement on other pages. */
  variant?: "header" | "floating";
};

export function ThemeSwitcher({ variant = "floating" }: ThemeSwitcherProps) {
  const { colorScheme, toggleColorScheme } = useTheme();
  const isDark = colorScheme === "dark";
  const label = isDark ? "Switch to light mode" : "Switch to dark mode";
  const iconColor = variant === "header" ? "var(--sm-header-accent)" : "var(--sm-text)";

  const button = (
    <ActionButton
      isQuiet
      aria-label={label}
      onPress={toggleColorScheme}
      UNSAFE_style={{
        color: iconColor,
        minWidth: "auto",
      }}
    >
      <Icon
        icon={isDark ? "ion:sunny-outline" : "ion:moon-outline"}
        className="sm-theme-switcher__icon"
        aria-hidden
      />
    </ActionButton>
  );

  if (variant === "floating") {
    return (
      <div
        style={{
          position: "fixed",
          top: "0.75rem",
          right: "0.75rem",
          zIndex: 1000,
          backgroundColor: "var(--sm-surface)",
          border: "1px solid var(--sm-border)",
          borderRadius: "9999px",
          boxShadow: "0 1px 4px rgba(71, 18, 14, 0.12)",
        }}
      >
        <TooltipTrigger>
          {button}
          <Tooltip>{label}</Tooltip>
        </TooltipTrigger>
      </div>
    );
  }

  return (
    <TooltipTrigger>
      {button}
      <Tooltip>{label}</Tooltip>
    </TooltipTrigger>
  );
}
