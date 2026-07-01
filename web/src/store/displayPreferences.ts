import { create } from "zustand";
import { persist } from "zustand/middleware";

interface DisplayPreferencesState {
  maskDiskNames: boolean;
  setMaskDiskNames: (value: boolean) => void;
}

export const useDisplayPreferences = create<DisplayPreferencesState>()(
  persist(
    (set) => ({
      maskDiskNames: false,
      setMaskDiskNames: (value) => set({ maskDiskNames: value }),
    }),
    { name: "lcloud.maskDiskNames" },
  ),
);
