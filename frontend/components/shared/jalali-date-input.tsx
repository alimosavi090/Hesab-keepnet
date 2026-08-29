"use client";

import { useMemo } from "react";
import { format, getDaysInMonth, parse } from "date-fns-jalali";
import { Label } from "@/components/ui/label";
import { Select as UiSelect, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const MONTHS = [
  "فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
  "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
];

const pad = (n: number) => String(n).padStart(2, "0");

/** Gregorian "YYYY-MM-DD" → Jalali parts (via local noon to avoid TZ drift). */
function toJalaliParts(ymd: string): { y: number; m: number; d: number } | null {
  const date = new Date(`${ymd}T12:00:00`);
  if (Number.isNaN(date.getTime())) return null;
  return {
    y: Number(format(date, "yyyy")),
    m: Number(format(date, "MM")),
    d: Number(format(date, "dd")),
  };
}

/** Jalali parts → Gregorian "YYYY-MM-DD". Clamps the day to the month length. */
function toGregorianYmd(y: number, m: number, d: number): string {
  const ref = parse(`${y}/${pad(m)}/01`, "yyyy/MM/dd", new Date());
  const maxDay = getDaysInMonth(ref);
  const date = parse(`${y}/${pad(m)}/${pad(Math.min(d, maxDay))}`, "yyyy/MM/dd", new Date());
  if (Number.isNaN(date.getTime())) return "";
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
}

type Props = {
  id: string;
  label: string;
  /** Gregorian "YYYY-MM-DD" */
  value: string;
  /** Receives the new Gregorian "YYYY-MM-DD" */
  onChange: (ymd: string) => void;
};

/* Jalali (Persian) date picker: three selects — year, named month, day.
   Keeps the caller working with plain Gregorian "YYYY-MM-DD" strings. */
export function JalaliDateInput({ id, label, value, onChange }: Props) {
  const todayJ = useMemo(
    () => toJalaliParts(new Date().toISOString().slice(0, 10)),
    []
  );
  const parts = toJalaliParts(value) ?? todayJ;

  const years = useMemo(() => {
    const end = todayJ?.y ?? Number(format(new Date(), "yyyy"));
    const list: number[] = [];
    for (let y = end; y >= end - 30; y--) list.push(y);
    return list;
  }, [todayJ?.y]);

  const daysInMonth = useMemo(() => {
    if (!parts) return 31;
    return getDaysInMonth(parse(`${parts.y}/${pad(parts.m)}/01`, "yyyy/MM/dd", new Date()));
  }, [parts?.y, parts?.m]);

  function update(next: { y?: number; m?: number; d?: number }) {
    if (!parts) return;
    const ymd = toGregorianYmd(
      next.y ?? parts.y,
      next.m ?? parts.m,
      next.d ?? parts.d
    );
    if (ymd) onChange(ymd);
  }

  return (
    <div className="space-y-1.5">
      <Label htmlFor={id}>{label}</Label>
      <div className="grid grid-cols-[1fr_1.6fr_0.8fr] gap-2">
        <UiSelect value={String(parts?.d ?? "")} onValueChange={(v) => update({ d: Number(v) })}>
          <SelectTrigger id={id} aria-label="روز"><SelectValue /></SelectTrigger>
          <SelectContent>
            {Array.from({ length: daysInMonth }, (_, i) => i + 1).map((day) => (
              <SelectItem key={day} value={String(day)}>{day.toLocaleString("fa-IR")}</SelectItem>
            ))}
          </SelectContent>
        </UiSelect>

        <UiSelect value={String(parts?.m ?? "")} onValueChange={(v) => update({ m: Number(v) })}>
          <SelectTrigger aria-label="ماه"><SelectValue /></SelectTrigger>
          <SelectContent>
            {MONTHS.map((name, i) => (
              <SelectItem key={name} value={String(i + 1)}>{name}</SelectItem>
            ))}
          </SelectContent>
        </UiSelect>

        <UiSelect value={String(parts?.y ?? "")} onValueChange={(v) => update({ y: Number(v) })}>
          <SelectTrigger aria-label="سال"><SelectValue /></SelectTrigger>
          <SelectContent>
            {years.map((y) => (
              <SelectItem key={y} value={String(y)}>{y.toLocaleString("fa-IR")}</SelectItem>
            ))}
          </SelectContent>
        </UiSelect>
      </div>
    </div>
  );
}
