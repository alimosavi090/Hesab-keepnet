import { IconTool } from "@tabler/icons-react";
import { EmptyState } from "@/components/shared/empty-state";
import { Badge } from "@/components/ui/badge";

type ComingSoonProps = {
  title: string;
  description?: string;
};

export function ComingSoon({ title, description }: ComingSoonProps) {
  return (
    <div className="space-y-6">
      <PageHeader
        title={title}
        actions={
          <Badge variant="outline" className="text-warning border-warning/40">
            در انتظار فازهای بعدی
          </Badge>
        }
      />
      <EmptyState
        icon={IconTool}
        title="این بخش به‌زودی فعال می‌شود"
        description={
          description ??
          "پیاده‌سازی این بخش طبق نقشه راه پروژه در فازهای بعدی انجام می‌شود."
        }
      />
    </div>
  );
}

function PageHeader({
  title,
  actions,
}: {
  title: string;
  actions?: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-3">
      <h2 className="text-lg font-semibold">{title}</h2>
      {actions}
    </div>
  );
}
