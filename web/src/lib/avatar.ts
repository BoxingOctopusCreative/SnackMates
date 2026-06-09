/** Spectrum Avatar requires `src: string` in types, but `null` avoids empty-string img warnings. */
export function avatarImageSrc(url?: string | null): string {
  return (url || null) as string;
}
