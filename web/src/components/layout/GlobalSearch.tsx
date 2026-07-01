import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Input } from "@/components/ui/input";
import { useGlobalSearch } from "@/hooks/useGlobalSearch";
import { formatBytes } from "@/lib/utils";

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

export function GlobalSearch() {
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [mimePrefix, setMimePrefix] = useState("");

  const searchParams = useMemo(
    () => ({
      q: query,
      mime_prefix: mimePrefix || undefined,
    }),
    [query, mimePrefix],
  );

  const searchQuery = useGlobalSearch(searchParams);
  const isActive = Boolean(query.trim() || mimePrefix);
  const showPanel =
    isActive &&
    (searchQuery.isFetching || searchQuery.data !== undefined);

  return (
    <>
      <div className="border-b border-hairline px-4 py-4">
        <label
          className="mb-2 block text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-soft"
          htmlFor="global-search"
        >
          Search files
        </label>
        <div className="space-y-2">
          <Input
            id="global-search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder='e.g. "vacation"'
            className="h-9 text-sm"
          />
          <select
            className="h-9 w-full rounded-md border border-hairline bg-surface-card px-2 text-xs text-body"
            value={mimePrefix}
            onChange={(event) => setMimePrefix(event.target.value)}
            aria-label="Filter by file type"
          >
            {mimeOptions.map((option) => (
              <option key={option.label} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {showPanel ? (
        <div className="fixed inset-y-0 left-64 z-40 flex w-[min(28rem,calc(100vw-16rem))] flex-col border-r border-hairline bg-surface-card shadow-xl">
          <div className="border-b border-hairline px-4 py-3">
            <p className="text-sm font-semibold text-ink">Search results</p>
            <p className="mt-0.5 text-xs text-muted">Across all volumes</p>
          </div>

          <div className="flex-1 overflow-y-auto p-2">
            {searchQuery.isFetching ? (
              <p className="px-2 py-3 text-sm text-muted">Searching...</p>
            ) : null}

            {searchQuery.data && searchQuery.data.total === 0 ? (
              <p className="px-2 py-3 text-sm text-muted">No matching files.</p>
            ) : null}

            {searchQuery.data && searchQuery.data.total > 0 ? (
              <ul className="space-y-1">
                {searchQuery.data.results.map((result) => (
                  <li key={`${result.volume_id}:${result.relative_path}`}>
                    <button
                      type="button"
                      className="w-full rounded-md px-3 py-2 text-left transition-colors hover:bg-surface-elevated"
                      onClick={() => {
                        const path = parentPath(result.relative_path);
                        const params = new URLSearchParams({ path });
                        setQuery("");
                        navigate(`/volumes/${result.volume_id}?${params}`);
                      }}
                    >
                      <p className="truncate text-sm font-medium text-ink">{result.name}</p>
                      <p className="truncate text-xs text-muted">{result.volume_name}</p>
                      <p className="truncate text-xs text-muted-soft">{result.relative_path}</p>
                      <p className="mt-1 text-xs text-body">{formatBytes(result.size_bytes)}</p>
                    </button>
                  </li>
                ))}
              </ul>
            ) : null}
          </div>
        </div>
      ) : null}
    </>
  );
}
