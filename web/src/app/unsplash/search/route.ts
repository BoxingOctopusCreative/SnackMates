import { NextRequest, NextResponse } from "next/server";
import { searchUnsplashPhotos } from "@/lib/unsplash";

export async function GET(request: NextRequest) {
  const q = request.nextUrl.searchParams.get("q")?.trim() ?? "";
  if (!q) {
    return NextResponse.json({ results: [] });
  }

  const results = await searchUnsplashPhotos(q, 12);
  return NextResponse.json({ results });
}
