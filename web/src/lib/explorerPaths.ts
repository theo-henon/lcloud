export function joinPath(parent: string, name: string): string {
  const cleanParent = parent === "." || parent === "" ? "" : parent;
  if (!cleanParent) {
    return name;
  }
  return `${cleanParent}/${name}`.replace(/\/+/g, "/");
}

export function parentPath(path: string): string {
  if (path === "." || path === "") {
    return ".";
  }
  const parts = path.split("/").filter(Boolean);
  parts.pop();
  return parts.length === 0 ? "." : parts.join("/");
}

export function fileName(path: string): string {
  const parts = path.split("/").filter(Boolean);
  return parts[parts.length - 1] ?? path;
}

export function normalizeExplorerPath(path: string): string {
  if (path === "" || path === ".") {
    return ".";
  }
  return path.replace(/\/+/g, "/").replace(/^\/+|\/+$/g, "");
}

/** Paths that should be expanded so the selected folder is visible, including one level under it. */
export function treeExpandedPathsForSelection(path: string): string[] {
  const normalized = normalizeExplorerPath(path);
  if (normalized === ".") {
    return ["."];
  }

  const parts = normalized.split("/").filter(Boolean);
  const expanded = ["."];
  let current = "";
  for (const part of parts) {
    current = current ? `${current}/${part}` : part;
    expanded.push(current);
  }
  return expanded;
}

export const LcloudDragType = "application/x-lcloud-path";
