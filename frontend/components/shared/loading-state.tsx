import { IconLoader2 } from "@tabler/icons-react";
import { cn } from "@/lib/utils";

type LoadingStateProps = {
  label?: string;
  className?: string;
};

export function LoadingState({
  label = "در حال بارگذاری…",
  className,
}: LoadingStateProps) {
  return (
    <div
      role="status"
      aria-live="polite"
      className={cn(
        "flex items-center justify-center gap-2 py-10 text-muted-foreground",
        className
      )}
    >
      <IconLoader2 className="size-5 animate-spin" stroke={1.6} />
      <span className="text-sm">{label}</span>
    </div>
  );
}
