import { useEffect, useState } from "react";
import { Link, useNavigate, useParams, useSearchParams } from "react-router-dom";
import { Header } from "@/components/layout/Header";
import { VolumeExplorer } from "@/components/volumes/explorer/VolumeExplorer";
import { ProtocolSettingsPanel } from "@/components/volumes/ProtocolSettingsPanel";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { useVolume } from "@/hooks/useVolumes";
import { formatBytes } from "@/lib/utils";

export function VolumeDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const initialPath = searchParams.get("path") ?? ".";
  const isTrashView = searchParams.get("view") === "trash";
  const [currentPath, setCurrentPath] = useState(initialPath);

  useEffect(() => {
    if (searchParams.get("view") === "trash") {
      return;
    }
    setCurrentPath(searchParams.get("path") ?? ".");
  }, [searchParams]);

  const volumeQuery = useVolume(id);
  const volume = volumeQuery.data;

  const handlePathChange = (path: string) => {
    const normalized = path || ".";
    setCurrentPath(normalized);
    if (!id) {
      return;
    }
    const query = normalized === "." ? "" : `?path=${encodeURIComponent(normalized)}`;
    navigate(`/volumes/${id}${query}`, { replace: true });
  };

  const handleTrashSelect = () => {
    if (!id) {
      return;
    }
    navigate(`/volumes/${id}?view=trash`, { replace: true });
  };

  if (volumeQuery.isLoading) {
    return <div className="px-8 py-6 text-sm text-muted">Loading volume...</div>;
  }

  if (!volume || !id) {
    return (
      <div className="px-8 py-6">
        <Card className="p-6 text-sm text-body">Volume not found.</Card>
      </div>
    );
  }

  const quotaLabel =
    volume.quota_bytes > 0
      ? `${formatBytes(volume.used_bytes)} / ${formatBytes(volume.quota_bytes)}`
      : `${formatBytes(volume.used_bytes)} used — unlimited quota`;

  return (
    <>
      <Header
        title={volume.name}
        description={
          volume.disk_path ? `${volume.disk_path} · ${quotaLabel}` : quotaLabel
        }
        action={
          <Link
            to="/volumes"
            className="inline-flex h-10 items-center justify-center rounded-md border border-hairline px-4 text-sm font-semibold text-body hover:border-hairline-strong hover:text-ink"
          >
            Back to volumes
          </Link>
        }
      />

      <section className="px-8 py-6">
        <VolumeExplorer
          volumeId={id}
          currentPath={currentPath}
          onPathChange={handlePathChange}
          isTrashView={isTrashView}
          onTrashSelect={handleTrashSelect}
        />
        <ProtocolSettingsPanel volumeId={id} />
      </section>
    </>
  );
}
