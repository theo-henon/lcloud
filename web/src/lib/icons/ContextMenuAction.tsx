import type { ReactNode } from "react";
import { ActionIcon, type ActionIconKey } from "@/lib/icons";
import { cn } from "@/lib/utils";

type ContextMenuActionProps = {
  action: ActionIconKey;
  children: ReactNode;
  destructive?: boolean;
};

export function ContextMenuAction({ action, children, destructive }: ContextMenuActionProps) {
  return (
    <span className={cn("flex items-center", destructive && "text-accent-rose")}>
      <ActionIcon action={action} className="mr-2" destructive={destructive} />
      {children}
    </span>
  );
}
