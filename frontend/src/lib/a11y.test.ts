import { describe, expect, it } from "vitest";

import { isTypingTarget, resolveEscapeLayer } from "@/lib/a11y";

describe("resolveEscapeLayer", () => {
  it("prefers library, then config, then result", () => {
    expect(
      resolveEscapeLayer({
        libraryOpen: true,
        configOpen: true,
        resultOpen: true,
      }),
    ).toBe("library");
    expect(
      resolveEscapeLayer({
        libraryOpen: false,
        configOpen: true,
        resultOpen: true,
      }),
    ).toBe("config");
    expect(
      resolveEscapeLayer({
        libraryOpen: false,
        configOpen: false,
        resultOpen: true,
      }),
    ).toBe("result");
    expect(
      resolveEscapeLayer({
        libraryOpen: false,
        configOpen: false,
        resultOpen: false,
      }),
    ).toBe("none");
  });
});

describe("isTypingTarget", () => {
  it("detects form-like targets without a DOM", () => {
    expect(isTypingTarget(null)).toBe(false);
    expect(isTypingTarget({ tagName: "INPUT" } as unknown as EventTarget)).toBe(
      true,
    );
    expect(isTypingTarget({ tagName: "DIV" } as unknown as EventTarget)).toBe(
      false,
    );
    expect(
      isTypingTarget({
        tagName: "DIV",
        isContentEditable: true,
      } as unknown as EventTarget),
    ).toBe(true);
  });
});
