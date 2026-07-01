export function Header({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <header className="border-b border-hairline px-8 py-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-ink">{title}</h1>
          {description ? <p className="mt-2 text-sm text-muted">{description}</p> : null}
        </div>
        {action ? <div>{action}</div> : null}
      </div>
    </header>
  );
}
