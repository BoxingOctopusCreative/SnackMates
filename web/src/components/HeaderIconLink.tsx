"use client";

import { Icon } from "@iconify/react";
import Link from "next/link";
import { useState } from "react";

type HeaderIconLinkProps = {
  href: string;
  label: string;
  outlineIcon: string;
  filledIcon: string;
};

export function HeaderIconLink({
  href,
  label,
  outlineIcon,
  filledIcon,
}: HeaderIconLinkProps) {
  const [hovered, setHovered] = useState(false);

  return (
    <Link
      href={href}
      aria-label={label}
      title={label}
      className="sm-header-icon-link"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocus={() => setHovered(true)}
      onBlur={() => setHovered(false)}
    >
      <Icon
        icon={hovered ? filledIcon : outlineIcon}
        className="sm-header-icon"
        aria-hidden
      />
    </Link>
  );
}
