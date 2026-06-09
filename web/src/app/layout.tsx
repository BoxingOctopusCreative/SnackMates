import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "@/components/Providers";

export const metadata: Metadata = {
  title: "SnackMates",
  description: "Build snack wishlists and get matched with snack pen pals.",
  icons: {
    icon: "https://assets.snackmates.food/brand/logokit_favicon.png",
    shortcut: "https://assets.snackmates.food/brand/logokit_favicon.png",
    apple: "https://assets.snackmates.food/brand/logokit_favicon.png",
  },
};

const themeInitScript = `
(function () {
  try {
    var stored = localStorage.getItem("sm-color-scheme");
    var prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
    document.documentElement.dataset.theme =
      stored === "dark" || (!stored && prefersDark) ? "dark" : "light";
  } catch (e) {}
})();
`;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="stylesheet" href="https://use.typekit.net/mtj8rsn.css" />
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
