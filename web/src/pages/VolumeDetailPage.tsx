import { useEffect, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { Header } from "@/components/layout/Header";
import { BreadcrumbNav } from "@/components/volumes/BreadcrumbNav";
import { FileBrowser } from "@/components/volumes/FileBrowser";
import { UploadZone, formatUploadError } from "@/components/volumes/UploadZone";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  useCreateDirectory,
  useDeleteFile,
  useUploadFile,
  useVolumeFiles,
} from "@/hooks/useVolumeFiles";
import { useVolume } from "@/hooks/useVolumes";
import { api } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

export function VolumeDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [searchParams] = useSearchParams();
  const initialPath = searchParams.get("path") ?? ".";
  const [currentPath, setCurrentPath] = useState(initialPath);
  const [folderName, setFolderName] = useState("");
  const [uploadError, setUploadError] = useState<string | null>(null);

  useEffect(() => {
    setCurrentPath(searchParams.get("path") ?? ".");
  }, [searchParams]);

  const volumeQuery = useVolume(id);
  const filesQuery = useVolumeFiles(id, currentPath);
  const uploadFile = useUploadFile(id ?? "");
  const createDirectory = useCreateDirectory(id ?? "");
  const deleteFile = useDeleteFile(id ?? "");

  const volume = volumeQuery.data;
  const listing = filesQuery.data;

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
        description={`${volume.disk_path} · ${quotaLabel}`}
        action={
          <Link
            to="/volumes"
            className="inline-flex h-10 items-center justify-center rounded-md border border-hairline px-4 text-sm font-semibold text-body hover:border-hairline-strong hover:text-ink"
          >
            Back to volumes
          </Link>
        }
      />

      <section className="space-y-6 px-8 py-6">
        <BreadcrumbNav
          volumeId={id}
          path={currentPath}
          onNavigate={(path) => setCurrentPath(path || ".")}
        />

        <div className="flex flex-wrap items-end gap-3">
          <div className="min-w-[240px] flex-1 space-y-2">
            <label className="text-sm font-medium text-body-strong" htmlFor="folder-name">
              New folder
            </label>
            <Input
              id="folder-name"
              value={folderName}
              onChange={(event) => setFolderName(event.target.value)}
              placeholder="2024/vacation"
            />
          </div>
          <Button
            disabled={!folderName.trim() || createDirectory.isPending}
            onClick={() => {
              const target =
                currentPath === "."
                  ? folderName.trim()
                  : `${currentPath}/${folderName.trim()}`.replace(/\/+/g, "/");
              void createDirectory.mutateAsync(target).then(() => setFolderName(""));
            }}
          >
            Create folder
          </Button>
        </div>

        <UploadZone
          uploading={uploadFile.isPending}
          onUpload={(files) => {
            setUploadError(null);
            void uploadFile
              .mutateAsync({ file: files[0], path: currentPath })
              .catch((error) => setUploadError(formatUploadError(error)));
          }}
        />

        {uploadError ? (
          <Card className="border-accent-rose/40 p-4 text-sm text-accent-rose">{uploadError}</Card>
        ) : null}

        {filesQuery.isLoading ? (
          <p className="text-sm text-muted">Loading files...</p>
        ) : (
          <FileBrowser
            volumeId={id}
            entries={listing?.entries ?? []}
            onOpenDirectory={(path) => setCurrentPath(path)}
            onDownload={(path) => {
              void api.downloadFile(id, path, path.split("/").pop() ?? "download");
            }}
            onDelete={(path) => {
              if (window.confirm("Delete this file?")) {
                void deleteFile.mutateAsync(path);
              }
            }}
          />
        )}
      </section>
    </>
  );
}
