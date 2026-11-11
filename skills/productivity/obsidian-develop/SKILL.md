---
name: obsidian-develop
description: >
  Build, prototype, extend, debug and review Obsidian community plugins against a live vault:
  inspect the running app's UI, DOM, console and plugin state, prototype in the runtime, then
  verify each change with a build, reload and check loop. Use whenever the user is working on an
  Obsidian plugin, even when they just say "add a command", "make a sidebar view", "my settings
  tab is blank", "why isn't my view updating" or "get this ready for the community directory";
  when they ask about the plugin API (obsidian.d.ts, Workspace, leaves, Vault, MetadataCache,
  FileManager, views, modals, settings tabs, CodeMirror 6 editor extensions, markdown
  post-processors, events); or when they want a running Obsidian instance's DOM, CSS, console or
  plugins inspected. Not for general note-taking, Obsidian Markdown syntax, themes without a
  plugin, or managing vault content (use ingest, to-vault or vault-recall for those).
---

# Obsidian plugin development

Turn a plain-language request into a plugin change that works in the running app and passes
community review. One loop does it: **spec → explore → prototype → implement → build, reload,
verify → review**. Code written against what is actually in the live app beats code written from
memory of the API, because Obsidian's DOM and private internals move between releases.

`obsidian.d.ts` in `node_modules/obsidian/` is the source of truth for every public signature.
Read it rather than guessing. Docs: <https://docs.obsidian.md>.

## Step 0: find the runtime backend

The loop needs six capabilities. Work out which backend gives them before writing code.

1. **An Obsidian MCP server.** List the connected tools and match on purpose, not exact names.
   Names differ between servers (`obsidian_eval`, `obsidian_execute_js`, …).
2. **The official Obsidian CLI** (installer 1.12.7+, Settings → General → Command line
   interface). Check with `obsidian version` in Bash. It covers every capability itself, and the
   best-maintained MCP server is a thin wrapper around it.
3. **Neither.** Build, then ask the user to reload the plugin and paste the console output (and a
   screenshot if the change is visual). Say which checks you could not run yourself.

| Capability | Typical MCP tool | CLI command |
|---|---|---|
| Eval JS | `*_eval`, `*_execute_js` | `obsidian eval code="…"` |
| Read console and errors | `*_console`, `*_errors` | `obsidian dev:console [level=error] [clear]`, `obsidian dev:errors` |
| Query DOM and CSS | `*_dom`, `*_css`, `*_dom_query` | `obsidian dev:dom selector=… [text] [all]`, `obsidian dev:css selector=…` |
| Screenshot | `*_screenshot` | `obsidian dev:screenshot path=<file>` |
| Reload or enable a plugin | `*_plugin_reload`, `*_enable_plugin` | `obsidian plugin:reload id=<id>` |
| List or run commands | `*_commands`, `*_command` | `obsidian commands filter=<id>`, `obsidian command id=<id>` |

CDP-based servers need Obsidian started with the debug port, after quitting any running instance:
`/Applications/Obsidian.app/Contents/MacOS/Obsidian --remote-debugging-port=9222`. CLI-based
servers need the CLI setting on. Setup for either, plus CLI gotchas: `references/runtime-inspection.md`.

**Safety.** Prefer a dedicated test vault. Exploration evals must be read-only. Ask before any
eval that writes, renames, moves or deletes vault files or changes another plugin's settings. The
user's vault is not a sandbox.

## The loop

### 1. Clarify the spec

Restate the request as three short lists before touching code:

- **Behaviours**: what the user sees or can do, including when it updates.
- **API surfaces**: which classes and events (see `references/api-map.md`).
- **Done when**: the observable checks you will run in step 5, such as "view lists 3 notes in the
  test vault", "no errors in console after reload", "command appears in palette".

Ask only if a behaviour is ambiguous in a way that changes the code. Otherwise state your
assumption and go.

### 2. Explore before writing

Inspect the live app so the code targets what is there:

- the DOM and CSS classes of the UI you will touch or sit beside
- existing views and leaves, registered commands, enabled plugins
- the metadata cache for real notes the feature reads
- the plugin's own loaded state (`app.plugins.plugins["<id>"]`)

Snippets: `references/runtime-inspection.md`. Check signatures in `obsidian.d.ts`. If the only way
to do something is an undocumented `app.*` internal, say so to the user, isolate it behind one
function, and guard it so a missing property degrades instead of throwing.

