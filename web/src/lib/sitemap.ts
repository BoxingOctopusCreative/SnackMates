export type SitemapEntry = {
  href: string;
  label: string;
  description?: string;
};

export const PUBLIC_SITEMAP: SitemapEntry[] = [
  { href: "/", label: "Home", description: "Landing page for SnackMates." },
  { href: "/register", label: "Register", description: "Create a new SnackMates account." },
  { href: "/login", label: "Sign in", description: "Sign in to your account." },
  { href: "/forgot-password", label: "Forgot password", description: "Request a password reset link." },
  { href: "/reset-password", label: "Reset password", description: "Choose a new password from your reset link." },
  { href: "/verify-email", label: "Verify email", description: "Confirm your email address." },
  { href: "/confirm-account", label: "Confirm account", description: "Finish setting up your account." },
  { href: "/terms-of-use", label: "Terms of use", description: "Rules for using SnackMates." },
  { href: "/privacy-policy", label: "Privacy policy", description: "How we handle your data." },
  { href: "/sitemap", label: "Sitemap", description: "Links to public pages on SnackMates." },
];
