import type { Metadata } from "next";
import { MarkdownContent } from "@/components/MarkdownContent";
import { PublicDocumentShell } from "@/components/PublicDocumentShell";
import { fetchLegalMarkdown } from "@/lib/markdown";
import { DEFAULT_UNSPLASH_QUERY, fetchRandomUnsplashPhoto } from "@/lib/unsplash";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "Privacy Policy | SnackMates",
  description: "Privacy Policy for SnackMates.",
};

export default async function PrivacyPolicyPage() {
  const [source, background] = await Promise.all([
    fetchLegalMarkdown("privacy-policy.md"),
    fetchRandomUnsplashPhoto(DEFAULT_UNSPLASH_QUERY),
  ]);

  return (
    <PublicDocumentShell title="Privacy Policy" background={background}>
      <article className="sm-prose">
        <MarkdownContent source={source} />
      </article>
    </PublicDocumentShell>
  );
}
