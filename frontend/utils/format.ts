import { format } from "date-fns-jalali";

const faFormatter = new Intl.NumberFormat("fa-IR");

export function formatNumber(value: number): string {
  return faFormatter.format(value);
}

export function formatMoney(amount: number, currency: "RIAL" | "USD"): string {
  const formatted = faFormatter.format(amount);
  return currency === "USD" ? `$ ${formatted}` : `${formatted} ریال`;
}

export function formatSignedMoney(
  amount: number,
  currency: "RIAL" | "USD",
  sign: "positive" | "negative"
): string {
  const prefix = sign === "negative" ? "−" : "+";
  const formatted = faFormatter.format(Math.abs(amount));
  return currency === "USD" ? `${prefix} $${formatted}` : `${prefix} ${formatted} ریال`;
}

export function formatJalali(iso: string): string {
  if (!iso) return "-";
  try {
    return format(new Date(iso), "yyyy/MM/dd");
  } catch {
    return iso.slice(0, 10);
  }
}

export function bpsToPercentLabel(bps: number): string {
  const whole = Math.floor(bps / 100);
  const frac = bps % 100;
  return frac === 0 ? `${whole}٪` : `${whole}.${frac}٪`;
}

export function toISOLocal(date: Date): string {
  const offset = date.getTimezoneOffset();
  return new Date(date.getTime() - offset * 60_000).toISOString().slice(0, 19);
}
