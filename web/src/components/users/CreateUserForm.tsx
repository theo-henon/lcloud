import { FormEvent, useState } from "react";
import type { UserRole } from "@/lib/api";
import { ApiError } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

type CreateUserFormProps = {
  onSubmit: (input: { email: string; password: string; role: UserRole }) => void;
  onCancel: () => void;
  isPending?: boolean;
};

export function CreateUserForm({ onSubmit, onCancel, isPending }: CreateUserFormProps) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<UserRole>("user");

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    onSubmit({ email, password, role });
  }

  return (
    <Card className="p-6">
      <h3 className="mb-4 text-base font-semibold text-ink">Create user</h3>
      <form className="space-y-4" onSubmit={handleSubmit}>
        <div className="space-y-2">
          <label className="text-sm text-muted" htmlFor="create-user-email">
            Email
          </label>
          <Input
            id="create-user-email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm text-muted" htmlFor="create-user-password">
            Password
          </label>
          <Input
            id="create-user-password"
            type="password"
            required
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label className="text-sm text-muted" htmlFor="create-user-role">
            Role
          </label>
          <select
            id="create-user-role"
            className="w-full rounded-md border border-hairline bg-surface px-3 py-2 text-sm text-ink"
            value={role}
            onChange={(e) => setRole(e.target.value as UserRole)}
          >
            <option value="user">user</option>
            <option value="admin">admin</option>
          </select>
        </div>
        <div className="flex gap-3">
          <Button type="submit" disabled={isPending}>
            {isPending ? "Creating..." : "Create user"}
          </Button>
          <Button type="button" variant="outline" onClick={onCancel} disabled={isPending}>
            Cancel
          </Button>
        </div>
      </form>
    </Card>
  );
}

export function getCreateUserErrorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    return err.message;
  }
  return "Unable to create user.";
}
