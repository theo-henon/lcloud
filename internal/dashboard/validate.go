package dashboard

import (
	"fmt"

	"github.com/theo-henon/lcloud/internal/auth"
)

const (
	minWidgets = 1
	maxWidgets = 20
	gridCols   = 12
	maxHeight  = 4
)

func ValidateLayout(layout Layout, role auth.Role) error {
	if layout.Version != LayoutVersion {
		return fmt.Errorf("%w: unsupported version", ErrInvalidLayout)
	}
	if len(layout.Widgets) < minWidgets || len(layout.Widgets) > maxWidgets {
		return fmt.Errorf("%w: widget count must be between %d and %d", ErrInvalidLayout, minWidgets, maxWidgets)
	}

	seenIDs := make(map[string]struct{}, len(layout.Widgets))
	for _, widget := range layout.Widgets {
		if widget.ID == "" {
			return fmt.Errorf("%w: widget id required", ErrInvalidLayout)
		}
		if _, exists := seenIDs[widget.ID]; exists {
			return fmt.Errorf("%w: duplicate widget id %q", ErrInvalidLayout, widget.ID)
		}
		seenIDs[widget.ID] = struct{}{}

		entry, ok := CatalogEntryForType(widget.Type)
		if !ok {
			return fmt.Errorf("%w: %q", ErrUnknownWidget, widget.Type)
		}
		if entry.AdminOnly && role != auth.RoleAdmin {
			return fmt.Errorf("%w: %q", ErrWidgetNotAllowed, widget.Type)
		}
		if widget.W != entry.DefaultW || widget.H != entry.DefaultH {
			return fmt.Errorf("%w: widget %q size must be %dx%d", ErrInvalidLayout, widget.Type, entry.DefaultW, entry.DefaultH)
		}
		if widget.X < 0 || widget.Y < 0 {
			return fmt.Errorf("%w: negative coordinates for widget %q", ErrInvalidLayout, widget.ID)
		}
		if widget.W < 1 || widget.W > gridCols || widget.X+widget.W > gridCols {
			return fmt.Errorf("%w: widget %q out of horizontal bounds", ErrInvalidLayout, widget.ID)
		}
		if widget.H < 1 || widget.H > maxHeight {
			return fmt.Errorf("%w: widget %q height out of bounds", ErrInvalidLayout, widget.ID)
		}
	}

	for i := 0; i < len(layout.Widgets); i++ {
		for j := i + 1; j < len(layout.Widgets); j++ {
			if widgetsOverlap(layout.Widgets[i], layout.Widgets[j]) {
				return fmt.Errorf("%w: widgets %q and %q overlap", ErrInvalidLayout, layout.Widgets[i].ID, layout.Widgets[j].ID)
			}
		}
	}

	return nil
}

func widgetsOverlap(a, b WidgetPlacement) bool {
	return a.X < b.X+b.W && a.X+a.W > b.X && a.Y < b.Y+b.H && a.Y+a.H > b.Y
}
