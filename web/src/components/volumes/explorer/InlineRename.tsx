import { useEffect, useRef, useState } from "react";
import { Input } from "@/components/ui/input";

type InlineRenameProps = {
  value: string;
  active: boolean;
  onCommit: (next: string) => void;
  onCancel: () => void;
  children: React.ReactNode;
  onRequestRename: () => void;
};

export function InlineRename({
  value,
  active,
  onCommit,
  onCancel,
  children,
  onRequestRename,
}: InlineRenameProps) {
  const [draft, setDraft] = useState(value);
  const inputRef = useRef<HTMLInputElement>(null);
  const clickTimer = useRef<number | null>(null);

  useEffect(() => {
    if (active) {
      setDraft(value);
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [active, value]);

  if (active) {
    return (
      <Input
        ref={inputRef}
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            onCommit(draft.trim());
          }
          if (event.key === "Escape") {
            onCancel();
          }
        }}
        onBlur={() => onCommit(draft.trim())}
        className="h-8 py-1"
      />
    );
  }

  return (
    <span
      role="button"
      tabIndex={0}
      onDoubleClick={() => {
        if (clickTimer.current) {
          window.clearTimeout(clickTimer.current);
          clickTimer.current = null;
          onRequestRename();
        }
      }}
      onClick={() => {
        if (clickTimer.current) {
          window.clearTimeout(clickTimer.current);
        }
        clickTimer.current = window.setTimeout(() => {
          clickTimer.current = null;
        }, 300);
      }}
      onKeyDown={(event) => {
        if (event.key === "F2") {
          event.preventDefault();
          onRequestRename();
        }
      }}
    >
      {children}
    </span>
  );
}
