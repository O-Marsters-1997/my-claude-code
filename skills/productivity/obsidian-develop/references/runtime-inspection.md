# Runtime inspection

How to connect to a running Obsidian, and ready-made evals for each thing worth looking at.
Snippets use the global `app`, which is fine for inspection. Shipped code uses `this.app`.

## Contents

- Backend setup
- CLI gotchas
- Snippets: plugins, commands, workspace, metadata, DOM, settings
- Recipes: "renders blank", "no effect", "fires twice"

## Backend setup

**Official CLI** (installer 1.12.7+). Obsidian must be running. Settings → General → Command line
interface, then accept the PATH registration (macOS links `/usr/local/bin/obsidian`). Pass
`vault=<name>` first to target a vault other than the focused one. Reference:
<https://help.obsidian.md/cli>.

**MCP over the CLI**, e.g. `live-mcp-for-obsidian`
(<https://github.com/gapmiss/live-mcp-for-obsidian>):
`claude mcp add obsidian-live -- npx live-mcp-for-obsidian --vault "<name>"`. Needs the CLI setting
above. No debug port.

**MCP over CDP**, e.g. `obsidian-cdp-mcp`: quit Obsidian, then start it with
`/Applications/Obsidian.app/Contents/MacOS/Obsidian --remote-debugging-port=9222`.
`open -a Obsidian --args …` does nothing if Obsidian is already running. Check with
`curl -s http://localhost:9222/json`.

## CLI gotchas

- Exit code is 0 even on failure. Check stdout for a line starting `Error:`.
- Eval output starts with `=> `. `undefined` prints as blank. An async IIFE prints blank unless it
  `return`s. `Map` and `Set` print `{}`, so spread them into arrays.
- Returning `app`, a plugin, a `TFile` or a DOM node gives a circular-JSON error or noise. Return
  projections.
- Never run `obsidian <subcommand> --help`. It is parsed as content and creates an `Untitled` note.
- `dev:console` captures only after `obsidian dev:debug on`. Turn it on once per session before
  reloading.
- Quote code for the shell: `obsidian eval code='(() => { … })()'`, using single quotes outside
  and double quotes inside.
- Calls time out after about 10 seconds in MCP wrappers, so keep evals small.

## Snippets

### Plugins

```js
// enabled plugins with versions
[...app.plugins.enabledPlugins].map(id => ({ id, v: app.plugins.manifests[id]?.version }))

// is mine loaded, and what are its live settings
(() => { const p = app.plugins.plugins["<id>"]; return { loaded: !!p, settings: p?.settings }; })()

// reload (disable then enable)
(async () => { await app.plugins.disablePlugin("<id>"); await app.plugins.enablePlugin("<id>"); return "ok"; })()

// first install: plugin:reload fails with "Plugin not found" until manifests are re-read
(async () => { await app.plugins.loadManifests(); await app.plugins.enablePluginAndSave("<id>"); return "ok"; })()
```

### Commands

```js
Object.values(app.commands.commands).filter(c => c.id.startsWith("<id>:")).map(c => ({ id: c.id, name: c.name }))
app.commands.executeCommandById("<id>:<command-id>")
```

### Workspace

```js
(() => { const out = []; app.workspace.iterateAllLeaves(l => out.push({ type: l.view.getViewType(), title: l.getDisplayText(), file: l.view.file?.path })); return out; })()
app.workspace.getLeavesOfType("<view-type>").length
```

### Files and metadata

```js
(() => { const f = app.workspace.getActiveFile(); return f && { path: f.path, fm: app.metadataCache.getFileCache(f)?.frontmatter, tags: app.metadataCache.getFileCache(f)?.tags?.map(t => t.tag) }; })()

// prototype a query, e.g. notes tagged #todo with a due field
app.vault.getMarkdownFiles().map(f => ({ f, c: app.metadataCache.getFileCache(f) }))
  .map(({ f, c }) => ({ f, c, tags: [...(c?.tags ?? []).map(t => t.tag), ...[].concat(c?.frontmatter?.tags ?? []).map(t => "#" + String(t).replace(/^#/, ""))] }))
  .filter(({ tags }) => tags.includes("#todo"))
  .map(({ f, c }) => ({ path: f.path, due: c?.frontmatter?.due }))
```

The `obsidian` module isn't a global, so `getAllTags` is unavailable in eval. The inline union
above matches it. In plugin code, import and use `getAllTags`.

### DOM and CSS

```js
// what is actually rendered inside a view or container
(() => { const el = document.querySelector(".workspace-leaf.mod-active .view-content"); return el && { children: el.childElementCount, html: el.innerHTML.slice(0, 800) }; })()
```

Or `obsidian dev:dom selector=".todo-list li" text all`, and
`obsidian dev:css selector=".todo-list" prop=color` to see which rule wins and where it comes from.
Prefer Obsidian's own variables (`--text-muted`, `--background-modifier-border`,
`--size-4-2`) when copying a look from neighbouring UI.

### Settings tab

```js
(() => { app.setting.open(); app.setting.openTabById("<id>"); const c = app.setting.activeTab?.containerEl; return { tab: app.setting.activeTab?.id, children: c?.childElementCount, text: c?.innerText.slice(0, 300) }; })()
```

On 1.13+ settings may open in a separate window, so a screenshot of the main window can miss
it. Rely on the eval result, or take the screenshot by selector.

## Recipes

### "Renders blank"

1. Console first: `dev:errors`, or `dev:console level=error`. A throw inside `display()` or
   `onOpen()` stops rendering.
2. Open the UI by eval (settings snippet above, or `getLeavesOfType`), then count children of
   its container. Zero children and no error means the elements were built somewhere else, such
   as a global `createDiv()` that was never attached, or a different container.
3. Check the plugin's state: `settings` undefined means `loadSettings` was not awaited or ran
   after the UI was registered.
4. Children present but invisible: query computed CSS (`display`, `height`, `opacity`) on the
   container.

### "No effect"

1. Is the plugin loaded and is the latest build deployed? Compare `app.plugins.manifests[id].version`,
   and grep the deployed `main.js` for a string you just added.
2. Is the command registered? List commands for the id.
3. Is the event firing? Temporarily wrap it by eval:
   `app.vault.on("modify", f => console.log("modify", f.path))`. This listener survives until
   app reload, so say so.

### "Fires twice"

Reload-leak. See `lifecycle-and-cleanup.md`, "Checking for leaks in the live app".
