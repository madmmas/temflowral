import { GraphCanvas } from "@/components/graph-canvas";

export default function Home() {
  return (
    <main className="flex h-screen flex-col">
      <header className="flex items-center gap-3 border-b border-black/10 px-4 py-2 dark:border-white/15">
        <h1 className="font-sans text-lg font-semibold tracking-tight">
          temflowral
        </h1>
        <span className="text-xs text-black/50 dark:text-white/50">
          workflow canvas
        </span>
      </header>
      <div className="min-h-0 flex-1">
        <GraphCanvas />
      </div>
    </main>
  );
}
