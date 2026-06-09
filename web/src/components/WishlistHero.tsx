"use client";

import { Icon } from "@iconify/react";
import { useState } from "react";
import { Flex, Heading, Text, TextArea, TextField } from "@adobe/react-spectrum";
import { AppModal } from "@/components/AppModal";
import { api, getToken } from "@/lib/api";
import { WishlistBannerEditor } from "@/components/WishlistBannerEditor";
import { PageHero } from "@/components/PageHero";

type WishlistHeroProps = {
  wishlistSlug: string;
  title: string;
  description?: string;
  isPublic: boolean;
  bannerUrl?: string;
  isOwner?: boolean;
  onAddSnack?: () => void;
  onBannerChange: (url: string | undefined) => void;
  onDetailsChange: (patch: { title?: string; description?: string; slug?: string }) => void;
};

function HeroIconButton({
  label,
  onPress,
  className,
  outlineIcon,
  filledIcon,
}: {
  label: string;
  onPress: () => void;
  className?: string;
  outlineIcon: string;
  filledIcon: string;
}) {
  const [hovered, setHovered] = useState(false);

  return (
    <button
      type="button"
      className={["sm-page-hero__edit-button", className].filter(Boolean).join(" ")}
      aria-label={label}
      title={label}
      onClick={onPress}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      onFocus={() => setHovered(true)}
      onBlur={() => setHovered(false)}
    >
      <Icon
        icon={hovered ? filledIcon : outlineIcon}
        className="sm-page-hero__edit-icon"
        aria-hidden
      />
    </button>
  );
}

