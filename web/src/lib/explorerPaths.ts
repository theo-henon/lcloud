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

export const LcloudDragType = "application/x-lcloud-path";
