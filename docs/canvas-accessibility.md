# Reference canvas accessibility (#114)

Baseline for the Next.js reference UI — **not** a full WCAG audit or
compliance claim. Split follow-ups as needed.

## Verified (this baseline)

| Area | What we check |
| --- | --- |
| Keyboard path | Tab reaches palette, graph name, Open… / Save / Run; Enter/Space adds a palette node; Escape closes library → node config → run result (in that order) |
| Focus rings | `:focus-visible` rings on primary buttons and toolbar controls; inputs keep existing blue rings |
| Status / errors | Text labels on run status, execution badges, and `role="alert"` banners — not color alone |
| Delete keys | Backspace/Delete do not remove nodes while focus is in an input / textarea / select |
| Docs | This note lists gaps honestly |

Automated coverage: `frontend/e2e/a11y-baseline.spec.ts` plus unit tests in
`frontend/src/lib/a11y.test.ts`.

## Known gaps (not fixed here)

- **React Flow node/edge navigation** — selecting and connecting nodes is still
  primarily pointer-driven; no full keyboard graph authoring path.
- **Focus restore** — closing overlays does not always return focus to the
  control that opened them.
- **Focus trap** — workflow library is Escape-dismissible and focuses search,
  but is not a full modal focus trap.
- **Screen reader graph description** — custom nodes expose labels visually;
  there is no live region summarizing the graph structure.
- **Contrast audit** — dark-theme tokens were spot-checked; muted
  `white/40`–`white/50` helper text may still fail strict AA on some displays.
- **MiniMap / Controls chrome** — relies on React Flow defaults beyond our
  show/hide toggle.

## Manual smoke (optional)

1. Prefer keyboard only: add Start from the palette, name the graph, Run.
2. Open a node config, edit Label, press Escape — panel closes.
3. Complete a run, open the result drawer, Escape — drawer hides.
4. Toggle OS dark mode and confirm toolbar / palette / error banner remain
   readable.
