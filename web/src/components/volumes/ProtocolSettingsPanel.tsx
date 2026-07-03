import { useState } from "react";
import { Card } from "@/components/ui/card";
import { useSettings } from "@/hooks/useSettings";
import { useVolumeProtocols, useUpdateVolumeProtocols } from "@/hooks/useVolumeProtocols";
import type { ProtocolToggle } from "@/lib/api";

function ToggleSwitch({
  label,
  enabled,
  disabled,
  onChange,
}: {
  label: string;
  enabled: boolean;
  disabled?: boolean;
  onChange: (value: boolean) => void;
}) {
  return (
    <label className="inline-flex cursor-pointer items-center gap-3 text-sm text-body">
      <span>{label}</span>
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

function CopyField({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopied(false);
    }
  };

  return (
    <div className="space-y-1">
      <p className="text-xs font-semibold uppercase tracking-wide text-muted">{label}</p>
      <div className="flex items-center gap-2">
        <code className="flex-1 truncate rounded-md border border-hairline bg-surface-soft px-3 py-2 text-xs text-body">
          {value}
        </code>
        <button
          type="button"
          onClick={() => void handleCopy()}
          className="shrink-0 rounded-md border border-hairline px-3 py-2 text-xs font-semibold text-body hover:border-hairline-strong hover:text-ink"
        >
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
    </div>
  );
}

type ProtocolSettingsPanelProps = {
  volumeId: string;
};

export function ProtocolSettingsPanel({ volumeId }: ProtocolSettingsPanelProps) {
  const settingsQuery = useSettings();
  const protocolsQuery = useVolumeProtocols(volumeId);
  const updateProtocols = useUpdateVolumeProtocols(volumeId);

  const instanceWebDAV = settingsQuery.data?.protocols_webdav_enabled ?? false;
  const instanceFTP = settingsQuery.data?.protocols_ftp_enabled ?? false;

  const protocols = protocolsQuery.data;
  const pending = updateProtocols.isPending;

  const patchToggle = (key: "webdav" | "ftp", enabled: boolean) => {
    const payload: { webdav?: ProtocolToggle; ftp?: ProtocolToggle } = {
      [key]: { enabled },
    };
    void updateProtocols.mutateAsync(payload);
  };

  if (protocolsQuery.isLoading) {
    return <p className="text-sm text-muted">Loading protocol settings...</p>;
  }

  if (protocolsQuery.isError || !protocols) {
    return <p className="text-sm text-accent-rose">Unable to load protocol settings.</p>;
  }

  return (
    <Card className="mt-8 space-y-6 p-6">
      <div>
        <h2 className="text-sm font-semibold uppercase tracking-wide text-muted">
          Native access (WebDAV & FTP)
        </h2>
        <p className="mt-2 text-sm text-body">
          Mount this volume in Finder, Explorer, or any FTP client. Both protocols are off by
          default — enable them here once an administrator has turned on instance-wide access.
        </p>
      </div>

      {!instanceWebDAV && !instanceFTP ? (
        <p className="rounded-md border border-hairline bg-surface-soft px-4 py-3 text-sm text-body">
          WebDAV and FTP are disabled for this lcloud instance. Ask an administrator to enable them
          in Settings.
        </p>
      ) : null}

      <div className="grid gap-4 md:grid-cols-2">
        <div className="space-y-3 rounded-md border border-hairline p-4">
          <ToggleSwitch
            label="WebDAV"
            enabled={protocols.webdav.enabled}
            disabled={!instanceWebDAV || pending}
            onChange={(value) => patchToggle("webdav", value)}
          />
          {!instanceWebDAV ? (
            <p className="text-xs text-muted">Disabled instance-wide by admin.</p>
          ) : null}
        </div>
        <div className="space-y-3 rounded-md border border-hairline p-4">
          <ToggleSwitch
            label="FTP"
            enabled={protocols.ftp.enabled}
            disabled={!instanceFTP || pending}
            onChange={(value) => patchToggle("ftp", value)}
          />
          {!instanceFTP ? (
            <p className="text-xs text-muted">Disabled instance-wide by admin.</p>
          ) : null}
        </div>
      </div>

      {protocols.webdav.enabled || protocols.ftp.enabled ? (
        <div className="space-y-4">
          {protocols.webdav.enabled ? (
            <CopyField label="WebDAV URL" value={protocols.connection.webdav_url} />
          ) : null}
          {protocols.ftp.enabled ? (
            <>
              <CopyField
                label="FTP host:port"
                value={`${protocols.connection.ftp_host}:${protocols.connection.ftp_port}`}
              />
              <CopyField label="FTP username (volume UUID)" value={protocols.connection.ftp_username} />
            </>
          ) : null}
          <p className="text-xs text-muted">
            WebDAV username: {protocols.connection.webdav_username_hint}. Password: your lcloud
            account password.
          </p>
        </div>
      ) : null}

      <p className="rounded-md border border-warning/30 bg-surface-soft px-4 py-3 text-xs text-body">
        <span className="font-semibold text-warning">Security note:</span> FTP sends credentials in
        plain text. Prefer WebDAV behind HTTPS when possible.
      </p>
    </Card>
  );
}
