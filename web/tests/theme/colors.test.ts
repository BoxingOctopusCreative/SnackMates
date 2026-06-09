import { describe, expect, it } from "vitest";
import { brand, darkPalette, lightPalette } from "@/theme/colors";

describe("theme colors", () => {
  it("defines brand tokens", () => {
    expect(brand.toffeeGold).toBe("#EDC21D");
    expect(brand.strawberryRed).toBe("#D4382B");
    expect(brand.cocoaBrown).toBe("#47120E");
  });

  it("uses brand colors in light palette", () => {
    expect(lightPalette.accent).toBe(brand.strawberryRed);
    expect(lightPalette.header).toBe(brand.strawberryRed);
    expect(lightPalette.highlight).toBe(brand.toffeeGold);
  });

  it("uses brand colors in dark palette", () => {
    expect(darkPalette.highlight).toBe(brand.toffeeGold);
    expect(darkPalette.headerAccent).toBe(brand.toffeeGold);
  });
});
