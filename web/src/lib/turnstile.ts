export function turnstileSiteKey(): string {
  return process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY?.trim() ?? "";
}

export function turnstileEnabled(): boolean {
  return turnstileSiteKey() !== "";
}
