import { describe, expect, it } from "vitest";

import {
  headersObjectToRows,
  headersRowsToObject,
  resetHeaderRowSeqForTests,
} from "@/lib/headers-editor";

describe("headers rows ↔ object", () => {
  it("round-trips non-empty headers", () => {
    resetHeaderRowSeqForTests();
    const rows = headersObjectToRows({ Authorization: "Bearer x", Accept: "json" });
    expect(rows).toHaveLength(2);
    expect(headersRowsToObject(rows)).toEqual({
      Authorization: "Bearer x",
      Accept: "json",
    });
  });

  it("omits blank keys and yields undefined when empty", () => {
    expect(
      headersRowsToObject([
        { id: "1", key: "  ", value: "x" },
        { id: "2", key: "", value: "" },
      ]),
    ).toBeUndefined();
  });

  it("starts with one blank row for empty/missing values", () => {
    resetHeaderRowSeqForTests();
    expect(headersObjectToRows(undefined)).toEqual([
      { id: "hdr-1", key: "", value: "" },
    ]);
  });
});
