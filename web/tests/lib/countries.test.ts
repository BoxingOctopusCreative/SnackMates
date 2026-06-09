import { describe, expect, it } from "vitest";
import { COUNTRIES, countryName } from "@/lib/countries";

describe("countries", () => {
  it("includes a blank placeholder option", () => {
    expect(COUNTRIES[0]).toEqual({ id: "", name: "Select country" });
  });

  it("resolves known country codes", () => {
    expect(countryName("JP")).toBe("Japan");
    expect(countryName("US")).toBe("United States");
  });

  it("returns the code when unknown", () => {
    expect(countryName("ZZ")).toBe("ZZ");
  });
});
