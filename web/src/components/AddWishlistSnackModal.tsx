"use client";

import { useEffect, useState } from "react";
import {
  Button,
  Flex,
  Form,
  Item,
  Picker,
  TextField,
} from "@adobe/react-spectrum";
import { AppModal } from "@/components/AppModal";
import { api, getToken } from "@/lib/api";
import { DEFAULT_SNACK_TYPE, normalizeSnackType, SNACK_TYPES } from "@/lib/snack-types";

type AddWishlistSnackModalProps = {
  isOpen: boolean;
  onClose: () => void;
  wishlistSlug: string;
  onAdded: () => void | Promise<void>;
};

export function AddWishlistSnackModal({
  isOpen,
  onClose,
  wishlistSlug,
  onAdded,
}: AddWishlistSnackModalProps) {
  const [name, setName] = useState("");
  const [snackType, setSnackType] = useState(DEFAULT_SNACK_TYPE);
  const [brand, setBrand] = useState("");
  const [notes, setNotes] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!isOpen) {
      setName("");
      setSnackType(DEFAULT_SNACK_TYPE);
      setBrand("");
      setNotes("");
      setSaving(false);
    }
  }, [isOpen]);

  async function addItem(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await api.addItem(wishlistSlug, { name, type: snackType, brand, notes }, getToken());
      onClose();
      await onAdded();
    } finally {
      setSaving(false);
    }
  }

  return (
    <AppModal
      isOpen={isOpen}
      onClose={onClose}
      title="Add Snack"
      titleId="sm-add-snack-modal-title"
      size="narrow"
    >
      <Form maxWidth="100%" onSubmit={addItem}>
        <Flex direction="column" gap="size-200">
          <TextField label="Snack Name" value={name} onChange={setName} isRequired />
          <Picker
            label="Type"
            selectedKey={snackType}
            onSelectionChange={(key) => setSnackType(normalizeSnackType(String(key)))}
            isRequired
          >
            {SNACK_TYPES.map((type) => (
              <Item key={type}>{type}</Item>
            ))}
          </Picker>
          <TextField label="Brand" value={brand} onChange={setBrand} />
          <TextField label="Notes" value={notes} onChange={setNotes} />
          <Button type="submit" variant="accent" isDisabled={saving}>
            Add to wishlist
          </Button>
        </Flex>
      </Form>
    </AppModal>
  );
}
