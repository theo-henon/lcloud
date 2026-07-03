import type { UserRole } from "@/lib/api";
import { Badge } from "@/components/ui/badge";

export function RoleBadge({ role }: { role: UserRole }) {
  if (role === "admin") {
    return (
      <Badge className="border-primary/40 text-primary">admin</Badge>
    );
  }
  return <Badge>user</Badge>;
}
