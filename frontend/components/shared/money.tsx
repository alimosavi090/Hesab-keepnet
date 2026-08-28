import { cn } from "@/lib/utils";
import type { Currency } from "@/types/api";
import { formatMoney, formatSignedMoney } from "@/utils/format";

export function Money({
  amount,
  currency,
  sign,
  className,
}: {
  amount: number;
  currency: Currency;
  sign?: "positive" | "negative";
  className?: string;
}) {
  return (
    <span
      dir="ltr"
      className={cn(
        "numeric inline-block",
        sign === "positive" && "text-income",
        sign === "negative" && "text-destructive",
        className
      )}
    >
      {sign ? formatSignedMoney(amount, currency, sign) : formatMoney(amount, currency)}
    </span>
  );
}
