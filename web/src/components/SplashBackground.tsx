"use client";

import { Provider, Text, defaultTheme } from "@adobe/react-spectrum";
import type { UnsplashPhoto } from "@/lib/unsplash";
import { SplashSiteFooter } from "@/components/SplashSiteFooter";

export function SplashBackground({
  background,
  children,
}: {
  background: UnsplashPhoto | null;
  children: React.ReactNode;
}) {
  return (
    <Provider theme={defaultTheme} colorScheme="dark" UNSAFE_className="sm-theme">
      <div
        data-theme="dark"
        className="sm-splash-background"
        style={{
          backgroundColor: "var(--sm-bg-subtle)",
          backgroundImage: background
            ? `var(--sm-hero-banner-overlay-image), url("${background.url}")`
            : undefined,
          backgroundSize: "cover",
          backgroundPosition: "center",
          backgroundRepeat: "no-repeat",
        }}
      >
        <div className="sm-splash-background__content">{children}</div>

        <SplashSiteFooter />

        {background && (
          <Text
            UNSAFE_style={{
              position: "absolute",
              right: "0.75rem",
              bottom: "0.75rem",
              zIndex: 1,
              fontSize: "0.75rem",
              color: "var(--sm-text-muted)",
            }}
          >
            Photo by{" "}
            <a
              href={`${background.photographerUrl}?utm_source=snackmates&utm_medium=referral`}
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: "inherit" }}
            >
              {background.photographer}
            </a>{" "}
            on{" "}
            <a
              href={`${background.unsplashUrl}?utm_source=snackmates&utm_medium=referral`}
              target="_blank"
              rel="noopener noreferrer"
              style={{ color: "inherit" }}
            >
              Unsplash
            </a>
          </Text>
        )}
      </div>
    </Provider>
  );
}
