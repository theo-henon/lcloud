import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function Badge({
  className,
  children,
}: {
  className?: string;
  children: ReactNode;
}) {
  return (
    <span
      className={cn(
        "rounded-full border border-hairline-strong bg-surface-elevated px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wide text-muted",
        className,
      )}
    >
      {children}
    </span>
  );
}
