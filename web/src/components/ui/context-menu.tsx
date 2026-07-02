import { useEffect, useRef, useState, type ReactNode } from "react";
import { cn } from "@/lib/utils";

type ContextMenuState = {
  x: number;
  y: number;
} | null;

type ContextMenuProps = {
  children: ReactNode;
  menu: ReactNode;
  disabled?: boolean;
};

export function ContextMenu({ children, menu, disabled }: ContextMenuProps) {
  const [state, setState] = useState<ContextMenuState>(null);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!state) {
      return;
    }
    const close = () => setState(null);
    const onClick = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) {
        close();
      }
    };
    document.addEventListener("mousedown", onClick);
    document.addEventListener("scroll", close, true);
    return () => {
      document.removeEventListener("mousedown", onClick);
      document.removeEventListener("scroll", close, true);
    };
  }, [state]);

  return (
    <>
      <div
        onContextMenu={(event) => {
          if (disabled) {
            return;
          }
          event.preventDefault();
          setState({ x: event.clientX, y: event.clientY });
        }}
      >
        {children}
      </div>
      {state ? (
        <div
          ref={menuRef}
          className={cn(
            "fixed z-50 min-w-[160px] rounded-md border border-hairline bg-surface-elevated py-1 shadow-lg",
          )}
          style={{ top: state.y, left: state.x }}
        >
          <div onClick={() => setState(null)}>{menu}</div>
        </div>
      ) : null}
    </>
  );
}

export function ContextMenuItem({
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
