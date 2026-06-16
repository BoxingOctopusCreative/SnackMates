import { api, clearToken, getToken, saveToken } from "@/lib/api";

/** Restores a client auth token from localStorage or the HTTP-only session cookie. */
export async function restoreSessionFromCookie(): Promise<string | null> {
  const existing = getToken();
  if (existing) {
    try {
      await api.me(existing);
      return existing;
    } catch {
      clearToken();
    }
  }

  try {
    const { token, remember } = await api.session();
    saveToken(token, { remember });
    return token;
  } catch {
    return null;
  }
}
