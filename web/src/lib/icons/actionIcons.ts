import {
  Columns3,
  Download,
  FolderOpen,
  FolderPlus,
  Grid3X3,
  List,
  LogOut,
  Pencil,
  Play,
  Plus,
  RefreshCw,
  Trash2,
  Upload,
  type LucideIcon,
} from "lucide-react";

export type ActionIconKey =
  | "upload"
  | "newFolder"
  | "download"
  | "delete"
  | "rename"
  | "refresh"
  | "run"
  | "edit"
  | "create"
  | "logout"
  | "open"
  | "listView"
  | "gridView"
  | "columns";

export const actionIcons: Record<ActionIconKey, LucideIcon> = {
  upload: Upload,
  newFolder: FolderPlus,
  download: Download,
  delete: Trash2,
  rename: Pencil,
  refresh: RefreshCw,
  run: Play,
  edit: Pencil,
  create: Plus,
  logout: LogOut,
  open: FolderOpen,
  listView: List,
  gridView: Grid3X3,
  columns: Columns3,
};

export const destructiveActions = new Set<ActionIconKey>(["delete"]);

export function isDestructiveAction(key: ActionIconKey): boolean {
  return destructiveActions.has(key);
}
