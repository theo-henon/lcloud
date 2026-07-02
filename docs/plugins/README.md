# lcloud Plugin SDK

Phase 1.3 introduces external plugin binaries that react to volume events via HashiCorp go-plugin.

## Quick start

1. Implement a handler:

```go
package main

import (
    plugincore "github.com/theo-henon/lcloud/internal/plugin"
    "github.com/theo-henon/lcloud/pkg/pluginsdk"
)

type myPlugin struct{}

func (p *myPlugin) HandleEvent(event plugincore.Event) (*plugincore.HandleResult, error) {
    return &plugincore.HandleResult{}, nil
}

func main() {
    pluginsdk.Main(&myPlugin{})
}
```

2. Add a manifest sidecar next to the binary (`my-plugin.json`):

```json
{
  "id": "my-plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "description": "Does something useful",
  "author": "you",
  "subscribe": ["file.uploaded"],
  "emit": ["my.custom.event"]
}
```

3. Build and install:

```bash
go build -o plugins/my-plugin ./path/to/plugin
cp my-plugin.json plugins/my-plugin.json
```

4. Restart lcloud — the plugin appears in **Plugins** manager UI.

## Example: file-type-validator

```bash
go build -o plugins/file-type-validator ./examples/file-type-validator
cp examples/file-type-validator/file-type-validator.json plugins/file-type-validator.json
docker compose restart app
```

### UC3.1 — plugin running

Open `/plugins` — **File Type Validator** shows status **running**.

### UC3.2 — validation notice

Upload a `.png` to a volume that allows `.png`. The activity feed shows `validation passed for photo.png`.

### Manual edge case: validation.failed

Core upload filter rejects disallowed extensions before save. To demo `validation.failed`:

1. Upload an allowed file, then change volume filters to exclude its extension, **or**
2. Copy a disallowed file directly into `{volume}/userdata/` and trigger a manual re-validation (post-MVP).

## Host API (MVP)

Plugins return logs and custom events via `HandleResult` — there is no reverse RPC channel from plugin to host in Phase 1.3:

```go
return &plugincore.HandleResult{
    Logs: []plugincore.LogRequest{{Message: "done", Level: plugincore.LogLevelInfo}},
    Emit: []plugincore.Event{plugincore.NewCustomEvent("my.event", volumeID, payload)},
}, nil
```

The host creates `{volume}/plugins/{plugin-id}/` automatically when dispatching the first event for that volume. Plugins should write state files only under that directory.

## Environment

| Variable | Default | Description |
|---|---|---|
| `PLUGINS_PATH` | `./plugins` | Directory scanned for plugin binaries + manifests |

Docker mounts `./plugins` read-only at `/plugins`.

## Security notes

- Plugins cannot access PostgreSQL directly
- Plugins write only to `{volume}/plugins/{plugin-id}/`
- Enable/disable requires admin role
- Plugin crash does not stop lcloud core
