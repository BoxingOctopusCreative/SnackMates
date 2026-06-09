"use client";

import { useEffect, useState, type ReactNode } from "react";
import { Flex, Heading, Text } from "@adobe/react-spectrum";
import { DEFAULT_UNSPLASH_QUERY, type UnsplashPhoto } from "@/lib/unsplash";

export function usePageHeroBanner(bannerUrl?: string) {
  const [defaultPhoto, setDefaultPhoto] = useState<UnsplashPhoto | null>(null);

  useEffect(() => {
    if (bannerUrl) return;

    let cancelled = false;
    fetch(`/api/unsplash/random?q=${encodeURIComponent(DEFAULT_UNSPLASH_QUERY)}`)
      .then((res) => res.json())
      .then((data: { photo: UnsplashPhoto | null }) => {
        if (!cancelled) setDefaultPhoto(data.photo ?? null);
      })
      .catch(() => {
        if (!cancelled) setDefaultPhoto(null);
      });

    return () => {
      cancelled = true;
    };
  }, [bannerUrl]);

  const displayUrl = bannerUrl ?? defaultPhoto?.url;
  const hasBanner = Boolean(displayUrl);
  const attribution = bannerUrl ? null : defaultPhoto;

  return { displayUrl, hasBanner, attribution };
}

type PageHeroBannerProps = {
  title: string;
  bannerUrl?: string;
  overlay?: ReactNode;
};

export function PageHeroBanner({ title, bannerUrl, overlay }: PageHeroBannerProps) {
  const { displayUrl, hasBanner, attribution } = usePageHeroBanner(bannerUrl);

  return (
    <div className="sm-page-hero__banner-wrap">
      <div
        className={`sm-page-hero__banner${hasBanner ? "" : " sm-page-hero__banner--default"}`}
        style={hasBanner ? { backgroundImage: `url(${displayUrl})` } : undefined}
        role={hasBanner ? "img" : undefined}
        aria-label={hasBanner ? `${title} banner` : undefined}
      />
      {(overlay || attribution) && (
        <div className="sm-page-hero__banner-overlay">
          {overlay}
          {attribution && (
            <p className="sm-page-hero__attribution">
              Photo by{" "}
              <a
                href={`${attribution.photographerUrl}?utm_source=snackmates&utm_medium=referral`}
                target="_blank"
                rel="noopener noreferrer"
              >
                {attribution.photographer}
              </a>{" "}
              on{" "}
              <a
                href={`${attribution.unsplashUrl}?utm_source=snackmates&utm_medium=referral`}
                target="_blank"
                rel="noopener noreferrer"
              >
                Unsplash
              </a>
            </p>
          )}
        </div>
      )}
    </div>
  );
}

type PageHeroProps = {
  title: string;
  description?: string;
  bannerUrl?: string;
  bannerOverlay?: ReactNode;
  ariaLabel?: string;
  children?: ReactNode;
};

export function PageHero({
  title,
  description,
  bannerUrl,
  bannerOverlay,
  ariaLabel,
  children,
}: PageHeroProps) {
  return (
    <section className="sm-page-hero" aria-label={ariaLabel ?? title}>
      <PageHeroBanner title={title} bannerUrl={bannerUrl} overlay={bannerOverlay} />
      <div className="sm-page-hero__body">
        {children ?? (
          <Flex direction="column" gap="size-75">
            <Heading level={1} margin={0}>
              {title}
            </Heading>
            {description && <Text>{description}</Text>}
          </Flex>
        )}
      </div>
    </section>
  );
}
