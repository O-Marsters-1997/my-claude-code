# API map: which API for which job

Check every signature in `node_modules/obsidian/obsidian.d.ts`
(<https://github.com/obsidianmd/obsidian-api>). `@since` tags there give the minimum app version.
Raise `minAppVersion` when you use something newer, because `obsidianmd/no-unsupported-api` flags it.

## Contents

- Files and content
- Metadata
- Workspace, leaves and views
- Editing text
- Events
- Platform, network, secrets
- Private API

## Files and content

| Job | Use | Not |
|---|---|---|
| Edit the note open in the editor | `Editor` (`editor.replaceSelection`, `transaction`) | `vault.modify`, which resets cursor and undo |
| Edit a note in the background | `vault.process(file, fn)` (atomic) | `vault.read` + `vault.modify` (race) |
| Change frontmatter | `app.fileManager.processFrontMatter(file, fm => {…})` | regex on the YAML |
| Read for display only | `vault.cachedRead(file)` | `vault.read` unless you'll write back |
| Find a file by path | `vault.getFileByPath(normalizePath(p))`, `getFolderByPath`, `getAbstractFileByPath` | looping `getMarkdownFiles()` |
| All notes | `vault.getMarkdownFiles()` | `getFiles()` filtered by extension |
| Create | `vault.create`, `vault.createFolder` | adapter writes |
| Rename or move | `app.fileManager.renameFile` (updates links) | `vault.rename` |
| Delete | `app.fileManager.trashFile` (respects user setting) | `vault.delete`, `vault.trash` |
| Config folder | `vault.configDir` | hardcoded `.obsidian` |
| Files outside the vault index (plugin cache) | `vault.adapter` | |

`vault.adapter` bypasses the cache and serialisation, so concurrent writes can clobber each other.

## Metadata

`app.metadataCache.getFileCache(file)` gives `frontmatter`, `tags`, `links`, `headings`,
`sections`, `listItems`. Use `getAllTags(cache)` for body tags plus frontmatter tags.

- Cache can lag a just-written file. Listen to `metadataCache.on("changed", (file, data, cache) => …)`
  rather than reading straight after a write.
- `metadataCache.on("resolved")` fires after every resolve pass and is noisy. Use it once at
  startup, guarded, or prefer `changed`.
- Link graph: `metadataCache.resolvedLinks` and `unresolvedLinks`.
- Resolve a link: `metadataCache.getFirstLinkpathDest(linkpath, sourcePath)`.

Docs: <https://docs.obsidian.md/Reference/TypeScript+API/MetadataCache>

## Workspace, leaves and views

| Job | Use |
|---|---|
| Active markdown view | `workspace.getActiveViewOfType(MarkdownView)` |
| Active editor (incl. canvas and embeds) | `workspace.activeEditor?.editor` |
| Active file | `workspace.getActiveFile()` |
| Find your view | `workspace.getLeavesOfType(VIEW_TYPE)` |
| Open your view in the right sidebar | `workspace.getRightLeaf(false)` then `leaf.setViewState({type, active: true})`, then `workspace.revealLeaf(leaf)` |
| Open a file | `workspace.getLeaf(newLeaf?).openFile(file)` or `workspace.openLinkText(link, sourcePath)` |
| Iterate all leaves | `workspace.iterateAllLeaves(leaf => …)` |
| Wait for layout | `workspace.onLayoutReady(cb)` |

`revealLeaf` is async since 1.7.2 and may load a deferred view. Deferred views (1.7.2+): a leaf's
`view` may be a placeholder until revealed, so check `leaf.view instanceof MyView` before using it.

Docs: <https://docs.obsidian.md/Plugins/User+interface/Workspace>,
<https://docs.obsidian.md/Plugins/Guides/Understanding+deferred+views>

## Editing text

| Job | Use |
|---|---|
| One-off edit from a command | `editorCallback: (editor, ctx) => …` |
| Decorations, live behaviour while typing | CM6 extension via `registerEditorExtension` (ViewPlugin or StateField) |
| Change rendered Reading view | `registerMarkdownPostProcessor` |
| Custom fenced block | `registerMarkdownCodeBlockProcessor(lang, …)` |
| Reconfigure an extension | mutate the same registered array, then `workspace.updateOptions()` |

Import CodeMirror from `@codemirror/state` and `@codemirror/view`. They are external in the
build and Obsidian supplies them, so don't bundle your own copy.

Docs: <https://docs.obsidian.md/Plugins/Editor/Editor+extensions>

## Events

| Source | Events |
|---|---|
| `app.vault` | `create`, `modify`, `delete`, `rename` |
| `app.metadataCache` | `changed`, `deleted`, `resolved` |
| `app.workspace` | `file-open`, `active-leaf-change`, `layout-change`, `editor-change`, `file-menu`, `editor-menu`, `css-change` |

Always `this.registerEvent(source.on(...))`. Vault `create` fires for every file during startup
indexing, so register it inside `onLayoutReady`.

## Platform, network, secrets

- `Platform.isMobile`, `isDesktopApp`, `isIosApp`… for branching.
- `requestUrl({url, method, headers, body})` for HTTP. It bypasses CORS and works on mobile.
- `app.secretStorage` with `SecretComponent` in settings (1.11.4+) for API keys, rather than
  plain `data.json`.
- `getLanguage()` (1.8.7+) for the UI locale.

## Private API

`app.plugins`, `app.commands`, `app.setting`, `app.internalPlugins` and anything not in
`obsidian.d.ts` are private. Fine for eval-based inspection. In shipped code, wrap them in one
function, add a local type declaration, guard for absence, and tell the user they rely on it.
