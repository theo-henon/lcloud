import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useVolumeSearch } from "@/hooks/useVolumeSearch";
import type { VolumeSummary } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

type VolumeSearchPanelProps = {
  volumes: VolumeSummary[];
};

const mimeOptions = [
  { label: "All types", value: "" },
  { label: "Images", value: "image/" },
  { label: "Documents", value: "application/pdf" },
  { label: "Videos", value: "video/" },
  { label: "Audio", value: "audio/" },
];

function parentPath(relativePath: string): string {
  const parts = relativePath.split("/");
  parts.pop();
  const joined = parts.join("/");
  return joined || ".";
}

export function VolumeSearchPanel({ volumes }: VolumeSearchPanelProps) {
  const navigate = useNavigate();
  const [volumeId, setVolumeId] = useState(volumes[0]?.id ?? "");
  const [query, setQuery] = useState("");
  const [mimePrefix, setMimePrefix] = useState("");

  const searchParams = useMemo(
    () => ({
      q: query,
      mime_prefix: mimePrefix || undefined,
    }),
    [query, mimePrefix],
  );

  const searchQuery = useVolumeSearch(volumeId || undefined, searchParams);

  return (
    <Card className="space-y-4 p-6">
      <div>
        <h3 className="text-lg font-semibold text-ink">Search files</h3>
        <p className="mt-1 text-sm text-muted">
          Find files by name within a volume. Click a result to open its folder.
        </p>
      </div>

      <div className="grid gap-3 md:grid-cols-3">
        <select
          className="h-10 rounded-md border border-hairline bg-surface-card px-3 text-sm text-body"
          value={volumeId}
          onChange={(event) => setVolumeId(event.target.value)}
        >
          {volumes.map((volume) => (
            <option key={volume.id} value={volume.id}>
              {volume.name}
            </option>
          ))}
        </select>
        <Input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder='Search e.g. "vacation"'
        />
        <select
          className="h-10 rounded-md border border-hairline bg-surface-card px-3 text-sm text-body"
          value={mimePrefix}
          onChange={(event) => setMimePrefix(event.target.value)}
        >
          {mimeOptions.map((option) => (
            <option key={option.label} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      {searchQuery.isFetching ? (
        <p className="text-sm text-muted">Searching...</p>
      ) : null}

      {searchQuery.data && searchQuery.data.total === 0 ? (
        <p className="text-sm text-muted">No matching files.</p>
      ) : null}

      {searchQuery.data && searchQuery.data.total > 0 ? (
        <div className="overflow-hidden rounded-lg border border-hairline">
          <table className="min-w-full text-left text-sm">
            <thead className="bg-surface-soft text-muted">
              <tr>
                <th className="px-4 py-3 font-medium">Name</th>
                <th className="px-4 py-3 font-medium">Path</th>
                <th className="px-4 py-3 font-medium">Size</th>
              </tr>
            </thead>
            <tbody>
              {searchQuery.data.results.map((result) => (
                <tr
                  key={result.relative_path}
                  className="cursor-pointer border-t border-hairline hover:bg-surface-soft"
                  onClick={() => {
                    const path = parentPath(result.relative_path);
                    const params = new URLSearchParams({ path });
                    navigate(`/volumes/${volumeId}?${params}`);
                  }}
                >
                  <td className="px-4 py-3 text-ink">{result.name}</td>
                  <td className="px-4 py-3 text-muted">{result.relative_path}</td>
                  <td className="px-4 py-3 text-body">{formatBytes(result.size_bytes)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </Card>
  );
}
