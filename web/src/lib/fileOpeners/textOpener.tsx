import { registerFileOpener } from "@/lib/fileOpeners/registry";

const TEXT_EXTENSIONS = new Set([
  ".txt",
  ".md",
  ".json",
  ".yaml",
  ".yml",
  ".log",
  ".csv",
]);

function isTextEntry(entry: { mime_type?: string; name: string }) {
  const ext = entry.name.includes(".") ? entry.name.slice(entry.name.lastIndexOf(".")).toLowerCase() : "";
  return entry.mime_type?.startsWith("text/") || TEXT_EXTENSIONS.has(ext);
}

async function fetchTextContent(url: string, token: string | null): Promise<string> {
  const headers = token ? { Authorization: `Bearer ${token}` } : undefined;
  const response = await fetch(url, { headers });
  if (!response.ok) {
    throw new Error("Unable to load file content");
  }
  const blob = await response.blob();
  const maxBytes = 512 * 1024;
  const slice = blob.size > maxBytes ? blob.slice(0, maxBytes) : blob;
  const text = await slice.text();
  if (blob.size > maxBytes) {
    return `${text}\n\n— Preview truncated (512 KiB max) —`;
  }
  return text;
}

registerFileOpener({
  id: "core.text",
  label: "Text preview",
  priority: 90,
  canOpen: (entry) => entry.type === "file" && isTextEntry(entry),
  open: async (ctx, panel) => {
    panel.render(<p className="text-sm text-muted">Loading preview…</p>);
    const text = await fetchTextContent(ctx.contentUrl, ctx.accessToken);
    panel.render(
      <pre className="max-h-[70vh] overflow-auto whitespace-pre-wrap break-words font-mono text-sm text-body">
        {text}
      </pre>,
    );
  },
});
