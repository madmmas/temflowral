"use client";

import { useEffect, useState } from "react";

import type { components } from "@/api";

type CurrentWait = components["schemas"]["CurrentWait"];

type RunSignalPanelProps = {
  runId: string;
  currentWait: CurrentWait;
  busy: boolean;
  onSend: (signal: string, payload: unknown | undefined) => Promise<void>;
};

/**
 * Shown while GET /runs/{id} reports `currentWait`. Sends
 * POST /runs/{id}/signal with the wait node's signal name.
 */
export function RunSignalPanel({
  runId,
  currentWait,
  busy,
  onSend,
}: RunSignalPanelProps) {
  const [signal, setSignal] = useState(currentWait.signal);
  const [payloadText, setPayloadText] = useState("");
  const [localError, setLocalError] = useState<string | null>(null);
  const [sending, setSending] = useState(false);

  useEffect(() => {
    setSignal(currentWait.signal);
    setLocalError(null);
  }, [currentWait.signal, currentWait.nodeId, runId]);

  const submit = async () => {
    setLocalError(null);
    let payload: unknown | undefined;
    const trimmed = payloadText.trim();
    if (trimmed) {
      try {
        payload = JSON.parse(trimmed) as unknown;
      } catch {
        setLocalError("Payload must be valid JSON");
        return;
      }
    }

    setSending(true);
    try {
      await onSend(signal.trim(), payload);
    } catch (error) {
      setLocalError(
        error instanceof Error ? error.message : "Failed to send signal",
      );
    } finally {
      setSending(false);
    }
  };

  return (
    <div
      data-testid="run-signal-panel"
      className="flex w-full flex-wrap items-end gap-2 rounded-md border border-amber-500/40 bg-amber-50/80 px-2 py-2 dark:border-amber-400/30 dark:bg-amber-950/40"
    >
      <div className="min-w-40 flex-1">
        <p className="text-[10px] font-medium uppercase tracking-wide text-amber-800/80 dark:text-amber-200/80">
          Waiting on signal
        </p>
        <p className="text-[11px] text-amber-900/70 dark:text-amber-100/70">
          Node <span className="font-mono">{currentWait.nodeId}</span>
        </p>
      </div>
      <label className="flex min-w-36 flex-col gap-0.5 text-[10px] text-black/60 dark:text-white/60">
        Signal
        <input
          aria-label="Signal name"
          data-testid="signal-name"
          value={signal}
          onChange={(event) => setSignal(event.target.value)}
          disabled={busy || sending}
          className="rounded border border-black/10 bg-white px-2 py-1 text-xs dark:border-white/15 dark:bg-neutral-900"
        />
      </label>
      <label className="flex min-w-40 flex-1 flex-col gap-0.5 text-[10px] text-black/60 dark:text-white/60">
        Payload (JSON, optional)
        <input
          aria-label="Signal payload"
          data-testid="signal-payload"
          value={payloadText}
          onChange={(event) => setPayloadText(event.target.value)}
          placeholder='{"approved":true}'
          disabled={busy || sending}
          className="rounded border border-black/10 bg-white px-2 py-1 font-mono text-xs dark:border-white/15 dark:bg-neutral-900"
        />
      </label>
      <button
        type="button"
        data-testid="send-signal"
        onClick={() => void submit()}
        disabled={busy || sending || !signal.trim()}
        className="rounded-md bg-amber-700 px-3 py-1.5 text-xs font-medium text-white hover:bg-amber-800 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {sending ? "Sending…" : "Send signal"}
      </button>
      {localError && (
        <span className="w-full text-xs text-red-600 dark:text-red-400">
          {localError}
        </span>
      )}
    </div>
  );
}
