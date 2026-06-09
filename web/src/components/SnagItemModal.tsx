"use client";

import { useEffect, useState } from "react";
import {
  Button,
  Flex,
  Form,
  Radio,
  RadioGroup,
  TextField,
} from "@adobe/react-spectrum";
import { AppModal } from "@/components/AppModal";
import type { SnagDeliveryMethod } from "@/lib/snag-delivery";

type SnagItemModalProps = {
  isOpen: boolean;
  snackName?: string;
  onClose: () => void;
  onConfirm: (delivery: {
    delivery_method: SnagDeliveryMethod;
    tracking_number?: string;
  }) => Promise<void>;
};

export function SnagItemModal({
  isOpen,
  snackName,
  onClose,
  onConfirm,
}: SnagItemModalProps) {
  const [deliveryMethod, setDeliveryMethod] = useState<SnagDeliveryMethod>("in_person");
  const [trackingNumber, setTrackingNumber] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!isOpen) {
      setDeliveryMethod("in_person");
      setTrackingNumber("");
      setSaving(false);
    }
  }, [isOpen]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await onConfirm({
        delivery_method: deliveryMethod,
        tracking_number:
          deliveryMethod === "mail" && trackingNumber.trim()
            ? trackingNumber.trim()
            : undefined,
      });
      onClose();
    } finally {
      setSaving(false);
    }
  }

  return (
    <AppModal
      isOpen={isOpen}
      onClose={onClose}
      title="Mark Snagged"
      titleId="sm-snag-item-modal-title"
      size="narrow"
    >
      <Form maxWidth="100%" onSubmit={handleSubmit}>
        <Flex direction="column" gap="size-200">
          {snackName && (
            <p className="sm-snag-item-modal__intro">
              How will you get <strong>{snackName}</strong> to your snack mate?
            </p>
          )}
          <RadioGroup
            label="Delivery"
            value={deliveryMethod}
            onChange={(value) => setDeliveryMethod(value as SnagDeliveryMethod)}
            isRequired
          >
            <Radio value="in_person">In person</Radio>
            <Radio value="mail">By mail</Radio>
          </RadioGroup>
          {deliveryMethod === "mail" && (
            <TextField
              label="Tracking number"
              value={trackingNumber}
              onChange={setTrackingNumber}
              description="Optional carrier tracking info for your snack mate."
            />
          )}
          <Button type="submit" variant="accent" isDisabled={saving}>
            {saving ? "Saving..." : "Mark snagged"}
          </Button>
        </Flex>
      </Form>
    </AppModal>
  );
}
