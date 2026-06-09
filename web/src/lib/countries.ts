export const COUNTRIES: { id: string; name: string }[] = [
  { id: "", name: "Select country" },
  { id: "US", name: "United States" },
  { id: "CA", name: "Canada" },
  { id: "GB", name: "United Kingdom" },
  { id: "AU", name: "Australia" },
  { id: "DE", name: "Germany" },
  { id: "FR", name: "France" },
  { id: "JP", name: "Japan" },
  { id: "KR", name: "South Korea" },
  { id: "MX", name: "Mexico" },
  { id: "BR", name: "Brazil" },
  { id: "IN", name: "India" },
  { id: "NL", name: "Netherlands" },
  { id: "SE", name: "Sweden" },
  { id: "NO", name: "Norway" },
  { id: "NZ", name: "New Zealand" },
  { id: "IE", name: "Ireland" },
  { id: "ES", name: "Spain" },
  { id: "IT", name: "Italy" },
  { id: "PL", name: "Poland" },
  { id: "PH", name: "Philippines" },
  { id: "SG", name: "Singapore" },
];

export function countryName(code: string): string {
  return COUNTRIES.find((c) => c.id === code)?.name ?? code;
}
