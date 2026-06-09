export const DEFAULT_UNSPLASH_QUERY = "cartoon snack food";

const RANDOM_PHOTO_FALLBACK_QUERIES = ["snack food", "cartoon food", "food"] as const;
const RANDOM_PHOTO_CACHE_TTL_MS = 60 * 60 * 1000;
const RANDOM_PHOTO_FAILURE_CACHE_TTL_MS = 5 * 60 * 1000;

export type UnsplashPhoto = {
  url: string;
  photographer: string;
  photographerUrl: string;
  unsplashUrl: string;
};

type UnsplashApiPhoto = {
  urls: { regular: string };
  links: { html: string; download_location: string };
  user: { name: string; links: { html: string } };
};

type UnsplashSearchResult = {
  results: UnsplashApiPhoto[];
};

const randomPhotoCache = new Map<string, { value: UnsplashPhoto | null; expiresAt: number }>();

export function clearRandomUnsplashPhotoCache(): void {
  randomPhotoCache.clear();
}

function mapUnsplashResult(photo: UnsplashApiPhoto): UnsplashPhoto {
  return {
    url: photo.urls.regular,
    photographer: photo.user.name,
    photographerUrl: photo.user.links.html,
    unsplashUrl: photo.links.html,
  };
}

export async function searchUnsplashPhotos(
  query: string,
  perPage = 12,
): Promise<UnsplashPhoto[]> {
  const accessKey = process.env.UNSPLASH_ACCESS_KEY;
  if (!accessKey || !query.trim()) return [];

  const params = new URLSearchParams({
    query: query.trim(),
    per_page: String(perPage),
    orientation: "landscape",
  });

  const res = await fetch(`https://api.unsplash.com/search/photos?${params}`, {
    headers: { Authorization: `Client-ID ${accessKey}` },
    cache: "no-store",
  });

  if (!res.ok) return [];

  const data = (await res.json()) as UnsplashSearchResult;
  return (data.results ?? []).map(mapUnsplashResult);
}

export async function trackUnsplashDownload(downloadLocation: string): Promise<void> {
  const accessKey = process.env.UNSPLASH_ACCESS_KEY;
  if (!accessKey) return;

  try {
    await fetch(`${downloadLocation}?client_id=${accessKey}`, { cache: "no-store" });
  } catch {
    // Unsplash download tracking is best-effort.
  }
}

export async function fetchRandomUnsplashPhoto(query: string): Promise<UnsplashPhoto | null> {
  const cacheKey = query.trim() || DEFAULT_UNSPLASH_QUERY;
  const cached = randomPhotoCache.get(cacheKey);
  if (cached && cached.expiresAt > Date.now()) {
    return cached.value;
  }

  const photo = await fetchRandomUnsplashPhotoUncached(cacheKey);
  randomPhotoCache.set(cacheKey, {
    value: photo,
    expiresAt: Date.now() + (photo ? RANDOM_PHOTO_CACHE_TTL_MS : RANDOM_PHOTO_FAILURE_CACHE_TTL_MS),
  });
  return photo;
}

function queriesToTry(primaryQuery: string): string[] {
  const trimmed = primaryQuery.trim();
  return [...new Set([trimmed, ...RANDOM_PHOTO_FALLBACK_QUERIES].filter(Boolean))];
}

async function fetchRandomUnsplashPhotoUncached(query: string): Promise<UnsplashPhoto | null> {
  for (const q of queriesToTry(query)) {
    const photo = await fetchRandomPhotoForQuery(q);
    if (photo) return photo;
  }
  return null;
}

async function fetchRandomPhotoForQuery(query: string): Promise<UnsplashPhoto | null> {
  const accessKey = process.env.UNSPLASH_ACCESS_KEY;
  if (!accessKey || !query.trim()) return null;

  const params = new URLSearchParams({
    query: query.trim(),
    orientation: "landscape",
  });

  const res = await fetch(`https://api.unsplash.com/photos/random?${params}`, {
    headers: { Authorization: `Client-ID ${accessKey}` },
    cache: "no-store",
  });

  if (res.status === 404 || !res.ok) return null;

  const photo = (await res.json()) as UnsplashApiPhoto;
  if (!photo?.urls?.regular) return null;

  await trackUnsplashDownload(photo.links.download_location);
  return mapUnsplashResult(photo);
}
