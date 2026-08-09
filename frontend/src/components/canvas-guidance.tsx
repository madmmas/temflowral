"use client";

import {
  AUTHORING_TIP_TEXT,
  EMPTY_CANVAS_BODY,
  EMPTY_CANVAS_HEADING,
} from "@/lib/canvas-guidance";

type EmptyCanvasGuideProps = {
  visible: boolean;
};

/** Centered first-run empty state when the canvas has no nodes (#109). */
export function EmptyCanvasGuide({ visible }: EmptyCanvasGuideProps) {
  if (!visible) return null;

  return (
    <div
      data-testid="empty-canvas-guide"
      role="status"
      className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center p-6"
    >
      <div className="max-w-sm rounded-lg border border-black/10 bg-white/90 px-5 py-4 text-center shadow-sm dark:border-white/15 dark:bg-neutral-900/90">
        <p className="text-sm font-semibold text-black/80 dark:text-white/80">
          {EMPTY_CANVAS_HEADING}
        </p>
        <p className="mt-1.5 text-xs leading-relaxed text-black/55 dark:text-white/55">
          {EMPTY_CANVAS_BODY}
        </p>
      </div>
    </div>
  );
}

type AuthoringTipProps = {
  visible: boolean;
  onDismiss: () => void;
};

/** Dismissible floating authoring tip; preference lives in localStorage (#109). */
export function AuthoringTip({ visible, onDismiss }: AuthoringTipProps) {
  if (!visible) return null;

  return (
    <div
      data-testid="authoring-tip"
      className="flex max-w-xs items-start gap-2 rounded-md bg-white/80 px-2 py-1.5 text-xs text-black/60 shadow-sm dark:bg-neutral-900/80 dark:text-white/60"
    >
      <p className="min-w-0 flex-1 leading-snug">{AUTHORING_TIP_TEXT}</p>
      <button
        type="button"
        data-testid="authoring-tip-dismiss"
        aria-label="Dismiss authoring tip"
        onClick={onDismiss}
        className="shrink-0 rounded px-1.5 py-0.5 text-[10px] font-medium text-black/50 hover:bg-black/5 hover:text-black/80 dark:text-white/50 dark:hover:bg-white/10 dark:hover:text-white/80"
      >
        Dismiss
      </button>
    </div>
  );
}
