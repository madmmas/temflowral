"use client";

import { useEffect, useState } from "react";

import {
  headersObjectToRows,
  headersRowsToObject,
  nextHeaderRowId,
  type HeaderRow,
} from "@/lib/headers-editor";

type HeadersFieldEditorProps = {
  name: string;
  value: unknown;
  syncKey: string;
  required?: boolean;
  onChange: (next: Record<string, string> | undefined) => void;
};

/**
 * Key/value row editor for string-map config fields (HTTP headers) (#112).
 */
export function HeadersFieldEditor({
  name,
  value,
  syncKey,
  required = false,
  onChange,
}: HeadersFieldEditorProps) {
  const [rows, setRows] = useState<HeaderRow[]>(() => headersObjectToRows(value));

  useEffect(() => {
    setRows(headersObjectToRows(value));
    // Re-hydrate when switching nodes/fields — not on every keystroke commit.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- syncKey gates external resets
  }, [syncKey]);

  const commit = (nextRows: HeaderRow[]) => {
    setRows(nextRows);
    onChange(headersRowsToObject(nextRows));
  };

  return (
    <div data-testid="headers-field-editor" className="flex flex-col gap-1.5 text-xs">
      <span className="font-medium text-black/70 dark:text-white/70">
        {name}
        {required ? " *" : ""}
      </span>
      <div className="flex flex-col gap-1.5">
        {rows.map((row, index) => (
          <div key={row.id} className="flex items-center gap-1.5">
            <input
              aria-label={`${name} key ${index + 1}`}
              data-testid={`header-key-${index}`}
              value={row.key}
              placeholder="Header"
              onChange={(event) => {
                const next = rows.map((entry) =>
                  entry.id === row.id
                    ? { ...entry, key: event.target.value }
                    : entry,
                );
                commit(next);
              }}
              className="min-w-0 flex-1 rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
            />
            <input
              aria-label={`${name} value ${index + 1}`}
              data-testid={`header-value-${index}`}
              value={row.value}
              placeholder="Value"
              onChange={(event) => {
                const next = rows.map((entry) =>
                  entry.id === row.id
                    ? { ...entry, value: event.target.value }
                    : entry,
                );
                commit(next);
              }}
              className="min-w-0 flex-[1.4] rounded-md border border-black/10 bg-white px-2 py-1.5 font-mono text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
            />
            <button
              type="button"
              data-testid={`header-remove-${index}`}
              aria-label={`Remove ${name} row ${index + 1}`}
              onClick={() => {
                const next = rows.filter((entry) => entry.id !== row.id);
                commit(
                  next.length > 0
                    ? next
                    : [{ id: nextHeaderRowId(), key: "", value: "" }],
                );
              }}
              className="shrink-0 rounded px-1.5 py-1 text-[10px] text-black/50 hover:bg-black/5 dark:text-white/50 dark:hover:bg-white/10"
            >
              Remove
            </button>
          </div>
        ))}
      </div>
      <button
        type="button"
        data-testid="header-add-row"
        onClick={() =>
          commit([...rows, { id: nextHeaderRowId(), key: "", value: "" }])
        }
        className="self-start rounded-md border border-black/10 px-2 py-1 text-[11px] font-medium text-black/60 hover:bg-black/5 dark:border-white/15 dark:text-white/60 dark:hover:bg-white/10"
      >
        Add header
      </button>
    </div>
  );
}
