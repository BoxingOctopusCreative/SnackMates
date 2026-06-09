"use client";

import Image from "next/image";
import { useRef, useState } from "react";
import { Button, Flex, Heading, Text, TextField, View } from "@adobe/react-spectrum";
import { api, ApiError, getToken } from "@/lib/api";
import type { UnsplashPhoto } from "@/lib/unsplash";

type BannerEditorProps = {
  bannerUrl?: string;
  onBannerChange: (url: string | undefined) => void;
  onMessage: (message: string) => void;
  embedded?: boolean;
};

export function BannerEditor({
  bannerUrl,
  onBannerChange,
  onMessage,
  embedded = false,
}: BannerEditorProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [searching, setSearching] = useState(false);
  const [results, setResults] = useState<UnsplashPhoto[]>([]);
  const [searchError, setSearchError] = useState("");
  const [bannerError, setBannerError] = useState("");
  const [selectingUrl, setSelectingUrl] = useState<string | null>(null);

  async function onBannerSelected(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    try {
      const res = await api.uploadBanner(file, getToken());
      onBannerChange(res.banner_url);
      onMessage("Profile banner updated.");
    } catch (err) {
      onMessage(err instanceof Error ? err.message : "Banner upload failed");
    } finally {
      e.target.value = "";
    }
  }

  async function searchUnsplash(e: React.FormEvent) {
    e.preventDefault();
    const q = searchQuery.trim();
    if (!q) return;

    setSearching(true);
    setSearchError("");
    try {
      const res = await fetch(`/api/unsplash/search?q=${encodeURIComponent(q)}`);
      const data = (await res.json()) as { results: UnsplashPhoto[] };
      setResults(data.results ?? []);
      if ((data.results ?? []).length === 0) {
        setSearchError("No photos found. Try a different search.");
      }
    } catch {
      setSearchError("Could not search Unsplash.");
    } finally {
      setSearching(false);
    }
  }

  async function selectUnsplashPhoto(photo: UnsplashPhoto) {
    setBannerError("");
    setSelectingUrl(photo.url);
    try {
      const res = await api.setBannerUrl(photo.url, getToken());
      onBannerChange(res.banner_url ?? photo.url);
      onMessage("Profile banner updated from Unsplash.");
      setResults([]);
    } catch (err) {
      const message =
        err instanceof ApiError
          ? err.message
          : err instanceof Error
            ? err.message
            : "Failed to set banner";
      setBannerError(message);
      onMessage(message);
    } finally {
      setSelectingUrl(null);
    }
  }

  async function removeBanner() {
    try {
      await api.setBannerUrl("", getToken());
      onBannerChange(undefined);
      onMessage("Profile banner removed.");
    } catch (err) {
      onMessage(err instanceof Error ? err.message : "Failed to remove banner");
    }
  }

  return (
    <Flex
      direction="column"
      gap={embedded ? "size-150" : "size-200"}
      UNSAFE_className={embedded ? "sm-settings-modal__banner-editor" : undefined}
    >
      {!embedded && (
        <>
          <Heading level={4} margin={0}>
            Profile Banner
          </Heading>
          <Text>
            Add a wide banner image for your public profile. Upload your own or search Unsplash.
          </Text>
        </>
      )}

      <div
        className={`sm-banner-editor__preview${bannerUrl ? "" : " sm-banner-editor__preview--empty"}${embedded ? " sm-banner-editor__preview--compact" : ""}`}
        style={bannerUrl ? { backgroundImage: `url(${bannerUrl})` } : undefined}
      />

      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        hidden
        onChange={onBannerSelected}
      />
      <Flex gap="size-150" wrap>
        <Button variant="secondary" onPress={() => fileInputRef.current?.click()}>
          Upload banner
        </Button>
        {bannerUrl && (
          <Button variant="secondary" onPress={removeBanner}>
            Remove banner
          </Button>
        )}
      </Flex>

      <form onSubmit={searchUnsplash}>
        <Flex direction="column" gap="size-150">
          <TextField
            label="Search Unsplash"
            value={searchQuery}
            onChange={setSearchQuery}
            width="100%"
          />
          <Button type="submit" variant="accent" isDisabled={searching || !searchQuery.trim()}>
            {searching ? "Searching…" : "Search"}
          </Button>
        </Flex>
      </form>

      {searchError && <Text>{searchError}</Text>}

      {bannerError && (
        <Text UNSAFE_style={{ color: "var(--sm-error)" }}>{bannerError}</Text>
      )}

      {results.length > 0 && (
        <View>
          <Text marginBottom="size-100">Select a photo:</Text>
          <div className="sm-banner-editor__grid">
            {results.map((photo) => (
              <button
                key={photo.url}
                type="button"
                className="sm-banner-editor__thumb"
                disabled={selectingUrl !== null}
                onClick={() => selectUnsplashPhoto(photo)}
                aria-label={`Use photo by ${photo.photographer}`}
                aria-busy={selectingUrl === photo.url}
              >
                <Image
                  src={photo.url}
                  alt=""
                  fill
                  sizes="(max-width: 768px) 33vw, 8rem"
                  className="sm-banner-editor__thumb-image"
                  unoptimized
                />
                {selectingUrl === photo.url && (
                  <span className="sm-banner-editor__thumb-loading">Saving…</span>
                )}
              </button>
            ))}
          </div>
        </View>
      )}
    </Flex>
  );
}
