import { HomeSplash } from "@/components/HomeSplash";
import { fetchRandomUnsplashPhoto, DEFAULT_UNSPLASH_QUERY } from "@/lib/unsplash";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const background = await fetchRandomUnsplashPhoto(DEFAULT_UNSPLASH_QUERY);

  return <HomeSplash background={background} />;
}
