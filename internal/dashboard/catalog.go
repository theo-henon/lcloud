package dashboard

import "github.com/theo-henon/lcloud/internal/auth"

var catalog = []CatalogEntry{
	{
		Type:        "storage-summary",
		Title:       "Storage",
		Description: "Disk space and volume usage summary",
		DefaultW:    6,
		DefaultH:    2,
	},
	{
		Type:        "volumes-at-risk",
		Title:       "Volumes at risk",
		Description: "Volumes approaching quota limits",
		DefaultW:    6,
		DefaultH:    2,
	},
	{
		Type:        "tasks-overview",
		Title:       "Tasks",
		Description: "Scheduled task status overview",
		DefaultW:    6,
		DefaultH:    2,
	},
	{
		Type:        "plugins-status",
		Title:       "Plugins",
		Description: "Plugin runtime status summary",
		DefaultW:    4,
		DefaultH:    2,
	},
	{
		Type:        "quick-actions",
		Title:       "Quick actions",
		Description: "Shortcuts to common pages",
		DefaultW:    4,
		DefaultH:    1,
	},
	{
		Type:        "welcome",
		Title:       "Welcome",
		Description: "Personal greeting",
		DefaultW:    4,
		DefaultH:    1,
	},
	{
		Type:        "pending-deletions",
		Title:       "Pending deletions",
		Description: "Volume deletion requests awaiting admin review",
		DefaultW:    6,
		DefaultH:    2,
		AdminOnly:   true,
	},
}

func CatalogForRole(role auth.Role) []CatalogEntry {
	out := make([]CatalogEntry, 0, len(catalog))
	for _, entry := range catalog {
		if entry.AdminOnly && role != auth.RoleAdmin {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func CatalogEntryForType(widgetType string) (CatalogEntry, bool) {
	for _, entry := range catalog {
		if entry.Type == widgetType {
			return entry, true
		}
	}
	return CatalogEntry{}, false
}

func DefaultLayout(role auth.Role) Layout {
	if role == auth.RoleAdmin {
		return Layout{
			Version: LayoutVersion,
			Widgets: []WidgetPlacement{
				{ID: "default-pending-deletions", Type: "pending-deletions", X: 0, Y: 0, W: 6, H: 2},
				{ID: "default-quick-actions", Type: "quick-actions", X: 6, Y: 0, W: 4, H: 1},
				{ID: "default-plugins-status", Type: "plugins-status", X: 8, Y: 0, W: 4, H: 2},
				{ID: "default-storage-summary", Type: "storage-summary", X: 0, Y: 1, W: 6, H: 2},
				{ID: "default-volumes-at-risk", Type: "volumes-at-risk", X: 6, Y: 2, W: 6, H: 2},
				{ID: "default-tasks-overview", Type: "tasks-overview", X: 0, Y: 3, W: 6, H: 2},
			},
		}
	}

	return Layout{
		Version: LayoutVersion,
		Widgets: []WidgetPlacement{
			{ID: "default-welcome", Type: "welcome", X: 0, Y: 0, W: 4, H: 1},
			{ID: "default-quick-actions", Type: "quick-actions", X: 4, Y: 0, W: 4, H: 1},
			{ID: "default-plugins-status", Type: "plugins-status", X: 8, Y: 0, W: 4, H: 2},
			{ID: "default-storage-summary", Type: "storage-summary", X: 0, Y: 1, W: 6, H: 2},
			{ID: "default-volumes-at-risk", Type: "volumes-at-risk", X: 6, Y: 2, W: 6, H: 2},
			{ID: "default-tasks-overview", Type: "tasks-overview", X: 0, Y: 3, W: 6, H: 2},
		},
	}
}
