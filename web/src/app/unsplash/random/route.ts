import { NextRequest, NextResponse } from "next/server";
import { fetchRandomUnsplashPhoto, DEFAULT_UNSPLASH_QUERY } from "@/lib/unsplash";

export async function GET(request: NextRequest) {
  const q = request.nextUrl.searchParams.get("q")?.trim() || DEFAULT_UNSPLASH_QUERY;
  const photo = await fetchRandomUnsplashPhoto(q);
  return NextResponse.json(
    { photo },
    {
      headers: {
        "Cache-Control": photo
          ? "public, s-maxage=3600, stale-while-revalidate=86400"
          : "public, s-maxage=300, stale-while-revalidate=600",
      },
    },
  );
}
