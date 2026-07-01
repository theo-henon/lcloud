import * as React from "react";
import { cn } from "@/lib/utils";

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "default" | "ghost" | "outline";
};

export function Button({
  className,
  variant = "default",
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(
        "inline-flex h-10 items-center justify-center rounded-md px-4 text-sm font-semibold transition-colors disabled:cursor-not-allowed disabled:opacity-50",
        variant === "default" &&
          "bg-primary text-on-primary hover:bg-primary-active",
        variant === "ghost" &&
          "text-body hover:bg-surface-elevated hover:text-ink",
        variant === "outline" &&
          "border border-hairline bg-transparent text-body hover:border-hairline-strong hover:text-ink",
        className,
      )}
      {...props}
    />
  );
}
