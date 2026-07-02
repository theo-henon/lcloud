import "@/lib/fileOpeners/imageOpener";
import "@/lib/fileOpeners/textOpener";
import "@/lib/fileOpeners/pdfOpener";

export { registerFileOpener, resolveFileOpener, listFileOpeners } from "@/lib/fileOpeners/registry";
export type { FileOpener, FileOpenerContext, PreviewPanelApi } from "@/lib/fileOpeners/registry";
