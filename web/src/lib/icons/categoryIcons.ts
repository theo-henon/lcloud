import {
  File,
  FileArchive,
  FileAudio,
  FileImage,
  FileText,
  FileVideo,
  type LucideIcon,
} from "lucide-react";

const categoryIcons: Record<string, LucideIcon> = {
  images: FileImage,
  videos: FileVideo,
  audio: FileAudio,
  documents: FileText,
  archives: FileArchive,
  other: File,
};

export function getCategoryIcon(category: string): LucideIcon {
  return categoryIcons[category] ?? File;
}
