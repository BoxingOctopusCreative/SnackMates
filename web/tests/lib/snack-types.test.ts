import { describe, expect, it } from "vitest";
import { DEFAULT_SNACK_TYPE, normalizeSnackType, SNACK_TYPES } from "@/lib/snack-types";

describe("snack types", () => {
  it("defines the supported snack categories", () => {
    expect(SNACK_TYPES).toEqual([
      "Candy",
      "Baked Goods",
      "Beverages",
      "Pantry",
      "Chips/Crackers",
    ]);
    expect(DEFAULT_SNACK_TYPE).toBe("Candy");
  });

  it("normalizes unknown types to the default", () => {
    expect(normalizeSnackType("Candy")).toBe("Candy");
    expect(normalizeSnackType("Chocolate")).toBe("Candy");
    expect(normalizeSnackType("")).toBe("Candy");
  });
});