export function WishlistHero({
  wishlistSlug,
  title,
  description,
  isPublic,
  bannerUrl,
  isOwner,
  onAddSnack,
  onBannerChange,
  onDetailsChange,
}: WishlistHeroProps) {
  const [bannerEditorOpen, setBannerEditorOpen] = useState(false);
  const [titleEditing, setTitleEditing] = useState(false);
  const [descriptionEditing, setDescriptionEditing] = useState(false);
  const [bannerMessage, setBannerMessage] = useState("");
  const [detailsMessage, setDetailsMessage] = useState("");
  const [editTitle, setEditTitle] = useState(title);
  const [editDescription, setEditDescription] = useState(description ?? "");
  const [savingDetails, setSavingDetails] = useState(false);

  async function saveDetails(nextTitle: string, nextDescription: string) {
    setSavingDetails(true);
    setDetailsMessage("");
    try {
      const res = await api.updateWishlist(
        wishlistSlug,
        {
          title: nextTitle.trim(),
          description: nextDescription.trim(),
          is_public: isPublic,
        },
        getToken(),
      );
      onDetailsChange({
        title: nextTitle.trim(),
        description: nextDescription.trim(),
        slug: res.slug,
      });
      setDetailsMessage("Wishlist updated.");
    } catch (err) {
      setDetailsMessage(err instanceof Error ? err.message : "Failed to update wishlist");
    } finally {
      setSavingDetails(false);
    }
  }

  async function commitTitleEdit() {
    const nextTitle = editTitle.trim();
    if (!nextTitle) {
      setEditTitle(title);
      setTitleEditing(false);
      return;
    }

    if (nextTitle === title) {
      setTitleEditing(false);
      return;
    }

    await saveDetails(nextTitle, description ?? "");
    setTitleEditing(false);
  }

  async function commitDescriptionEdit() {
    const nextDescription = editDescription.trim();
    if (nextDescription === (description ?? "")) {
      setDescriptionEditing(false);
      return;
    }

    await saveDetails(title, nextDescription);
    setDescriptionEditing(false);
  }

  function cancelTitleEdit() {
    setEditTitle(title);
    setTitleEditing(false);
  }

  function cancelDescriptionEdit() {
    setEditDescription(description ?? "");
    setDescriptionEditing(false);
  }

  return (
    <>
      <PageHero
        title={title}
        ariaLabel={`${title} wishlist`}
        bannerUrl={bannerUrl}
        bannerOverlay={
          isOwner ? (
            <HeroIconButton
              label="Edit banner"
              className="sm-page-hero__edit-button--banner"
              outlineIcon="ion:pencil-outline"
              filledIcon="ion:pencil"
              onPress={() => setBannerEditorOpen(true)}
            />
          ) : undefined
        }
      >
        <Flex direction="column" gap="size-150">
          <Flex direction="column" gap="size-75">
            <Flex
              alignItems="center"
              justifyContent="space-between"
              gap="size-100"
              wrap
              UNSAFE_className="sm-page-hero__title-row"
            >
              <Flex alignItems="center" gap="size-100" wrap UNSAFE_className="sm-page-hero__title-row-main">
                {titleEditing ? (
                  <TextField
                    label="Title"
                    value={editTitle}
                    onChange={setEditTitle}
                    onBlur={() => {
                      void commitTitleEdit();
                    }}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") {
                        e.preventDefault();
                        void commitTitleEdit();
                      } else if (e.key === "Escape") {
                        cancelTitleEdit();
                      }
                    }}
                    isDisabled={savingDetails}
                    autoFocus
                    width="100%"
                    UNSAFE_className="sm-page-hero__inline-field sm-page-hero__inline-field--title"
                  />
                ) : (
                  <>
                    <Heading level={1} margin={0}>
                      {title}
                    </Heading>
                    {isOwner && (
                      <HeroIconButton
                        label="Edit title"
                        outlineIcon="ion:pencil-outline"
                        filledIcon="ion:pencil"
                        onPress={() => {
                          setEditTitle(title);
                          setTitleEditing(true);
                        }}
                      />
                    )}
                  </>
                )}
              </Flex>
              {isOwner && onAddSnack && !titleEditing && (
                <HeroIconButton
                  label="Add Snack"
                  className="sm-page-hero__add-button"
                  outlineIcon="ion:add-outline"
                  filledIcon="ion:add"
                  onPress={onAddSnack}
                />
              )}
            </Flex>
            {(description || isOwner) && (
              <Flex alignItems="start" gap="size-100" wrap UNSAFE_className="sm-page-hero__description-row">
                {descriptionEditing ? (
                  <TextArea
                    label="Description"
                    value={editDescription}
                    onChange={setEditDescription}
                    onBlur={() => {
                      void commitDescriptionEdit();
                    }}
                    onKeyDown={(e) => {
                      if (e.key === "Escape") {
                        cancelDescriptionEdit();
                      }
                    }}
                    isDisabled={savingDetails}
                    autoFocus
                    width="100%"
                    UNSAFE_className="sm-page-hero__inline-field"
                  />
                ) : (
                  <>
                    {description ? (
                      <Text>{description}</Text>
                    ) : (
                      <Text UNSAFE_style={{ color: "var(--sm-text-muted)" }}>Add a description</Text>
                    )}
                    {isOwner && (
                      <HeroIconButton
                        label="Edit description"
                        outlineIcon="ion:pencil-outline"
                        filledIcon="ion:pencil"
                        onPress={() => {
                          setEditDescription(description ?? "");
                          setDescriptionEditing(true);
                        }}
                      />
                    )}
                  </>
                )}
              </Flex>
            )}
          </Flex>
          {(bannerMessage || detailsMessage) && (
            <Text UNSAFE_style={{ color: "var(--sm-highlight)" }}>
              {bannerMessage || detailsMessage}
            </Text>
          )}
        </Flex>
      </PageHero>

      {isOwner && (
        <AppModal
          isOpen={bannerEditorOpen}
          onClose={() => setBannerEditorOpen(false)}
          title="Wishlist Banner"
          titleId="sm-wishlist-banner-modal-title"
          size="medium"
        >
          <WishlistBannerEditor
            wishlistSlug={wishlistSlug}
            bannerUrl={bannerUrl}
            onBannerChange={(url) => {
              onBannerChange(url);
              setBannerEditorOpen(false);
            }}
            onMessage={setBannerMessage}
          />
        </AppModal>
      )}
    </>
  );
}
