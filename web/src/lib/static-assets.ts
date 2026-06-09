const DEFAULT_STATIC_ASSETS_BASE_URL = "https://assets.snackmates.food";

export function getStaticAssetsBaseUrl(): string {
  const configured = process.env.STATIC_ASSETS_BASE_URL?.trim();
  const baseUrl = configured || DEFAULT_STATIC_ASSETS_BASE_URL;
  return baseUrl.replace(/\/$/, "");
}

export function staticAssetUrl(path: string): string {
  const normalizedPath = path.replace(/^\/+/, "");
  return `${getStaticAssetsBaseUrl()}/${normalizedPath}`;
}
