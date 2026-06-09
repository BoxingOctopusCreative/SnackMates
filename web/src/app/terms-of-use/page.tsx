import type { Metadata } from "next";
import { MarkdownContent } from "@/components/MarkdownContent";
import { PublicDocumentShell } from "@/components/PublicDocumentShell";
import { fetchLegalMarkdown } from "@/lib/markdown";
import { DEFAULT_UNSPLASH_QUERY, fetchRandomUnsplashPhoto } from "@/lib/unsplash";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Terms of Use | SnackMates",
  description: "Terms of Use for SnackMates.",
};

export default async function TermsOfUsePage() {
  const [source, background] = await Promise.all([
    fetchLegalMarkdown("terms-of-use.md"),
    fetchRandomUnsplashPhoto(DEFAULT_UNSPLASH_QUERY),
  ]);

  return (
    <PublicDocumentShell title="Terms of Use" background={background}>
      <article className="sm-prose">
        <MarkdownContent source={source} />
      </article>
    </PublicDocumentShell>
  );
}
