import { describe, expect, it } from "vitest";
import { ALL_WIDGET_TYPES, WIDGET_REGISTRY } from "@/lib/dashboard/registry";

describe("WIDGET_REGISTRY", () => {
  it("maps every catalog widget type to a component", () => {
    for (const type of ALL_WIDGET_TYPES) {
      const entry = WIDGET_REGISTRY[type];
      expect(entry).toBeDefined();
      expect(entry.component).toBeTypeOf("function");
      expect(entry.title.length).toBeGreaterThan(0);
      expect(entry.defaultW).toBeGreaterThan(0);
      expect(entry.defaultH).toBeGreaterThan(0);
    }
  });
});
