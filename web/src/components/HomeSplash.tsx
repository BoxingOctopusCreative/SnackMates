"use client";

import Link from "next/link";
import { Button, Flex, Text } from "@adobe/react-spectrum";
import type { UnsplashPhoto } from "@/lib/unsplash";
import { SplashBackground } from "@/components/SplashBackground";
import { SplashLogo } from "@/components/SplashLogo";

export function HomeSplash({ background }: { background: UnsplashPhoto | null }) {
  return (
    <SplashBackground background={background}>
      <Flex direction="column" alignItems="center" gap="size-300">
        <SplashLogo />
        <Text
          UNSAFE_style={{
            color: "var(--sm-text-muted)",
            fontSize: "1.25rem",
            maxWidth: 560,
            textAlign: "center",
          }}
        >
          Curate the snacks you crave, connect with snack mates, and send each other care packages
          like pen pals, but tastier.
        </Text>
        <Flex gap="size-200">
          <Link href="/register">
            <Button variant="accent" style="fill">
              Get started
            </Button>
          </Link>
          <Link href="/login">
            <Button variant="secondary" style="outline">
              Sign in
            </Button>
          </Link>
        </Flex>
        <ul className="sm-splash-footer-links" aria-label="Legal and site links">
          <li>
            <Link href="/terms-of-use">Terms of use</Link>
          </li>
          <li className="sm-splash-footer-links__separator" aria-hidden>
            ·
          </li>
          <li>
            <Link href="/privacy-policy">Privacy policy</Link>
          </li>
          <li className="sm-splash-footer-links__separator" aria-hidden>
            ·
          </li>
          <li>
            <Link href="/sitemap">Sitemap</Link>
          </li>
        </ul>
      </Flex>
    </SplashBackground>
  );
}
