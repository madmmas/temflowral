"use client";

import { useMemo, useState, type ChangeEvent } from "react";

import { filterTemplateSuggestions } from "@/lib/config-templates";
import { CONFIG_TEMPLATE_HINT } from "@/lib/graph-canvas";

type TemplateStringInputProps = {
  name: string;
  value: string;
  multiline?: boolean;
  required?: boolean;
  suggestions: readonly string[];
  onChange: (raw: string) => void;
};

/**
 * String config input with upstream template suggestions when typing `{{` (#112).
 */
export function TemplateStringInput({
  name,
  value,
  multiline = false,
  required = false,
  suggestions,
  onChange,
}: TemplateStringInputProps) {
  const [caret, setCaret] = useState(value.length);
  const [open, setOpen] = useState(false);

  const filtered = useMemo(
    () => filterTemplateSuggestions(suggestions, value, caret),
    [suggestions, value, caret],
  );

  const showMenu = open && filtered.length > 0;

  const applySuggestion = (suggestion: string) => {
    const before = value.slice(0, caret);
    const after = value.slice(caret);
    const openIdx = before.lastIndexOf("{{");
    if (openIdx < 0) return;
    const next = `${before.slice(0, openIdx)}${suggestion}${after}`;
    onChange(next);
    setOpen(false);
  };

  const onInput = (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const next = event.target.value;
    const nextCaret = event.target.selectionStart ?? next.length;
    onChange(next);
    setCaret(nextCaret);
    setOpen(true);
  };

  const commonClass =
    "rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900";

  return (
    <div className="relative flex flex-col gap-1 text-xs">
      <span className="font-medium text-black/70 dark:text-white/70">
        {name}
        {required ? " *" : ""}
      </span>
      {multiline ? (
        <textarea
          aria-label={name}
          rows={3}
          value={value}
          onChange={onInput}
          onSelect={(event) =>
            setCaret(event.currentTarget.selectionStart ?? value.length)
          }
          onFocus={() => setOpen(true)}
          onBlur={() => {
            // Allow click on a suggestion before closing.
            window.setTimeout(() => setOpen(false), 120);
          }}
          onKeyDown={(event) => {
            if (event.key === "Escape" && showMenu) {
              event.preventDefault();
              event.stopPropagation();
              setOpen(false);
            }
          }}
          placeholder={CONFIG_TEMPLATE_HINT}
          spellCheck={false}
          className={`${commonClass} font-mono text-xs`}
        />
      ) : (
        <input
          aria-label={name}
          type="text"
          value={value}
          onChange={onInput}
          onSelect={(event) =>
            setCaret(event.currentTarget.selectionStart ?? value.length)
          }
          onFocus={() => setOpen(true)}
          onBlur={() => {
            window.setTimeout(() => setOpen(false), 120);
          }}
          onKeyDown={(event) => {
            if (event.key === "Escape" && showMenu) {
              event.preventDefault();
              event.stopPropagation();
              setOpen(false);
            }
          }}
          placeholder={CONFIG_TEMPLATE_HINT}
          className={commonClass}
        />
      )}
      <span className="text-[10px] text-black/40 dark:text-white/40">
        Templates: {CONFIG_TEMPLATE_HINT}
        {suggestions.length > 0 ? " — type {{ for upstream suggestions" : ""}
      </span>
      {showMenu && (
        <ul
          data-testid="template-suggestions"
          className="absolute left-0 right-0 top-full z-30 mt-0.5 max-h-40 overflow-auto rounded-md border border-black/10 bg-white py-1 shadow-md dark:border-white/15 dark:bg-neutral-900"
        >
          {filtered.map((suggestion) => (
            <li key={suggestion}>
              <button
                type="button"
                data-testid="template-suggestion"
                className="block w-full truncate px-2 py-1 text-left font-mono text-[11px] hover:bg-black/5 dark:hover:bg-white/10"
                onMouseDown={(event) => {
                  event.preventDefault();
                  applySuggestion(suggestion);
                }}
              >
                {suggestion}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
