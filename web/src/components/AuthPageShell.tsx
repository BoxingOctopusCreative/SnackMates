"use client";

import { Flex, View } from "@adobe/react-spectrum";
import { useLayoutEffect } from "react";
import type { UnsplashPhoto } from "@/lib/unsplash";
import { SplashBackground } from "@/components/SplashBackground";
import { SplashLogo } from "@/components/SplashLogo";

function useSplashDarkMode() {
  useLayoutEffect(() => {
    const root = document.documentElement;
    const previousTheme = root.dataset.theme;
    const previousColorScheme = root.style.colorScheme;

    root.dataset.theme = "dark";
    root.style.colorScheme = "dark";

    return () => {
      root.dataset.theme = previousTheme ?? "";
      root.style.colorScheme = previousColorScheme ?? "";
    };
  }, []);
}

export function AuthPageShell({
  background,
  children,
}: {
  background: UnsplashPhoto | null;
  children: React.ReactNode;
}) {
  useSplashDarkMode();

  return (
    <SplashBackground background={background}>
      <Flex direction="column" alignItems="center" gap="size-300" width="100%">
        <SplashLogo />
        <View
          width="100%"
          maxWidth="size-3600"
          backgroundColor="gray-50"
          padding="size-400"
          borderRadius="large"
          borderWidth="thin"
          borderColor="gray-300"
          UNSAFE_className="sm-auth-card"
        >
          {children}
        </View>
      </Flex>
    </SplashBackground>
  );
}