### 3. Prototype in the runtime

Try the core logic with eval against the live app first, for example the query that collects the
data, or the DOM structure you plan to render. One eval answers in seconds what a
build-reload cycle answers in a minute. Return plain projections (`{path, fm}`), not `TFile` or
`App`, which cannot be serialised.

### 4. Implement

Write it into the plugin project using `references/patterns.md` and
`references/lifecycle-and-cleanup.md`. `main.ts` only wires things up. Each view, modal, settings
tab and editor extension lives in its own module. No project yet: `references/project-setup.md`.

### 5. Build, reload, verify

1. `npm run build`. Fix type errors before anything else.
2. Deploy and reload. Use `scripts/deploy.sh <vault-path>` when the project is outside the vault,
   or `plugin:reload` when it already lives in `<vault>/.obsidian/plugins/<id>/`.
3. Clear the console, exercise the feature (run the command or open the view via eval), then read
   the console. Any error from the plugin is a failure.
4. Query the DOM for the elements the spec promised, and screenshot the affected UI.
5. Check the "done when" list item by item. Loop until each one passes.

Report what you verified and how. If a check could not run, say so and why.

### 6. Review

Run `npx eslint .` from the repo root (not `eslint src`, which skips `manifest.json` and
`LICENSE`), then walk `references/review-checklist.md`. Fix every error. Treat warnings as
blockers too, since they count against the plugin in review.

## Rules that matter most

Each one maps to a review rejection or a real bug.

1. **Register everything.** `registerEvent`, `registerDomEvent`, `registerInterval`,
   `addCommand`, `registerView`, `registerEditorExtension`. Then `onunload` has nothing to do.
   Never detach leaves in `onunload`.
2. **Never hold a view reference on the plugin.** Find views with
   `workspace.getLeavesOfType(TYPE)`.
3. **Defer startup work** to `this.app.workspace.onLayoutReady(...)`. `onload` must stay fast.
4. **Files through the Vault and FileManager APIs.** Use `vault.process` for background edits,
   `fileManager.processFrontMatter` for frontmatter, the `Editor` for the open note, and
   `fileManager.trashFile` for deletes. Look files up with `vault.getFileByPath`, never by looping
   all files. Pass user paths through `normalizePath`. Use `vault.adapter` only for files outside
   the vault index.
5. **Active view via `workspace.getActiveViewOfType(MarkdownView)`** or
   `workspace.activeEditor?.editor`, never `workspace.activeLeaf`.
6. **Build DOM with `createEl`, `createDiv` and `createSpan`**, and clear with `el.empty()`. Never
   use `innerHTML`, `outerHTML` or `insertAdjacentHTML`, and never inline styles. Style with
   classes in `styles.css` using Obsidian CSS variables.
7. **Narrow types, never cast.** `if (f instanceof TFile)`. Strict TypeScript, no `any`.
8. **Network via `requestUrl`**, not `fetch`. It works on mobile and avoids CORS.
9. **Mobile by default.** No Node or Electron imports unless `isDesktopOnly: true`. Use
   `Platform` to branch. No regex lookbehind, which breaks older iOS.
10. **UI text in sentence case.** Settings headings via `new Setting(el).setName(…).setHeading()`,
    and only when there is more than one section. No default hotkeys. Command ids and names
    must not repeat the plugin id or name, since Obsidian prefixes them.

## Where to look

| When you need… | Read |
|---|---|
| Which API for a job: Vault vs Adapter vs FileManager, Workspace and leaves, MetadataCache, Editor vs CM6, events | `references/api-map.md` |
| Code shape for a view, modal, suggest, settings tab (incl. 1.13 declarative settings), command, ribbon, status bar, editor extension, post-processor | `references/patterns.md` |
| `onload`/`onunload`, the `register*` family, leaks, `onLayoutReady`, multi-window | `references/lifecycle-and-cleanup.md` |
| Eval snippets, backend setup, CLI gotchas, debugging "renders blank" or "no effect" | `references/runtime-inspection.md` |
| Pre-submission: guidelines, eslint-plugin-obsidianmd rules, manifest, release | `references/review-checklist.md` |
| New plugin, build config, hot reload, version bump, releases | `references/project-setup.md` |

## Output

Keep replies short: the spec lists, what you changed (files), what you verified and how, and
anything unverified or using private API. Show code only when the user needs to see it.
