import { registerFileOpener } from "@/lib/fileOpeners/registry";

registerFileOpener({
  id: "core.pdf",
  label: "PDF preview",
  priority: 80,
  canOpen: (entry) =>
    entry.mime_type === "application/pdf" || entry.name.toLowerCase().endsWith(".pdf"),
  open: async (ctx, panel) => {
    panel.render(<p className="text-sm text-muted">Loading preview…</p>);
    if (!ctx.accessToken) {
      panel.render(<p className="text-sm text-muted">Authentication required.</p>);
      return;
    }
    const response = await fetch(ctx.contentUrl, {
      headers: { Authorization: `Bearer ${ctx.accessToken}` },
    });
    if (!response.ok) {
      panel.render(<p className="text-sm text-accent-rose">Unable to load PDF.</p>);
      return;
    }
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    panel.render(
      <iframe
        title={ctx.entry.name}
        src={url}
        className="h-[70vh] w-full rounded border border-hairline bg-surface-soft"
      />,
    );
  },
});
