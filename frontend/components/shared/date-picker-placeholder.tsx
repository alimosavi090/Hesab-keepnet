"use client";

import { IconCalendar } from "@tabler/icons-react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type DatePickerPlaceholderProps = {
  label?: string;
};

export function DatePickerPlaceholder({
  label = "تاریخ",
}: DatePickerPlaceholderProps) {
  const fieldId = `date-picker-${label}`;

  return (
    <div className="w-full max-w-xs space-y-1.5 opacity-70">
      <Label htmlFor={fieldId} className="flex items-center gap-1.5">
        <IconCalendar className="size-3.5" />
        {label}
        <span className="rounded bg-muted px-1 py-0.5 text-[10px] leading-none text-muted-foreground">
          جلالی — به‌زودی
        </span>
      </Label>
      <Input id={fieldId} type="text" disabled placeholder="انتخاب تاریخ" dir="ltr" />
    </div>
  );
}
