import { Header } from "@/components/layout/Header";

export function NotFoundPage() {
  return (
    <>
      <Header title="Page not found" description="This route does not exist yet." />
      <section className="px-8 py-6">
        <div className="rounded-lg border border-hairline bg-surface-card p-6 text-sm text-muted">
          Check the sidebar or return to the dashboard.
        </div>
      </section>
    </>
  );
}
