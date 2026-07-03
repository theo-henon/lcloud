import { useState } from "react";
import type { DiskInfo, FilterMode } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { formatBytes, parseQuotaGB } from "@/lib/utils";

function diskOptionKey(disk: DiskInfo) {
  return disk.path || disk.id;
}

function diskOptionLabel(disk: DiskInfo) {
  const freeLabel = formatBytes(disk.free_bytes);
  const totalLabel = formatBytes(disk.total_bytes);
  return `${disk.label} — ${freeLabel} free of ${totalLabel}`;
}

type CreateVolumeFormProps = {
  disks: DiskInfo[];
  loading?: boolean;
  onSubmit: (values: {
    name: string;
    disk_path?: string;
    disk_id?: string;
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
  const [diskKey, setDiskKey] = useState(diskOptionKey(disks[0] ?? { id: "", path: "", name: "", label: "", total_bytes: 0, free_bytes: 0 }));
  const [quotaGB, setQuotaGB] = useState("");
  const [filterMode, setFilterMode] = useState<FilterMode>("allow");
  const [extensions, setExtensions] = useState("");

  const selectedDisk = disks.find((disk) => diskOptionKey(disk) === diskKey);

  return (
    <form
      className="space-y-4"
      onSubmit={(event) => {
        event.preventDefault();
        if (!selectedDisk) {
          return;
        }

        const normalizedExtensions = extensions
          .split(/[\s,]+/)
          .map((item) => item.trim())
          .filter(Boolean)
          .map((item) => (item.startsWith(".") ? item : `.${item}`));

        onSubmit({
          name: name.trim(),
          disk_path: selectedDisk.path || undefined,
          disk_id: selectedDisk.path ? undefined : selectedDisk.id,
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
          Storage
        </label>
        <select
          id="volume-disk"
          className="flex h-10 w-full rounded-md border border-hairline bg-surface-soft px-3 text-sm text-ink"
          value={diskKey}
          onChange={(event) => setDiskKey(event.target.value)}
          required
        >
          {disks.map((disk) => (
            <option key={diskOptionKey(disk)} value={diskOptionKey(disk)}>
              {diskOptionLabel(disk)}
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
        <Button type="submit" disabled={loading || !name.trim() || !selectedDisk}>
          {loading ? "Creating..." : "Create volume"}
        </Button>
      </div>
    </form>
  );
}
