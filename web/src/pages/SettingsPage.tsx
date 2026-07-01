import { Header } from "@/components/layout/Header";
import { Card } from "@/components/ui/card";
import { useSettings, useUpdateSettings } from "@/hooks/useSettings";
import { useAuthStore } from "@/store/auth";

function MaskDiskNamesToggle({
  enabled,
  disabled,
  onChange,
}: {
  enabled: boolean;
  disabled?: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <label className="inline-flex cursor-pointer items-center gap-3 text-sm text-body">
      <span>Hide disk names</span>
      <button
        type="button"
        role="switch"
        aria-checked={enabled}
        disabled={disabled}
        onClick={() => onChange(!enabled)}
        className={`relative h-6 w-11 rounded-full transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${
          enabled ? "bg-primary" : "bg-surface-elevated"
        }`}
      >
        <span
          className={`absolute top-0.5 left-0.5 h-5 w-5 rounded-full bg-ink transition-transform ${
            enabled ? "translate-x-5 bg-on-primary" : ""
          }`}
        />
      </button>
    </label>
  );
}

export function SettingsPage() {
  const user = useAuthStore((state) => state.user);
  const isAdmin = user?.role === "admin";
  const settingsQuery = useSettings();
  const updateSettings = useUpdateSettings();

  const maskDiskNames = settingsQuery.data?.mask_disk_names ?? false;

  return (
    <>
      <Header
        title="Settings"
        description="Instance-wide preferences for security and administration."
      />

      <section className="space-y-6 px-8 py-6">
        <Card className="space-y-4 p-6">
          <div>
            <h2 className="text-sm font-semibold uppercase tracking-wide text-muted">
              Security
            </h2>
            <p className="mt-2 text-sm text-body">
              When enabled, disk paths and labels are replaced with generic names such as
              &quot;Storage 1&quot; across Monitoring and Volumes. Only administrators can
              create volumes and see real disk names when doing so.
            </p>
          </div>

          {settingsQuery.isLoading ? (
            <p className="text-sm text-muted">Loading settings...</p>
          ) : settingsQuery.isError ? (
            <p className="text-sm text-accent-rose">Unable to load settings.</p>
          ) : isAdmin ? (
            <MaskDiskNamesToggle
              enabled={maskDiskNames}
              disabled={updateSettings.isPending}
              onChange={(value) => {
                void updateSettings.mutateAsync(value);
              }}
            />
          ) : (
            <p className="text-sm text-body">
              Disk names are currently{" "}
              <span className="font-medium text-ink">
                {maskDiskNames ? "hidden" : "visible"}
              </span>
              . Only an administrator can change this option.
            </p>
          )}
        </Card>
      </section>
    </>
  );
}
