import type { User } from "@/lib/api";
import { RoleBadge } from "@/components/users/RoleBadge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import { navIcons } from "@/lib/icons";

type UserTableProps = {
  users: User[];
  currentUserId: string | undefined;
  canModifyAdmin: (user: User) => boolean;
  onChangeRole: (user: User, role: "admin" | "user") => void;
  onToggleDisabled: (user: User, disabled: boolean) => void;
  onResetPassword: (user: User) => void;
  actionError: string | null;
};

function formatDate(value: string) {
  return new Date(value).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function StatusBadge({ user }: { user: User }) {
  if (user.disabled_at) {
    return <Badge className="border-accent-rose/40 text-accent-rose">Disabled</Badge>;
  }
  return <Badge className="border-accent-emerald/40 text-accent-emerald">Active</Badge>;
}

export function UserTable({
  users,
  currentUserId,
  canModifyAdmin,
  onChangeRole,
  onToggleDisabled,
  onResetPassword,
  actionError,
}: UserTableProps) {
  if (users.length === 0) {
    const EmptyIcon = navIcons["/settings/users"];
    return (
      <Card className="flex flex-col items-center gap-3 p-8 text-center text-sm text-muted">
        {EmptyIcon ? <EmptyIcon className="h-8 w-8 text-muted" aria-hidden /> : null}
        <p>No users yet — create one to share access.</p>
      </Card>
    );
  }

  return (
    <div className="space-y-3">
      {actionError ? (
        <p className="text-sm text-accent-rose">{actionError}</p>
      ) : null}
      <Card className="overflow-hidden p-0">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-hairline bg-surface-soft text-xs uppercase tracking-wide text-muted">
            <tr>
              <th className="px-4 py-3 font-semibold">Email</th>
              <th className="px-4 py-3 font-semibold">Role</th>
              <th className="px-4 py-3 font-semibold">Status</th>
              <th className="px-4 py-3 font-semibold">Created</th>
              <th className="px-4 py-3 font-semibold">Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => {
              const isSelf = user.id === currentUserId;
              const isDisabled = Boolean(user.disabled_at);
              const adminLocked = user.role === "admin" && !canModifyAdmin(user);

              return (
                <tr
                  key={user.id}
                  className={`border-b border-hairline last:border-b-0 ${isDisabled ? "text-muted" : "text-body"}`}
                >
                  <td className="px-4 py-3">
                    <span className="font-medium text-ink">{user.email}</span>
                    {isSelf ? (
                      <span className="ml-2 text-xs text-muted">(you)</span>
                    ) : null}
                  </td>
                  <td className="px-4 py-3">
                    <RoleBadge role={user.role} />
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge user={user} />
                  </td>
                  <td className="px-4 py-3">{formatDate(user.created_at)}</td>
                  <td className="px-4 py-3">
                    <DropdownMenu
                      align="right"
                      trigger={
                        <Button variant="outline" className="h-8 px-3 text-xs">
                          Actions
                        </Button>
                      }
                    >
                      {user.role === "user" ? (
                        <DropdownMenuItem
                          onSelect={() => onChangeRole(user, "admin")}
                        >
                          Promote to admin
                        </DropdownMenuItem>
                      ) : (
                        <DropdownMenuItem
                          disabled={adminLocked}
                          title={
                            adminLocked
                              ? "Cannot modify the last active admin"
                              : undefined
                          }
                          onSelect={() => onChangeRole(user, "user")}
                        >
                          Demote to user
                        </DropdownMenuItem>
                      )}
                      {isDisabled ? (
                        <DropdownMenuItem
                          onSelect={() => onToggleDisabled(user, false)}
                        >
                          Enable account
                        </DropdownMenuItem>
                      ) : (
                        <DropdownMenuItem
                          disabled={adminLocked}
                          title={
                            adminLocked
                              ? "Cannot modify the last active admin"
                              : undefined
                          }
                          onSelect={() => onToggleDisabled(user, true)}
                        >
                          Disable account
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuItem onSelect={() => onResetPassword(user)}>
                        Reset password
                      </DropdownMenuItem>
                    </DropdownMenu>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Card>
    </div>
  );
}

export function countOtherActiveAdmins(users: User[], targetId: string): number {
  return users.filter(
    (user) =>
      user.role === "admin" &&
      !user.disabled_at &&
      user.id !== targetId,
  ).length;
}
