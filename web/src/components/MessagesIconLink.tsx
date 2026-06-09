"use client";

import { Icon } from "@iconify/react";
import Link from "next/link";
import { useState } from "react";
import { useMessagesInbox } from "@/components/MessagesProvider";

export function MessagesIconLink() {
  const { unreadCount } = useMessagesInbox();
  const [hovered, setHovered] = useState(false);

  return (
    <Link
      href="/messages"
      aria-label={unreadCount > 0 ? `Messages, ${unreadCount} unread` : "Messages"}
      title="Messages"
      className="sm-header-icon-link sm-header-icon-link--badge"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocus={() => setHovered(true)}
      onBlur={() => setHovered(false)}
    >
      <Icon
        icon={hovered ? "ion:mail" : "ion:mail-outline"}
        className="sm-header-icon"
        aria-hidden
      />
      {unreadCount > 0 && (
        <span className="sm-header-icon-badge" aria-hidden>
          {unreadCount > 99 ? "99+" : unreadCount}
        </span>
      )}
    </Link>
  );
}
