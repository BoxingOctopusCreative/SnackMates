import { LoginForm } from "@/app/login/LoginForm";
import { fetchRandomUnsplashPhoto, DEFAULT_UNSPLASH_QUERY } from "@/lib/unsplash";

export const dynamic = "force-dynamic";

export default async function LoginPage() {
  const background = await fetchRandomUnsplashPhoto(DEFAULT_UNSPLASH_QUERY);

  return <LoginForm background={background} />;
}
