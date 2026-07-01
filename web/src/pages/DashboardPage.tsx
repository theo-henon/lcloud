import { Header } from "@/components/layout/Header";
import { useAuthStore } from "@/store/auth";

export function DashboardPage() {
  const user = useAuthStore((state) => state.user);

  return (
    <>
      <Header
        title="Dashboard"
        description="Your lcloud instance is running. Create volumes and upload files from the Volumes page."
      />
      <section className="px-8 py-6">
        <div className="rounded-lg border border-hairline bg-surface-card p-6">
          <p className="text-sm text-body">
            Welcome{user ? `, ${user.email}` : ""}. This is the empty dashboard shell for Phase 0.
          </p>
        </div>
      </section>
    </>
  );
}
