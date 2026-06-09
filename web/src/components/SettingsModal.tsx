"use client";

import { AppModal } from "@/components/AppModal";
import { SettingsPanel } from "@/components/SettingsPanel";

type SettingsModalProps = {
  isOpen: boolean;
  onClose: () => void;
};

export function SettingsModal({ isOpen, onClose }: SettingsModalProps) {
  return (
    <AppModal
      isOpen={isOpen}
      onClose={onClose}
      title="Account Settings"
      titleId="sm-settings-modal-title"
      size="wide"
      align="below-header"
      showCloseButton
      scrollableBody
    >
      <SettingsPanel showHeading={false} showViewProfileLink={false} embedded />
    </AppModal>
  );
}
