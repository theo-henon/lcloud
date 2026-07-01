import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

type ThumbnailPreviewProps = {
  volumeId: string;
  path: string;
};

export function ThumbnailPreview({ volumeId, path }: ThumbnailPreviewProps) {
  const accessToken = useAuthStore((state) => state.accessToken);
  const [src, setSrc] = useState<string | null>(null);

  useEffect(() => {
    if (!accessToken) {
      return;
    }

    let objectUrl: string | null = null;
    let cancelled = false;

    void (async () => {
      const response = await fetch(api.fileThumbnailUrl(volumeId, path), {
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
    return null;
  }

  return <img src={src} alt="" className="h-10 w-10 rounded object-cover" />;
}
