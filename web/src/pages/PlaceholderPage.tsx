import { Header } from "@/components/layout/Header";

export function PlaceholderPage({
  title,
  phase,
  description,
}: {
  title: string;
  phase: string;
  description: string;
}) {
  return (
    <>
      <Header title={title} description={description} />
      <section className="px-8 py-6">
        <div className="rounded-lg border border-dashed border-hairline-strong bg-surface-soft p-6">
          <p className="text-sm text-muted">
            Coming in <span className="font-mono text-body">{phase}</span>.
          </p>
        </div>
      </section>
    </>
  );
}
