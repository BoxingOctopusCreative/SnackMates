export const SNACK_TYPES = [
  "Candy",
  "Baked Goods",
  "Beverages",
  "Pantry",
  "Chips/Crackers",
] as const;

export type SnackType = (typeof SNACK_TYPES)[number];

export const DEFAULT_SNACK_TYPE: SnackType = "Candy";

export function normalizeSnackType(type: string): SnackType {
  const trimmed = type.trim();
  return (SNACK_TYPES as readonly string[]).includes(trimmed)
    ? (trimmed as SnackType)
    : DEFAULT_SNACK_TYPE;
}
