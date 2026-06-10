import { readFile } from "fs/promises";
import path from "path";

const LEGAL_MARKDOWN_DIR = path.join(process.cwd(), "content", "legal");

function legalMarkdownPath(filename: string): string {
  const base = path.basename(filename);
  if (base !== filename || !base.endsWith(".md")) {
    throw new Error(`Invalid legal markdown filename: ${filename}`);
  }
  return path.join(LEGAL_MARKDOWN_DIR, base);
}

export async function fetchLegalMarkdown(filename: string): Promise<string> {
  const filePath = legalMarkdownPath(filename);
  try {
    return await readFile(filePath, "utf8");
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to load legal markdown from ${filePath}: ${detail}`);
  }
}
