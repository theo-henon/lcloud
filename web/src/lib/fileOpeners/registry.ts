import type { ReactNode } from "react";
import type { FileEntry } from "@/lib/api";

export type PreviewPanelApi = {
  render: (node: ReactNode) => void;
  close: () => void;
};

export type FileOpenerContext = {
  volumeId: string;
  entry: FileEntry;
  contentUrl: string;
  accessToken: string | null;
};

export type FileOpener = {
  id: string;
  label: string;
  priority: number;
  canOpen: (entry: FileEntry) => boolean;
  open: (ctx: FileOpenerContext, panel: PreviewPanelApi) => void | Promise<void>;
};

const openers: FileOpener[] = [];

export function registerFileOpener(opener: FileOpener) {
  openers.push(opener);
  openers.sort((a, b) => b.priority - a.priority);
}

export function resolveFileOpener(entry: FileEntry): FileOpener | null {
  return openers.find((opener) => opener.canOpen(entry)) ?? null;
}

export function listFileOpeners(): FileOpener[] {
  return [...openers];
}
