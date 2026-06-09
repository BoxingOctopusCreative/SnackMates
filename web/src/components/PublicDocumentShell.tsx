import Link from "next/link";
import type { UnsplashPhoto } from "@/lib/unsplash";

export function PublicDocumentShell({
  title,
  background,
  children,
}: {
  title: string;
  background: UnsplashPhoto | null;
  children: React.ReactNode;
}) {
  return (
    <main
      className="sm-public-document-page"
      style={{
        backgroundImage: background
          ? `var(--sm-hero-banner-overlay-image), url("${background.url}")`
          : undefined,
      }}
    >
      <div className="sm-public-document-page__inner">
        <div className="sm-public-document-page__panel">
          <Link href="/" className="sm-public-document-page__back">
            ← Back to home
          </Link>
          <h1 className="sm-hero-title">{title}</h1>
          {children}
        </div>
      </div>

      {background && (
        <p className="sm-page-hero__attribution sm-public-document-page__attribution">
          Photo by{" "}
          <a
            href={`${background.photographerUrl}?utm_source=snackmates&utm_medium=referral`}
            target="_blank"
            rel="noopener noreferrer"
          >
            {background.photographer}
          </a>{" "}
          on{" "}
          <a
            href={`${background.unsplashUrl}?utm_source=snackmates&utm_medium=referral`}
            target="_blank"
            rel="noopener noreferrer"
          >
            Unsplash
          </a>
        </p>
      )}
    </main>
  );
}
