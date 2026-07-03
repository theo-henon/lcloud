import { Link } from "react-router-dom";
import { Card } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type WidgetShellProps = {
  title: string;
  deepLink?: string;
  className?: string;
  children: React.ReactNode;
};

export function WidgetShell({ title, deepLink, className, children }: WidgetShellProps) {
  return (
    <Card className={cn("flex h-full flex-col p-4", className)}>
      <div className="mb-3 flex items-center justify-between gap-2">
        <h2 className="text-sm font-semibold text-ink">{title}</h2>
        {deepLink ? (
          <Link
            to={deepLink}
            className="text-xs font-medium text-primary hover:text-primary-active"
          >
            Open →
          </Link>
        ) : null}
      </div>
      <div className="min-h-0 flex-1">{children}</div>
    </Card>
  );
}

export function WidgetLoading() {
  return <p className="text-sm text-muted">Loading…</p>;
}

export function WidgetError({ message = "Unable to load data." }: { message?: string }) {
  return <p className="text-sm text-accent-rose">{message}</p>;
}

export function WidgetEmpty({ message }: { message: string }) {
  return <p className="text-sm text-muted">{message}</p>;
}
