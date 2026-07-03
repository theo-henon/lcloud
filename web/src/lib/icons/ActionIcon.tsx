import { cn } from "@/lib/utils";
import { actionIcons, isDestructiveAction, type ActionIconKey } from "@/lib/icons/actionIcons";

type ActionIconProps = {
  action: ActionIconKey;
  className?: string;
  destructive?: boolean;
};

export function ActionIcon({ action, className, destructive }: ActionIconProps) {
  const Icon = actionIcons[action];
  const isDestructive = destructive ?? isDestructiveAction(action);

  return (
    <Icon
      className={cn(
        "h-4 w-4 shrink-0",
        isDestructive && "text-accent-rose",
        className,
      )}
      aria-hidden
    />
  );
}
