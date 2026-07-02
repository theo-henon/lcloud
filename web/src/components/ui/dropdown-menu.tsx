import { useEffect, useRef, useState, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type DropdownMenuProps = {
  trigger: ReactNode;
  children: ReactNode;
  align?: "left" | "right";
  disabled?: boolean;
};

export function DropdownMenu({ trigger, children, align = "left", disabled }: DropdownMenuProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    const onClick = (event: MouseEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", onClick);
    return () => document.removeEventListener("mousedown", onClick);
  }, [open]);

  useEffect(() => {
    if (disabled) {
      setOpen(false);
    }
  }, [disabled]);

  return (
    <div ref={rootRef} className="relative inline-block">
      <div
        onClick={() => {
          if (!disabled) {
            setOpen((value) => !value);
          }
        }}
      >
        {trigger}
      </div>
      {open ? (
        <div
          className={cn(
            "absolute z-50 mt-2 min-w-[180px] rounded-md border border-hairline bg-surface-elevated py-1 shadow-lg",
            align === "right" ? "right-0" : "left-0",
          )}
        >
          <div onClick={() => setOpen(false)}>{children}</div>
        </div>
      ) : null}
    </div>
  );
}

export function DropdownMenuItem({
  children,
  onSelect,
  disabled,
}: {
  children: ReactNode;
  onSelect?: () => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      className="flex w-full items-center px-3 py-2 text-left text-sm text-body hover:bg-surface-card hover:text-ink disabled:opacity-50"
      onClick={() => onSelect?.()}
    >
      {children}
    </button>
  );
}

export function DropdownMenuCheckboxItem({
  checked,
  label,
  onCheckedChange,
  disabled,
}: {
  checked: boolean;
  label: string;
  onCheckedChange: (checked: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-body hover:bg-surface-card hover:text-ink disabled:opacity-50"
      onClick={() => onCheckedChange(!checked)}
    >
      <span className="inline-flex h-4 w-4 items-center justify-center rounded border border-hairline text-xs">
        {checked ? "✓" : ""}
      </span>
      {label}
    </button>
  );
}
