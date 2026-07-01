import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { DiskInfo, FilterMode } from "@/lib/api";
import { parseQuotaGB } from "@/lib/utils";

type CreateVolumeFormProps = {
  disks: DiskInfo[];
  loading?: boolean;
  onSubmit: (values: {
    name: string;
    disk_path: string;
    quota_bytes: number;
    filters: { mode: FilterMode | ""; extensions: string[] };
  }) => void;
  onCancel: () => void;
};

export function CreateVolumeForm({
  disks,
  loading,
  onSubmit,
  onCancel,
}: CreateVolumeFormProps) {
  const [name, setName] = useState("");
  const [diskPath, setDiskPath] = useState(disks[0]?.path ?? "");
  const [quotaGB, setQuotaGB] = useState("");
  const [filterMode, setFilterMode] = useState<FilterMode>("allow");
  const [extensions, setExtensions] = useState("");

  return (
    <form
      className="space-y-4"
      onSubmit={(event) => {
        event.preventDefault();
        const normalizedExtensions = extensions
          .split(/[\s,]+/)
          .map((item) => item.trim())
          .filter(Boolean)
          .map((item) => (item.startsWith(".") ? item : `.${item}`));

        onSubmit({
          name: name.trim(),
          disk_path: diskPath,
          quota_bytes: parseQuotaGB(quotaGB),
          filters:
            normalizedExtensions.length === 0
              ? { mode: "", extensions: [] }
              : { mode: filterMode, extensions: normalizedExtensions },
        });
      }}
    >
      <div className="space-y-2">
        <label className="text-sm font-medium text-body-strong" htmlFor="volume-name">
          Name
        </label>
        <Input
          id="volume-name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder="Photos"
          required
        />
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-body-strong" htmlFor="volume-disk">
          Disk
        </label>
        <select
          id="volume-disk"
          className="flex h-10 w-full rounded-md border border-hairline bg-surface-soft px-3 text-sm text-ink"
          value={diskPath}
          onChange={(event) => setDiskPath(event.target.value)}
          required
        >
          {disks.map((disk) => (
            <option key={disk.path} value={disk.path}>
              {disk.label} — {Math.round((disk.free_bytes / disk.total_bytes) * 100 || 0)}% free
            </option>
          ))}
        </select>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-body-strong" htmlFor="volume-quota">
          Quota (GB, empty = unlimited)
        </label>
        <Input
          id="volume-quota"
          type="number"
          min="0"
          step="0.1"
          value={quotaGB}
          onChange={(event) => setQuotaGB(event.target.value)}
          placeholder="50"
        />
      </div>

      <div className="space-y-2">
        <span className="text-sm font-medium text-body-strong">File filter (optional)</span>
        <p className="text-xs text-muted">
          Leave extensions empty to accept all file types.
        </p>
        <div className="flex gap-2">
          <Button
            type="button"
            variant={filterMode === "allow" ? "default" : "outline"}
            onClick={() => setFilterMode("allow")}
          >
            Allow
          </Button>
          <Button
            type="button"
            variant={filterMode === "block" ? "default" : "outline"}
            onClick={() => setFilterMode("block")}
          >
            Block
          </Button>
        </div>
        <Input
          value={extensions}
          onChange={(event) => setExtensions(event.target.value)}
          placeholder=".jpg .png .webp"
        />
      </div>

      <div className="flex justify-end gap-2 pt-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button type="submit" disabled={loading || !name.trim() || !diskPath}>
          {loading ? "Creating..." : "Create volume"}
        </Button>
      </div>
    </form>
  );
}
