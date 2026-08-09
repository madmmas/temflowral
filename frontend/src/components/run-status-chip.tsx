"use client";

import {
  formatRunStatusLabel,
  runStatusTone,
  type RunStatus,
} from "@/lib/run-history";

type RunStatusChipProps = {
  status: RunStatus;
  waiting?: boolean;
  canOpenResult: boolean;
  resultPanelOpen: boolean;
  onOpenResult?: () => void;
};

const toneClass: Record<
  ReturnType<typeof runStatusTone> | "warning",
  string
> = {
  success:
    "border-green-500/40 bg-green-500/10 text-green-700 dark:text-green-300",
  danger: "border-red-500/40 bg-red-500/10 text-red-700 dark:text-red-300",
  warning:
    "border-amber-500/40 bg-amber-500/10 text-amber-800 dark:text-amber-200",
  info: "border-blue-500/40 bg-blue-500/10 text-blue-700 dark:text-blue-300",
};

/**
 * Compact run status chip for the footer (#106). Keeps `data-testid="run-status"`
 * and the `Run {status}` label for existing e2e.
 */
export function RunStatusChip({
  status,
  waiting = false,
  canOpenResult,
  resultPanelOpen,
  onOpenResult,
}: RunStatusChipProps) {
  const tone = waiting ? "warning" : runStatusTone(status);
  const label = waiting ? "Run waiting" : formatRunStatusLabel(status);
  const className = `inline-flex items-center rounded-full border px-2 py-0.5 text-[11px] font-medium ${toneClass[tone]}`;

  if (canOpenResult && onOpenResult) {
    return (
      <button
        type="button"
        data-testid="run-status"
        onClick={onOpenResult}
        title={
          resultPanelOpen ? "Run result panel is open" : "Show run result"
        }
        className={`${className} hover:underline`}
      >
        {label}
        {!resultPanelOpen && " · View result"}
      </button>
    );
  }

  return (
    <span data-testid="run-status" className={className}>
      {label}
    </span>
  );
}
