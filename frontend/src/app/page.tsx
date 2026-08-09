import { GraphCanvas } from "@/components/graph-canvas";

export default function Home() {
  return (
    <main className="flex h-screen flex-col">
      <a
        href="#graph-canvas-main"
        className="sr-only focus:not-sr-only focus:absolute focus:left-3 focus:top-3 focus:z-50 focus:rounded-md focus:bg-white focus:px-3 focus:py-2 focus:text-sm focus:font-medium focus:shadow focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:bg-neutral-900"
      >
        Skip to canvas
      </a>
      <header className="flex items-center gap-3 border-b border-black/10 px-4 py-2 dark:border-white/15">
        <h1 className="font-sans text-lg font-semibold tracking-tight">
          temflowral
        </h1>
        <span className="text-xs text-black/55 dark:text-white/55">
          workflow canvas
        </span>
      </header>
      <div className="min-h-0 flex-1">
        <GraphCanvas />
      </div>
    </main>
  );
}
