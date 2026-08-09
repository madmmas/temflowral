/**
 * Helpers for HTTP-style string-map headers config (#112).
 * Rows edit as key/value pairs and serialize to `Record<string, string>`.
 */

export type HeaderRow = { id: string; key: string; value: string };

let headerRowSeq = 0;

export function nextHeaderRowId(): string {
  headerRowSeq += 1;
  return `hdr-${headerRowSeq}`;
}

/** Reset id sequence (tests only). */
export function resetHeaderRowSeqForTests(): void {
  headerRowSeq = 0;
}

export function headersObjectToRows(
  value: unknown,
  newId: () => string = nextHeaderRowId,
): HeaderRow[] {
  if (value === undefined || value === null) {
    return [{ id: newId(), key: "", value: "" }];
  }
  if (typeof value !== "object" || Array.isArray(value)) {
    return [{ id: newId(), key: "", value: "" }];
  }
  const entries = Object.entries(value as Record<string, unknown>);
  if (entries.length === 0) {
    return [{ id: newId(), key: "", value: "" }];
  }
  return entries.map(([key, raw]) => ({
    id: newId(),
    key,
    value: raw == null ? "" : String(raw),
  }));
}

/** Drop blank keys; last duplicate key wins. */
export function headersRowsToObject(
  rows: readonly HeaderRow[],
): Record<string, string> | undefined {
  const next: Record<string, string> = {};
  let any = false;
  for (const row of rows) {
    const key = row.key.trim();
    if (!key) continue;
    next[key] = row.value;
    any = true;
  }
  return any ? next : undefined;
}

export function isStringMapConfigValue(value: unknown): boolean {
  if (value === undefined || value === null) return true;
  if (typeof value !== "object" || Array.isArray(value)) return false;
  return Object.values(value as Record<string, unknown>).every(
    (entry) => typeof entry === "string",
  );
}
