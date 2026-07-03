import { Button } from "@/components/ui/button";

type DashboardEditBarProps = {
  onDone: () => void;
  onCancel: () => void;
  onReset: () => void;
  onAddWidget: () => void;
  saving?: boolean;
  resetting?: boolean;
};

export function DashboardEditBar({
  onDone,
  onCancel,
  onReset,
  onAddWidget,
  saving = false,
  resetting = false,
}: DashboardEditBarProps) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <Button onClick={onDone} disabled={saving}>
        {saving ? "Saving…" : "Done"}
      </Button>
      <Button variant="outline" onClick={onCancel} disabled={saving || resetting}>
        Cancel
      </Button>
      <Button variant="outline" onClick={onReset} disabled={saving || resetting}>
        {resetting ? "Resetting…" : "Reset to default"}
      </Button>
      <Button variant="outline" onClick={onAddWidget} disabled={saving || resetting}>
        + Add widget
      </Button>
    </div>
  );
}
