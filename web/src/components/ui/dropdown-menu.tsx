import { useEffect, useLayoutEffect, useRef, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";

type DropdownMenuProps = {
  trigger: ReactNode;
  children: ReactNode;
  align?: "left" | "right";
  disabled?: boolean;
};

const MENU_GAP_PX = 8;

export function DropdownMenu({ trigger, children, align = "left", disabled }: DropdownMenuProps) {
  const [open, setOpen] = useState(false);
  const [menuStyle, setMenuStyle] = useState<React.CSSProperties>({ visibility: "hidden" });
  const rootRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useLayoutEffect(() => {
    if (!open) {
      return;
    }

    const triggerEl = rootRef.current;
    const menuEl = menuRef.current;
    if (!triggerEl || !menuEl) {
      return;
    }

    const triggerRect = triggerEl.getBoundingClientRect();
    const menuRect = menuEl.getBoundingClientRect();
    const spaceBelow = window.innerHeight - triggerRect.bottom;
    const openAbove =
      spaceBelow < menuRect.height + MENU_GAP_PX &&
      triggerRect.top > menuRect.height + MENU_GAP_PX;

    let top = openAbove
      ? triggerRect.top - menuRect.height - MENU_GAP_PX
      : triggerRect.bottom + MENU_GAP_PX;
    let left = align === "right" ? triggerRect.right - menuRect.width : triggerRect.left;

    const maxLeft = window.innerWidth - menuRect.width - MENU_GAP_PX;
    left = Math.max(MENU_GAP_PX, Math.min(left, maxLeft));
    top = Math.max(MENU_GAP_PX, top);

    setMenuStyle({
      position: "fixed",
      top,
      left,
      zIndex: 50,
      visibility: "visible",
    });
  }, [open, align]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const onClick = (event: MouseEvent) => {
      const target = event.target as Node;
      if (!rootRef.current?.contains(target) && !menuRef.current?.contains(target)) {
        setOpen(false);
      }
    };
    const onScroll = () => setOpen(false);
    document.addEventListener("mousedown", onClick);
    window.addEventListener("scroll", onScroll, true);
    return () => {
      document.removeEventListener("mousedown", onClick);
      window.removeEventListener("scroll", onScroll, true);
    };
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
      {open
        ? createPortal(
            <div
              ref={menuRef}
              style={menuStyle}
              className="min-w-[180px] rounded-md border border-hairline bg-surface-elevated py-1 shadow-lg"
            >
              <div onClick={() => setOpen(false)}>{children}</div>
            </div>,
            document.body,
          )
        : null}
    </div>
  );
}

export function DropdownMenuItem({
  children,
  onSelect,
  disabled,
  title,
}: {
  children: ReactNode;
  onSelect?: () => void;
  disabled?: boolean;
  title?: string;
}) {
  return (
    <button
      type="button"
      disabled={disabled}
      title={title}
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
