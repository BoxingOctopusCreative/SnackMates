import { RegisterForm } from "@/app/register/RegisterForm";
import { fetchRandomUnsplashPhoto, DEFAULT_UNSPLASH_QUERY } from "@/lib/unsplash";

export const dynamic = "force-dynamic";

export default async function RegisterPage() {
  const background = await fetchRandomUnsplashPhoto(DEFAULT_UNSPLASH_QUERY);

  return <RegisterForm background={background} />;
}
