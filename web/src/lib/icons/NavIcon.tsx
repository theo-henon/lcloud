import { cn } from "@/lib/utils";
import { navIcons } from "@/lib/icons/navIcons";

type NavIconProps = {
  to: string;
  active?: boolean;
  className?: string;
};

export function NavIcon({ to, active, className }: NavIconProps) {
  const Icon = navIcons[to];
  if (!Icon) {
    return null;
  }

  return (
    <Icon
      className={cn(
        "h-4 w-4 shrink-0",
        active ? "text-primary" : "text-muted",
        className,
      )}
      aria-hidden
    />
  );
}
