import { useCallback, useMemo, useState } from "react";
import {
  DndContext,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import { Header } from "@/components/layout/Header";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { DashboardEditBar } from "@/components/dashboard/DashboardEditBar";
import { DashboardGrid } from "@/components/dashboard/DashboardGrid";
import { WidgetPicker } from "@/components/dashboard/WidgetPicker";
import {
  useDashboardLayout,
  useResetDashboardLayout,
  useSaveDashboardLayout,
} from "@/hooks/useDashboardLayout";
import { addWidget, gridExtent, hasOverlap, moveWidget, removeWidget } from "@/lib/dashboard/placement";
import type { DashboardLayout, WidgetPlacement, WidgetType } from "@/lib/dashboard/types";
import { getWidgetDefinition } from "@/lib/dashboard/registry";

export function DashboardPage() {
  const layoutQuery = useDashboardLayout();
  const saveLayout = useSaveDashboardLayout();
  const resetLayout = useResetDashboardLayout();

  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState<DashboardLayout | null>(null);
  const [pickerOpen, setPickerOpen] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }));

  const activeLayout = editing && draft ? draft : layoutQuery.data?.layout;
  const widgets = activeLayout?.widgets ?? [];

  const { rows: maxRows } = useMemo(() => gridExtent(widgets), [widgets]);

  const existingTypes = useMemo(
    () => new Set(widgets.map((widget) => widget.type)),
    [widgets],
  );

  const enterEditMode = useCallback(() => {
    if (!layoutQuery.data) {
      return;
    }
    setDraft(structuredClone(layoutQuery.data.layout));
    setSaveError(null);
    setEditing(true);
  }, [layoutQuery.data]);

  const cancelEdit = useCallback(() => {
    setDraft(null);
    setSaveError(null);
    setEditing(false);
    setPickerOpen(false);
  }, []);

  const handleDone = useCallback(async () => {
    if (!draft) {
      return;
    }
    setSaveError(null);
    try {
      await saveLayout.mutateAsync(draft);
      setEditing(false);
      setDraft(null);
      setPickerOpen(false);
    } catch {
      setSaveError("Unable to save layout. Check for overlapping widgets.");
    }
  }, [draft, saveLayout]);

  const handleReset = useCallback(async () => {
    setSaveError(null);
    try {
      await resetLayout.mutateAsync();
      setEditing(false);
      setDraft(null);
      setPickerOpen(false);
    } catch {
      setSaveError("Unable to reset dashboard.");
    }
  }, [resetLayout]);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      if (!draft || !event.over) {
        return;
      }
      const widgetId = String(event.active.id);
      const overData = event.over.data.current as { x?: number; y?: number } | undefined;
      if (overData?.x == null || overData?.y == null) {
        return;
      }

      const widget = draft.widgets.find((item) => item.id === widgetId);
      if (!widget) {
        return;
      }

      const moved: WidgetPlacement = { ...widget, x: overData.x, y: overData.y };
      if (hasOverlap(draft.widgets, moved)) {
        return;
      }

      setDraft({
        ...draft,
        widgets: moveWidget(draft.widgets, widgetId, overData.x, overData.y),
      });
    },
    [draft],
  );

  const handleAddWidget = useCallback(
    (type: WidgetType) => {
      if (!draft) {
        return;
      }
      const definition = getWidgetDefinition(type);
      setDraft({
        ...draft,
        widgets: addWidget(draft.widgets, type, definition.defaultW, definition.defaultH),
      });
      setPickerOpen(false);
    },
    [draft],
  );

  const handleRemoveWidget = useCallback(
    (widgetId: string) => {
      if (!draft) {
        return;
      }
      const next = removeWidget(draft.widgets, widgetId);
      if (next.length === 0) {
        setSaveError("Keep at least one widget on your dashboard.");
        return;
      }
      setSaveError(null);
      setDraft({ ...draft, widgets: next });
    },
    [draft],
  );

  if (layoutQuery.isLoading) {
    return (
      <>
        <Header title="Dashboard" description="Your personal cockpit." />
        <section className="px-8 py-6">
          <p className="text-sm text-muted">Loading dashboard…</p>
        </section>
      </>
    );
  }

  if (layoutQuery.isError || !layoutQuery.data || !activeLayout) {
    return (
      <>
        <Header title="Dashboard" description="Your personal cockpit." />
        <section className="px-8 py-6">
          <Card className="p-6 text-sm text-accent-rose">Unable to load dashboard layout.</Card>
        </section>
      </>
    );
  }

  const grid = (
    <DashboardGrid
      widgets={widgets}
      editing={editing}
      onRemoveWidget={editing ? handleRemoveWidget : undefined}
      showDropCells={editing}
      maxRows={Math.max(maxRows, 6)}
    />
  );

  return (
    <>
      <Header
        title="Dashboard"
        description={editing ? "Customize dashboard" : "Your personal cockpit."}
        action={
          editing ? (
            <DashboardEditBar
              onDone={() => void handleDone()}
              onCancel={cancelEdit}
              onReset={() => void handleReset()}
              onAddWidget={() => setPickerOpen(true)}
              saving={saveLayout.isPending}
              resetting={resetLayout.isPending}
            />
          ) : (
            <Button variant="outline" onClick={enterEditMode}>
              Customize
            </Button>
          )
        }
      />
      <section className="px-8 py-6">
        {saveError ? (
          <Card className="mb-4 p-4 text-sm text-accent-rose">{saveError}</Card>
        ) : null}
        {editing ? (
          <DndContext sensors={sensors} onDragEnd={handleDragEnd}>
            {grid}
          </DndContext>
        ) : (
          grid
        )}
      </section>
      {pickerOpen && layoutQuery.data ? (
        <WidgetPicker
          catalog={layoutQuery.data.catalog}
          existingTypes={existingTypes}
          onSelect={handleAddWidget}
          onClose={() => setPickerOpen(false)}
        />
      ) : null}
    </>
  );
}
