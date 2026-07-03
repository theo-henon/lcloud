import {
  File,
  FileArchive,
  FileAudio,
  FileCode,
  FileImage,
  FileText,
  FileType,
  FileVideo,
  Folder,
  type LucideIcon,
} from "lucide-react";

type FileIconEntry = {
  type: string;
  mime_type?: string;
  name: string;
};

const TEXT_EXTENSIONS = new Set([
  ".txt",
  ".md",
  ".json",
  ".yaml",
  ".yml",
  ".log",
  ".csv",
]);

const ARCHIVE_EXTENSIONS = new Set([".zip", ".tar", ".gz", ".tgz", ".bz2", ".7z", ".rar"]);

const CODE_EXTENSIONS = new Set([
  ".js",
  ".ts",
  ".tsx",
  ".jsx",
  ".go",
  ".py",
  ".rs",
  ".xml",
  ".html",
  ".css",
  ".sh",
]);

function extension(name: string): string {
  if (!name.includes(".")) {
    return "";
  }
  return name.slice(name.lastIndexOf(".")).toLowerCase();
}

export function getFileIcon(entry: FileIconEntry): LucideIcon {
  if (entry.type === "directory") {
    return Folder;
  }

  const mime = entry.mime_type ?? "";
  const ext = extension(entry.name);

  if (mime.startsWith("image/")) {
    return FileImage;
  }
  if (mime.startsWith("video/")) {
    return FileVideo;
  }
  if (mime.startsWith("audio/")) {
    return FileAudio;
  }
  if (mime === "application/pdf" || ext === ".pdf") {
    return FileType;
  }
  if (
    mime.startsWith("text/") ||
    mime.includes("document") ||
    mime.includes("word") ||
    TEXT_EXTENSIONS.has(ext)
  ) {
    return FileText;
  }
  if (
    mime.includes("zip") ||
    mime.includes("gzip") ||
    mime.includes("x-tar") ||
    mime.includes("x-7z") ||
    ARCHIVE_EXTENSIONS.has(ext)
  ) {
    return FileArchive;
  }
  if (
    mime.includes("json") ||
    mime.includes("javascript") ||
    mime.includes("xml") ||
    CODE_EXTENSIONS.has(ext)
  ) {
    return FileCode;
  }

  return File;
}
