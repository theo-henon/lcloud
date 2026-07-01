import { useDisplayPreferences } from "@/store/displayPreferences";

export function DiskMaskToggle() {
  const maskDiskNames = useDisplayPreferences((state) => state.maskDiskNames);
  const setMaskDiskNames = useDisplayPreferences((state) => state.setMaskDiskNames);

  return (
    <label className="inline-flex cursor-pointer items-center gap-3 text-sm text-body">
      <span>Hide disk names</span>
      <button
        type="button"
        role="switch"
        aria-checked={maskDiskNames}
        onClick={() => setMaskDiskNames(!maskDiskNames)}
        className={`relative h-6 w-11 rounded-full transition-colors ${
          maskDiskNames ? "bg-primary" : "bg-surface-elevated"
        }`}
      >
        <span
          className={`absolute top-0.5 left-0.5 h-5 w-5 rounded-full bg-ink transition-transform ${
            maskDiskNames ? "translate-x-5 bg-on-primary" : ""
          }`}
        />
      </button>
    </label>
  );
}
