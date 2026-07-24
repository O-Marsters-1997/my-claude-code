# Open Design — rendering fix directions as low-fi wireframes

Used in stage 5's fan-out to turn each of `impeccable`'s fix directions into a viewable PNG. `impeccable` still owns the ideas; this is only the render/export engine. Optional — degrade to ASCII wireframes if any part of this isn't available (see bottom).

**What the page should look like is not decided here.** Fidelity, framing, labelling and annotation are governed by [wireframe-style.md](wireframe-style.md) — read it before authoring, and follow it over anything below if the two ever drift.

## Binaries

Two binaries on PATH: `od` (the daemon CLI — run bare `od` to start daemon + web UI) and `odo`/`pnpm tools-dev` (starts daemon + web + **desktop** in the background). **Use `odo`** — `od export … image` needs the desktop runtime's bundled Chromium to rasterize.

`/usr/bin/od` (the system octal-dump tool) can shadow the Open Design `od` depending on PATH order. Verify with `which -a od`; if the wrong one wins, call it by absolute path `~/open-design/apps/daemon/bin/od.mjs` or via the `$OD_BIN` env var the daemon injects into spawned agents.

Env vars the daemon injects when it spawns an agent (also usable to make CLI calls robust): `OD_BIN` (abs path to the CLI), `OD_DAEMON_URL` (e.g. `http://127.0.0.1:7456`), `OD_PROJECT_ID`, `OD_PROJECT_DIR`.

## The render loop

Per fix direction:

1. **Ground in the current code.** Before drawing, read the actual source of the audited surface to get real structure/labels as the starting point — the wireframe is a *modification of what exists*, not a blank invention. (E.g. an onboarding run might draw from `onboardingModal.tsx`, `nextStepsCard.tsx`, `enableChannelsForm.tsx`, `threadPageQueueEmptyStates.tsx`, `useNextStepsChecklist.ts` — whatever files the findings named.) Capture the current layout (regions, primary actions, copy) so the "before" is honest and the wireframe reads as *this* product.

2. **Author one self-contained low-fi HTML file per direction** — one direction per file, one screen per frame, never a multi-direction board:
   - Single file, all CSS inline, no external assets (renders deterministically).
   - Everything about how it looks — greyscale, placeholders, annotation style, labelling, spacing — comes from [wireframe-style.md](wireframe-style.md). Don't re-decide it here.
   - Optionally seed consistent styling from a built-in direction: `od tools directions --json` lists them (`modern-minimal`, `tech-utility`, etc.); `od tools directions --id <id>` prints palette/fonts/posture to bind into `:root`. For pure wireframes a neutral grayscale is usually better — leave this as an option, not a requirement.

3. **Register it as a project artifact** so it renders in the UI:
   ```bash
   od artifacts create --name <dir>/<direction-id>.html --input ./<local>.html [--project <id-or-name>]
   ```
   Omitting `--project` uses the active project. Existing target paths are rejected. `--encoding utf8|base64` if the HTML needs it.

4. **Export a PNG** for embedding in `findings.md`:
   ```bash
   od export <dir>/<direction-id>.html --project <id> --format image --image-format png --page --out ./<direction-id>.png
   ```
   Save PNGs into the run's artifacts dir (`friction-audit/<flow>-<YYYYMMDD>/`) alongside the snapshots, and reference them from `findings.md` (`![1A — pre-seeded demo thread](1A-demo-thread.png)`) in place of / next to the ASCII block. Keep the ASCII as a text fallback if you like, but the PNG is the artifact.

**Naming:** tie each artifact to the fix-direction label already used in `findings.md` (e.g. `pain1-1A-demo-thread`) so a re-run diffs cleanly, matching the skill's existing stable-ID discipline.

## Graceful degradation

If `od`/`odo` isn't installed, the daemon isn't reachable, or `od export` reports the renderer unavailable (no desktop runtime), fall back to the ASCII wireframes and say so in the report — never hard-fail the audit because a design tool is missing. Mirror how the skill already treats `playwright-cli` as a hard prerequisite but degrades elsewhere.

## Things to verify before relying on this (don't assume)

- **Project bootstrapping headlessly.** `od artifacts create`/`export` need a project (`--project <id-or-name>`, or the active one). There is no `od projects create` subcommand; `list_projects` exists on the `od mcp` surface. Confirm the intended headless path: reuse the active project, create one via the web UI once, or via an API/MCP call.
- **Export prerequisite.** Confirm `od export --format image` works with `odo`/`pnpm tools-dev` running (desktop reachable) and states a clear error otherwise — so the fallback above triggers correctly.
- **`--input` encoding/size** for `artifacts create` (`--encoding utf8|base64`) for typical wireframe HTML.

## Further reading

In `~/open-design`: `QUICKSTART.md` (daemon/desktop startup, `OD_*` env vars, execution modes), `docs/modes.md` (the seven normalized modes), `docs/skills-protocol.md` (functional-skill vs rendering-template split), `AGENTS.md` (daemon data directory contract). Repo: https://github.com/nexu-io/open-design.
