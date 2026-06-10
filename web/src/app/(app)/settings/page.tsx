"use client";

import { Suspense } from "react";
import { ProgressCircle, View } from "@adobe/react-spectrum";
import { SettingsRedirect } from "./SettingsRedirect";

export default function SettingsPage() {
  return (
    <Suspense
      fallback={
        <View UNSAFE_style={{ display: "grid", placeItems: "center", minHeight: 200 }}>
          <ProgressCircle isIndeterminate aria-label="Opening account settings" />
        </View>
      }
    >
      <SettingsRedirect />
    </Suspense>
  );
}
