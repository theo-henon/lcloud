import { FormEvent, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

type ResetPasswordPanelProps = {
  email: string;
  onSubmit: (password: string) => void;
  onCancel: () => void;
  isPending?: boolean;
};

export function ResetPasswordPanel({
  email,
  onSubmit,
  onCancel,
  isPending,
}: ResetPasswordPanelProps) {
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState<string | null>(null);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);

    if (password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }
    if (password !== confirm) {
      setError("Passwords do not match.");
      return;
    }

    onSubmit(password);
  }

  return (
    <Card className="p-6">
      <h3 className="mb-1 text-base font-semibold text-ink">Reset password</h3>
      <p className="mb-4 text-sm text-muted">Set a new password for {email}.</p>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <label className="text-sm text-muted" htmlFor="reset-password">
            New password
          </label>
          <Input
            id="reset-password"
            type="password"
            required
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm text-muted" htmlFor="reset-password-confirm">
            Confirm password
          </label>
          <Input
            id="reset-password-confirm"
            type="password"
            required
            minLength={8}
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
          />
        </div>
        {error ? <p className="text-sm text-accent-rose">{error}</p> : null}
        <div className="flex gap-3">
          <Button type="submit" disabled={isPending}>
            {isPending ? "Saving..." : "Save password"}
          </Button>
          <Button type="button" variant="outline" onClick={onCancel} disabled={isPending}>
            Cancel
          </Button>
        </div>
      </form>
    </Card>
  );
}
