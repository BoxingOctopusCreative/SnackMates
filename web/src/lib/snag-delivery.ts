import type { SnaggedBy } from "@/lib/api";

export type SnagDeliveryMethod = "in_person" | "mail";

export function snagStatusLabel(snaggedBy: SnaggedBy): string {
  const base = `Snagged by ${snaggedBy.display_name}`;
  if (snaggedBy.delivery_method === "mail") {
    if (snaggedBy.tracking_number) {
      return `${base} (delivery by mail - ${snaggedBy.tracking_number})`;
    }
    return `${base} (delivery by mail)`;
  }
  return `${base} (delivery in-person)`;
}
