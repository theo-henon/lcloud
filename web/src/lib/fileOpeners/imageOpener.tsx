import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { registerFileOpener } from "@/lib/fileOpeners/registry";

function AuthenticatedImage({
  volumeId,
  path,
  accessToken,
}: {
  volumeId: string;
  path: string;
  accessToken: string | null;
}) {
  const [src, setSrc] = useState<string | null>(null);

  useEffect(() => {
    if (!accessToken) {
      return;
    }
    let objectUrl: string | null = null;
    let cancelled = false;
    void (async () => {
      const response = await fetch(api.fileContentUrl(volumeId, path, "inline"), {
        headers: { Authorization: `Bearer ${accessToken}` },
      });
      if (!response.ok || cancelled) {
        return;
      }
      const blob = await response.blob();
      objectUrl = URL.createObjectURL(blob);
      if (!cancelled) {
        setSrc(objectUrl);
      }
    })();
    return () => {
      cancelled = true;
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl);
      }
    };
  }, [accessToken, volumeId, path]);

  if (!src) {
    return <p className="text-sm text-muted">Loading preview…</p>;
  }
  return <img src={src} alt="" className="max-h-[70vh] max-w-full object-contain" />;
}

registerFileOpener({
  id: "core.image",
  label: "Image preview",
  priority: 100,
  canOpen: (entry) => entry.mime_type?.startsWith("image/") ?? false,
  open: (ctx, panel) => {
    panel.render(
      <AuthenticatedImage
        volumeId={ctx.volumeId}
        path={ctx.entry.path}
        accessToken={ctx.accessToken}
      />,
    );
  },
});
