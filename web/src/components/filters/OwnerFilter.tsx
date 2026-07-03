import type { User } from "@/lib/api";

type OwnerFilterProps = {
  value: string;
  currentUserId: string;
  currentUserEmail: string;
  users: User[];
  onChange: (ownerId: string) => void;
};

export function OwnerFilter({
  value,
  currentUserId,
  currentUserEmail,
  users,
  onChange,
}: OwnerFilterProps) {
  const otherUsers = users.filter((user) => user.id !== currentUserId);

  return (
    <div className="flex items-center gap-3">
      <label className="text-sm text-muted">Filter by user</label>
      <select
        className="rounded-md border border-hairline bg-surface px-3 py-2 text-sm text-ink"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      >
        <option value="">All users</option>
        <option value={currentUserId}>Me ({currentUserEmail})</option>
        {otherUsers.map((user) => (
          <option key={user.id} value={user.id}>
            {user.email}
          </option>
        ))}
      </select>
    </div>
  );
}
