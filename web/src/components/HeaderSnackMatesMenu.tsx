"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Icon } from "@iconify/react";
import { ActionButton } from "@adobe/react-spectrum";
import { api, getToken } from "@/lib/api";
import { HeaderDropdown } from "@/components/HeaderDropdown";
import { HeaderDropdownItem, HeaderDropdownMenu } from "@/components/HeaderDropdownMenu";

export function HeaderSnackMatesMenu() {
  const router = useRouter();
  const [matching, setMatching] = useState(false);

  async function matchRandomly() {
    setMatching(true);
    try {
      const res = await api.runPairing(getToken());
      const params = new URLSearchParams();
      params.set("random", String(res.paired));
      router.push(`/matches?${params.toString()}`);
    } catch (err) {
      const params = new URLSearchParams();
      params.set("error", err instanceof Error ? err.message : "Matching failed");
      router.push(`/matches?${params.toString()}`);
    } finally {
      setMatching(false);
    }
  }

  return (
    <HeaderDropdown
      title="Snack Mates"
      panelWidth="narrow"
      renderTrigger={({ open, toggle, panelId }) => (
        <ActionButton
          isQuiet
          aria-label="Snack Mates"
          aria-expanded={open}
          aria-controls={open ? panelId : undefined}
          UNSAFE_className="sm-header-icon-link"
          onPress={toggle}
        >
          <Icon icon="ion:people-outline" className="sm-header-icon sm-header-icon--default" aria-hidden />
          <Icon icon="ion:people" className="sm-header-icon sm-header-icon--hover" aria-hidden />
        </ActionButton>
      )}
    >
      <HeaderDropdownMenu>
        <HeaderDropdownItem onPress={() => router.push("/search?type=people")}>
          Add Snack Mate
        </HeaderDropdownItem>
        <HeaderDropdownItem onPress={() => void matchRandomly()} disabled={matching}>
          {matching ? "Matching..." : "Match me Randomly"}
        </HeaderDropdownItem>
      </HeaderDropdownMenu>
    </HeaderDropdown>
  );
}
