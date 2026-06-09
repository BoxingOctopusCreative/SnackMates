"use client";

import { useEffect, useState } from "react";
import {
  Button,
  Checkbox,
  Flex,
  Form,
  Text,
  TextField,
} from "@adobe/react-spectrum";
import { AppModal } from "@/components/AppModal";
import { api, getToken } from "@/lib/api";

type CreateWishlistModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onCreated: () => void | Promise<void>;
};

export function CreateWishlistModal({ isOpen, onClose, onCreated }: CreateWishlistModalProps) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [isPublic, setIsPublic] = useState(true);
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!isOpen) {
      setTitle("");
      setDescription("");
      setIsPublic(true);
      setError("");
      setSaving(false);
    }
  }, [isOpen]);

  async function createList(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setSaving(true);
    try {
      await api.createWishlist({ title, description, is_public: isPublic }, getToken());
      onClose();
      await onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create wishlist");
    } finally {
      setSaving(false);
    }
  }

  return (
    <AppModal
      isOpen={isOpen}
      onClose={onClose}
      title="Create Wishlist"
      titleId="sm-create-wishlist-modal-title"
      size="narrow"
    >
      <Form maxWidth="100%" onSubmit={createList}>
        <Flex direction="column" gap="size-200">
          <TextField label="Title" value={title} onChange={setTitle} isRequired />
          <TextField label="Description" value={description} onChange={setDescription} />
          <Checkbox isSelected={isPublic} onChange={setIsPublic}>
            Public (required for matching)
          </Checkbox>
          {error && <Text UNSAFE_style={{ color: "var(--sm-error)" }}>{error}</Text>}
          <Button type="submit" variant="accent" isDisabled={saving}>
            Create wishlist
          </Button>
        </Flex>
      </Form>
    </AppModal>
  );
}
