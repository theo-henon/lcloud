import { useState } from "react";
import type { User } from "@/lib/api";
import { ApiError } from "@/lib/api";
import {
  CreateUserForm,
  getCreateUserErrorMessage,
} from "@/components/users/CreateUserForm";
import { ResetPasswordPanel } from "@/components/users/ResetPasswordPanel";
import {
  countOtherActiveAdmins,
  UserTable,
} from "@/components/users/UserTable";
import { Header } from "@/components/layout/Header";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
  useAdminUsers,
  useCreateAdminUser,
  usePatchAdminUser,
} from "@/hooks/useUsers";
import { ActionIcon } from "@/lib/icons";
import { useAuthStore } from "@/store/auth";

export function UsersPage() {
  const currentUser = useAuthStore((state) => state.user);
  const usersQuery = useAdminUsers();
  const createUser = useCreateAdminUser();
  const patchUser = usePatchAdminUser();

  const [showCreateForm, setShowCreateForm] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [resetTarget, setResetTarget] = useState<User | null>(null);

  if (usersQuery.isLoading) {
    return (
      <>
        <Header
          title="Users"
          description="Manage who can access this lcloud instance."
        />
        <section className="px-8 py-6">
          <p className="text-sm text-muted">Loading users...</p>
        </section>
      </>
    );
  }

  if (usersQuery.isError || !usersQuery.data) {
    return (
      <>
        <Header
          title="Users"
          description="Manage who can access this lcloud instance."
        />
        <section className="px-8 py-6">
          <Card className="p-6 text-sm text-accent-rose">Unable to load users.</Card>
        </section>
      </>
    );
  }

  const users = usersQuery.data.users;

  function canModifyAdmin(user: User): boolean {
    if (user.role !== "admin" || user.disabled_at) {
      return true;
    }
    return countOtherActiveAdmins(users, user.id) > 0;
  }

  function handlePatch(
    id: string,
    input: Parameters<typeof patchUser.mutate>[0]["input"],
    options?: { confirm?: string },
  ) {
    if (options?.confirm && !window.confirm(options.confirm)) {
      return;
    }
    setActionError(null);
    patchUser.mutate(
      { id, input },
      {
        onError: (err) => {
          setActionError(err instanceof ApiError ? err.message : "Unable to update user.");
        },
      },
    );
  }

  return (
    <>
      <Header
        title="Users"
        description="Manage who can access this lcloud instance."
      />

      <section className="space-y-6 px-8 py-6">
        <div className="flex flex-wrap items-center justify-end gap-4">
          <Button
            onClick={() => {
              setCreateError(null);
              setShowCreateForm((value) => !value);
              setResetTarget(null);
            }}
          >
            <ActionIcon action="create" className="mr-2" />
            Create user
          </Button>
        </div>

        {showCreateForm ? (
          <div className="space-y-2">
            <CreateUserForm
              isPending={createUser.isPending}
              onCancel={() => {
                setShowCreateForm(false);
                setCreateError(null);
              }}
              onSubmit={(input) => {
                setCreateError(null);
                createUser.mutate(input, {
                  onSuccess: () => setShowCreateForm(false),
                  onError: (err) => setCreateError(getCreateUserErrorMessage(err)),
                });
              }}
            />
            {createError ? (
              <p className="text-sm text-accent-rose">{createError}</p>
            ) : null}
          </div>
        ) : null}

        {resetTarget ? (
          <ResetPasswordPanel
            email={resetTarget.email}
            isPending={patchUser.isPending}
            onCancel={() => setResetTarget(null)}
            onSubmit={(password) => {
              handlePatch(resetTarget.id, { password });
              setResetTarget(null);
            }}
          />
        ) : null}

        <UserTable
          users={users}
          currentUserId={currentUser?.id}
          canModifyAdmin={canModifyAdmin}
          actionError={actionError}
          onChangeRole={(user, role) => {
            const confirm =
              user.role === "admin" && role === "user"
                ? user.id === currentUser?.id
                  ? "You are about to remove your own admin access. Continue?"
                  : "Demote this admin to user?"
                : undefined;
            handlePatch(user.id, { role }, { confirm });
          }}
          onToggleDisabled={(user, disabled) => {
            const confirm = disabled
              ? user.role === "admin"
                ? "Ensure at least one other active admin exists. Disable this account? Sessions will be revoked."
                : user.id === currentUser?.id
                  ? "You are about to disable your own account. Continue?"
                  : "Disable this account? The user will not be able to sign in."
              : undefined;
            handlePatch(user.id, { disabled }, { confirm });
          }}
          onResetPassword={(user) => {
            setResetTarget(user);
            setShowCreateForm(false);
          }}
        />
      </section>
    </>
  );
}
