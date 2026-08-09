"use client";

type CanvasErrorBannerProps = {
  message: string;
  onDismiss: () => void;
  onNew?: () => void;
};

/**
 * Dismissible top banner for open/save/run/list failures (#108).
 * Keeps `role="alert"` so errors stay obvious without hunting the footer.
 */
export function CanvasErrorBanner({
  message,
  onDismiss,
  onNew,
}: CanvasErrorBannerProps) {
  return (
    <div
      role="alert"
      data-testid="canvas-error-banner"
      className="flex items-start gap-3 border-b border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-800 dark:text-red-200"
    >
      <p className="min-w-0 flex-1 leading-snug" data-testid="canvas-error-message">
        {message}
      </p>
      <div className="flex shrink-0 items-center gap-1.5">
        {onNew && (
          <button
            type="button"
            data-testid="canvas-error-new"
            onClick={onNew}
            className="rounded-md border border-red-500/30 bg-white/70 px-2 py-1 text-xs font-medium hover:bg-white dark:bg-neutral-900/70 dark:hover:bg-neutral-900"
          >
            New
          </button>
        )}
        <button
          type="button"
          data-testid="canvas-error-dismiss"
          onClick={onDismiss}
          className="rounded-md border border-red-500/30 bg-white/70 px-2 py-1 text-xs font-medium hover:bg-white dark:bg-neutral-900/70 dark:hover:bg-neutral-900"
        >
          Dismiss
        </button>
      </div>
    </div>
  );
}
