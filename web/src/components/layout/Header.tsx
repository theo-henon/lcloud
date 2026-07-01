export function Header({ title, description }: { title: string; description?: string }) {
  return (
    <header className="border-b border-hairline px-8 py-6">
      <h1 className="text-2xl font-bold text-ink">{title}</h1>
      {description ? <p className="mt-2 text-sm text-muted">{description}</p> : null}
    </header>
  );
}
