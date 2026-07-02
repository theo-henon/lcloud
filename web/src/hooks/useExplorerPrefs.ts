import { useCallback, useEffect, useState } from "react";

export type ColumnId = "name" | "size" | "modified" | "type";
export type ViewMode = "list" | "grid";

export type ListColumnPref = {
  id: ColumnId;
  visible: boolean;
  width: number;
  order: number;
};

export type ExplorerPrefs = {
  viewMode: ViewMode;
  listColumns: ListColumnPref[];
  treeWidth: number;
};

const STORAGE_KEY = "lcloud.explorer.prefs.v1";

export const DEFAULT_EXPLORER_PREFS: ExplorerPrefs = {
  viewMode: "list",
  listColumns: [
    { id: "name", visible: true, width: 320, order: 0 },
    { id: "size", visible: true, width: 100, order: 1 },
    { id: "modified", visible: true, width: 180, order: 2 },
    { id: "type", visible: false, width: 120, order: 3 },
  ],
  treeWidth: 240,
};

export function loadExplorerPrefs(): ExplorerPrefs {
  if (typeof window === "undefined") {
    return DEFAULT_EXPLORER_PREFS;
  }
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return DEFAULT_EXPLORER_PREFS;
    }
    const parsed = JSON.parse(raw) as Partial<ExplorerPrefs>;
    return {
      ...DEFAULT_EXPLORER_PREFS,
      ...parsed,
      listColumns: parsed.listColumns ?? DEFAULT_EXPLORER_PREFS.listColumns,
    };
  } catch {
    return DEFAULT_EXPLORER_PREFS;
  }
}

export function saveExplorerPrefs(prefs: ExplorerPrefs) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs));
}

export function useExplorerPrefs() {
  const [prefs, setPrefsState] = useState<ExplorerPrefs>(loadExplorerPrefs);

  const setPrefs = useCallback((updater: ExplorerPrefs | ((prev: ExplorerPrefs) => ExplorerPrefs)) => {
    setPrefsState((prev) => {
      const next = typeof updater === "function" ? updater(prev) : updater;
      saveExplorerPrefs(next);
      return next;
    });
  }, []);

  useEffect(() => {
    saveExplorerPrefs(prefs);
  }, [prefs]);

  return { prefs, setPrefs };
}

export function visibleColumns(prefs: ExplorerPrefs): ListColumnPref[] {
  return [...prefs.listColumns]
    .filter((col) => col.visible)
    .sort((a, b) => a.order - b.order);
}
