"use client";

import { useRouter } from "next/navigation";
import { ActionButton, Avatar } from "@adobe/react-spectrum";
import { avatarImageSrc } from "@/lib/avatar";
import { clearToken, User } from "@/lib/api";
import { HeaderDropdown } from "@/components/HeaderDropdown";
import {
  HeaderDropdownDivider,
  HeaderDropdownItem,
  HeaderDropdownMenu,
  HeaderDropdownSection,
} from "@/components/HeaderDropdownMenu";
import { useSettingsModal } from "@/components/SettingsModalProvider";
import { useTheme } from "@/components/ThemeProvider";

export function UserMenu({ user }: { user: User }) {
  const router = useRouter();
  const { openSettings } = useSettingsModal();
  const { colorScheme, setColorScheme } = useTheme();

  return (
    <HeaderDropdown
      title="Account"
      panelWidth="narrow"
      rootClassName="sm-profile-menu-root"
      renderTrigger={({ open, toggle, panelId }) => (
        <ActionButton
          isQuiet
          aria-label="Account menu"
          aria-expanded={open}
          aria-controls={open ? panelId : undefined}
          UNSAFE_className="sm-profile-menu-trigger"
          onPress={toggle}
        >
          <Avatar
            src={avatarImageSrc(user.avatar_url)}
            alt={user.display_name}
            size="avatar-size-400"
          />
        </ActionButton>
      )}
    >
      <HeaderDropdownMenu>
        <HeaderDropdownItem onPress={openSettings}>
          Profile Settings
        </HeaderDropdownItem>
        <HeaderDropdownItem onPress={() => router.push(`/users/${user.username}`)}>
          View Profile
        </HeaderDropdownItem>
        <HeaderDropdownDivider />
        <HeaderDropdownSection title="Appearance">
          <HeaderDropdownItem onPress={() => setColorScheme("light")}>
            {colorScheme === "light" ? "✓ Light mode" : "Light mode"}
          </HeaderDropdownItem>
          <HeaderDropdownItem onPress={() => setColorScheme("dark")}>
            {colorScheme === "dark" ? "✓ Dark mode" : "Dark mode"}
          </HeaderDropdownItem>
        </HeaderDropdownSection>
        <HeaderDropdownDivider />
        <HeaderDropdownSection>
          <HeaderDropdownItem
            variant="danger"
            onPress={() => {
              clearToken();
              router.push("/login");
            }}
          >
            Log Out
          </HeaderDropdownItem>
        </HeaderDropdownSection>
      </HeaderDropdownMenu>
    </HeaderDropdown>
  );
}
