import { formatJalali } from "@/utils/format";

export function JalaliDate({ iso }: { iso: string }) {
  return (
    <span dir="ltr" className="numeric inline-block">
      {formatJalali(iso)}
    </span>
  );
}
