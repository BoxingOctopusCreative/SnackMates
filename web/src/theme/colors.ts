/** SnackMates brand colours — use sparingly; prefer semantic tokens in UI */
export const brand = {
  toffeeGold: "#EDC21D",
  strawberryRed: "#D4382B",
  cocoaBrown: "#47120E",
} as const;

/** Light: bright surfaces, neutral text, strawberry header */
export const lightPalette = {
  background: "#FFFCF5",
  backgroundSubtle: "#FFF6DC",
  surface: "#FFFFFF",
  text: "#1E1A18",
  textMuted: "#5C5650",
  border: "#E4E0D8",
  accent: brand.strawberryRed,
  highlight: brand.toffeeGold,
  header: brand.strawberryRed,
  headerText: "#FFFFFF",
  headerAccent: brand.toffeeGold,
  error: brand.strawberryRed,
} as const;

/** Dark: charcoal surfaces, cream text, gold accents */
export const darkPalette = {
  background: "#0E0E0E",
  backgroundSubtle: "#161616",
  surface: "#1C1C1C",
  surfaceRaised: "#262626",
  text: "#F2F0EB",
  textMuted: "#9A9690",
  border: "#333333",
  accent: "#F06A5E",
  highlight: brand.toffeeGold,
  header: "#121212",
  headerText: "#F2F0EB",
  headerAccent: brand.toffeeGold,
  error: "#F06A5E",
} as const;
