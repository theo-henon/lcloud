import { useCallback, useRef, useState } from "react";
import { api } from "@/lib/api";
import { formatUploadError } from "@/lib/uploadErrors";

export type UploadQueueItem = {
  id: string;
  file: File;
  targetPath: string;
  state: "queued" | "uploading" | "done" | "error";
  error?: string;
  progress?: number;
  bytesUploaded?: number;
  bytesTotal?: number;
  speedBytesPerSec?: number;
};

const MAX_PARALLEL = 3;

export function useUploadQueue(
  volumeId: string,
  onComplete?: () => void,
  maxUploadBytes?: number,
) {
  const [items, setItems] = useState<UploadQueueItem[]>([]);
  const activeCount = useRef(0);
  const itemsRef = useRef<UploadQueueItem[]>([]);
  itemsRef.current = items;

  const setItemState = useCallback((id: string, patch: Partial<UploadQueueItem>) => {
    setItems((prev) => {
      const next = prev.map((item) => (item.id === id ? { ...item, ...patch } : item));
      itemsRef.current = next;
      return next;
    });
  }, []);

  const uploadOne = useCallback(
    async (item: UploadQueueItem) => {
      activeCount.current += 1;
      const startedAt = Date.now();
      setItemState(item.id, {
        state: "uploading",
        progress: 0,
        bytesUploaded: 0,
        bytesTotal: item.file.size,
        speedBytesPerSec: 0,
      });
      try {
        await api.uploadFile(volumeId, item.file, item.targetPath, (progress) => {
          const elapsedSec = Math.max((Date.now() - startedAt) / 1000, 0.1);
          setItemState(item.id, {
            state: "uploading",
            progress: progress.percent,
            bytesUploaded: progress.loaded,
            bytesTotal: progress.total,
            speedBytesPerSec: progress.loaded / elapsedSec,
          });
        });
        setItemState(item.id, {
          state: "done",
          progress: 100,
          bytesUploaded: item.file.size,
          bytesTotal: item.file.size,
          speedBytesPerSec: 0,
        });
        onComplete?.();
      } catch (error) {
        setItemState(item.id, {
          state: "error",
          error: formatUploadError(error, maxUploadBytes),
          speedBytesPerSec: 0,
        });
      } finally {
        activeCount.current -= 1;
        pump();
      }
    },
    [maxUploadBytes, onComplete, setItemState, volumeId],
  );

  const pump = useCallback(() => {
    if (activeCount.current >= MAX_PARALLEL) {
      return;
    }
    const queued = itemsRef.current.filter((item) => item.state === "queued");
    const slots = MAX_PARALLEL - activeCount.current;
    for (let i = 0; i < Math.min(slots, queued.length); i += 1) {
      void uploadOne(queued[i]);
    }
  }, [uploadOne]);

  const enqueueFiles = useCallback(
    (files: FileList | File[], targetPath: string) => {
      const list = Array.from(files);
      if (list.length === 0) {
        return;
      }
      const newItems: UploadQueueItem[] = list.map((file) => ({
        id: `${Date.now()}-${Math.random().toString(36).slice(2)}`,
        file,
        targetPath,
        state: "queued",
      }));
      setItems((prev) => {
        const next = [...prev, ...newItems];
        itemsRef.current = next;
        return next;
      });
      setTimeout(() => pump(), 0);
    },
    [pump],
  );

  const dismissItem = useCallback((id: string) => {
    setItems((prev) => {
      const next = prev.filter((item) => item.id !== id);
      itemsRef.current = next;
      return next;
    });
  }, []);

  const cancelQueued = useCallback(
    (id: string) => {
      const item = itemsRef.current.find((entry) => entry.id === id);
      if (item?.state === "queued") {
        dismissItem(id);
      }
    },
    [dismissItem],
  );

  return { items, enqueueFiles, dismissItem, cancelQueued };
}
