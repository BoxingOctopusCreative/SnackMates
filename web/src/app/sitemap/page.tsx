import type { Metadata } from "next";
import Link from "next/link";
import { PublicDocumentShell } from "@/components/PublicDocumentShell";
import { PUBLIC_SITEMAP } from "@/lib/sitemap";
import { DEFAULT_UNSPLASH_QUERY, fetchRandomUnsplashPhoto } from "@/lib/unsplash";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Sitemap | SnackMates",
  description: "Public pages available on SnackMates.",
};

export default async function SitemapPage() {
  const background = await fetchRandomUnsplashPhoto(DEFAULT_UNSPLASH_QUERY);

  return (
    <PublicDocumentShell title="Sitemap" background={background}>
      <p className="mb-6 text-(--sm-text-muted)">
        Public pages you can visit without signing in.
      </p>
      <ul className="sm-sitemap-list">
        {PUBLIC_SITEMAP.map((entry) => (
          <li key={entry.href}>
            <Link href={entry.href}>{entry.label}</Link>
            {entry.description ? (
              <span className="block text-sm text-(--sm-text-muted)">{entry.description}</span>
            ) : null}
          </li>
        ))}
      </ul>
    </PublicDocumentShell>
  );
}
