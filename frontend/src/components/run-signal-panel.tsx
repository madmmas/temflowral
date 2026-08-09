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
 * Dedicated wait/signal drawer shown while GET /runs/{id} reports
 * `currentWait` (#110). Kept out of the crowded footer status strip.
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
    <section
      data-testid="run-signal-panel"
      aria-label="Wait for signal"
      className="flex shrink-0 flex-col gap-3 border-t border-amber-500/35 bg-amber-50/90 px-4 py-3 dark:border-amber-400/30 dark:bg-amber-950/50"
    >
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <div>
          <h2 className="text-xs font-semibold uppercase tracking-wide text-amber-900 dark:text-amber-100">
            Waiting for signal
          </h2>
          <p className="mt-0.5 text-xs text-amber-900/75 dark:text-amber-100/70">
            Run is blocked on node{" "}
            <span className="font-mono">{currentWait.nodeId}</span>
            {currentWait.signal ? (
              <>
                {" "}
                · expected{" "}
                <span className="font-mono">{currentWait.signal}</span>
              </>
            ) : null}
          </p>
        </div>
      </div>

      <div className="flex flex-wrap items-end gap-3">
        <label className="flex min-w-40 flex-col gap-1 text-xs text-amber-950/80 dark:text-amber-50/80">
          Signal name
          <input
            aria-label="Signal name"
            data-testid="signal-name"
            value={signal}
            onChange={(event) => setSignal(event.target.value)}
            disabled={busy || sending}
            className="rounded-md border border-amber-700/25 bg-white px-2.5 py-1.5 text-sm outline-none focus:ring-2 focus:ring-amber-600/40 dark:border-amber-300/20 dark:bg-neutral-900"
          />
        </label>
        <label className="flex min-w-56 flex-1 flex-col gap-1 text-xs text-amber-950/80 dark:text-amber-50/80">
          Payload (JSON, optional)
          <input
            aria-label="Signal payload"
            data-testid="signal-payload"
            value={payloadText}
            onChange={(event) => setPayloadText(event.target.value)}
            placeholder='{"approved":true}'
            disabled={busy || sending}
            className="rounded-md border border-amber-700/25 bg-white px-2.5 py-1.5 font-mono text-sm outline-none focus:ring-2 focus:ring-amber-600/40 dark:border-amber-300/20 dark:bg-neutral-900"
          />
        </label>
        <button
          type="button"
          data-testid="send-signal"
          onClick={() => void submit()}
          disabled={busy || sending || !signal.trim()}
          className="rounded-md bg-amber-800 px-4 py-2 text-sm font-medium text-white hover:bg-amber-900 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-amber-600 dark:hover:bg-amber-500"
        >
          {sending ? "Sending…" : "Send signal"}
        </button>
      </div>

      {localError && (
        <p role="alert" className="text-xs text-red-700 dark:text-red-300">
          {localError}
        </p>
      )}
    </section>
  );
}
