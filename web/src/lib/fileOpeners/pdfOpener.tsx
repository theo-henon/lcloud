import { useEffect, useRef, useState } from "react";
import { registerFileOpener } from "@/lib/fileOpeners/registry";

function AuthenticatedPdf({
  contentUrl,
  accessToken,
  name,
}: {
  contentUrl: string;
  accessToken: string | null;
  name: string;
}) {
  const [src, setSrc] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const objectUrlRef = useRef<string | null>(null);

  useEffect(() => {
    if (!accessToken) {
      setError("Authentication required.");
      return;
    }

    let cancelled = false;
    void (async () => {
      const response = await fetch(contentUrl, {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      if (!response.ok) {
        if (!cancelled) {
          setError("Unable to load PDF.");
        }
        return;
      }
      const blob = await response.blob();
      const objectUrl = URL.createObjectURL(blob);
      objectUrlRef.current = objectUrl;
      if (!cancelled) {
        setSrc(objectUrl);
      } else {
        URL.revokeObjectURL(objectUrl);
        objectUrlRef.current = null;
      }
    })();

    return () => {
      cancelled = true;
      if (objectUrlRef.current) {
        URL.revokeObjectURL(objectUrlRef.current);
        objectUrlRef.current = null;
      }
    };
  }, [accessToken, contentUrl]);

  if (error) {
    return <p className="text-sm text-accent-rose">{error}</p>;
  }
  if (!src) {
    return <p className="text-sm text-muted">Loading preview…</p>;
  }
  return (
    <iframe
      title={name}
      src={src}
      className="h-[70vh] w-full rounded border border-hairline bg-surface-soft"
    />
  );
}

registerFileOpener({
  id: "core.pdf",
  label: "PDF preview",
  priority: 80,
  canOpen: (entry) =>
    entry.mime_type === "application/pdf" || entry.name.toLowerCase().endsWith(".pdf"),
  open: (ctx, panel) => {
    panel.render(
      <AuthenticatedPdf
        contentUrl={ctx.contentUrl}
        accessToken={ctx.accessToken}
        name={ctx.entry.name}
      />,
    );
  },
});
