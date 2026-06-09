import { staticAssetUrl } from "@/lib/static-assets";

const LEGAL_MARKDOWN_REVALIDATE_SECONDS = 3600;

export async function fetchLegalMarkdown(filename: string): Promise<string> {
  const url = staticAssetUrl(`legal/${filename}`);
  const response = await fetch(url, {
    next: { revalidate: LEGAL_MARKDOWN_REVALIDATE_SECONDS },
  });

  if (!response.ok) {
    throw new Error(`Failed to load legal markdown from ${url}: ${response.status}`);
  }

  return response.text();
}
