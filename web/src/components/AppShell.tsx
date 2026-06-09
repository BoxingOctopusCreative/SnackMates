"use client";

import Image from "next/image";
import Link from "next/link";
import { View } from "@adobe/react-spectrum";
import { User } from "@/lib/api";
import { HeaderIconLink } from "@/components/HeaderIconLink";
import { HeaderSnackMatesMenu } from "@/components/HeaderSnackMatesMenu";
import {
  HeaderSearchInline,
  HeaderSearchPanel,
  HeaderSearchProvider,
  HeaderSearchToggle,
} from "@/components/HeaderSearch";
import { SettingsModalProvider } from "@/components/SettingsModalProvider";
import { MessagesProvider } from "@/components/MessagesProvider";
import { MessagesIconLink } from "@/components/MessagesIconLink";
import { ChatProvider } from "@/components/ChatProvider";
import { ChatWidget } from "@/components/ChatWidget";
import { NotificationMenu } from "@/components/NotificationMenu";
import { UserMenu } from "@/components/UserMenu";

const LOGO_URL = "https://assets.snackmates.food/brand/logokit_wide.png";

const contentLayout = {
  width: "100%",
  maxWidth: "var(--sm-content-max-width)",
  marginLeft: "auto",
  marginRight: "auto",
  paddingLeft: "var(--spectrum-global-dimension-size-200, 16px)",
  paddingRight: "var(--spectrum-global-dimension-size-200, 16px)",
} as const;

const mainContentLayout = {
  ...contentLayout,
  maxWidth: "var(--sm-main-content-max-width)",
} as const;

export function AppShell({
  children,
  user,
}: {
  children: React.ReactNode;
  user?: User;
}) {
  return (
    <SettingsModalProvider>
    <MessagesProvider>
    <ChatProvider>
    <View
      minHeight="100vh"
      UNSAFE_style={{ backgroundColor: "var(--sm-bg)" }}
    >
      <View
        paddingY="size-200"
        UNSAFE_className="sm-app-header"
        UNSAFE_style={{
          position: "sticky",
          top: 0,
          zIndex: 100,
          backgroundColor: "var(--sm-header-bg)",
          borderBottom: "3px solid var(--sm-cocoa-brown)",
        }}
      >
        <HeaderSearchProvider>
          <div className="sm-header-inner" style={contentLayout}>
            <div className="sm-header-row">
              <Link href="/dashboard" className="sm-header-logo">
                <Image
                  src={LOGO_URL}
                  alt="SnackMates"
                  width={308}
                  height={73}
                  priority
                />
              </Link>
              {user && (
                <>
                  <HeaderSearchInline />
                  <div className="sm-header-actions">
                    <HeaderSearchToggle />
                    <HeaderIconLink
                      href="/wishlists"
                      label="Wishlist"
                      outlineIcon="ion:heart-outline"
                      filledIcon="ion:heart"
                    />
                    <MessagesIconLink />
                    <HeaderSnackMatesMenu />
                    <NotificationMenu />
                    <UserMenu user={user} />
                  </div>
                </>
              )}
            </div>
            {user && <HeaderSearchPanel />}
          </div>
        </HeaderSearchProvider>
      </View>

      <View marginTop="size-200" padding="size-200" UNSAFE_style={mainContentLayout}>
        {children}
      </View>
      <ChatWidget />
    </View>
    </ChatProvider>
    </MessagesProvider>
    </SettingsModalProvider>
  );
}
