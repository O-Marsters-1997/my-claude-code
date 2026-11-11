# Patterns

Minimal shapes for each extension point. Adapt names. Every snippet assumes it is wired up from
`main.ts` and that registration happens through `this.*` so unload is automatic.

## Contents

- main.ts wiring
- Custom view (sidebar)
- Modal and suggest modal
- Settings tab (classic and 1.13 declarative)
- Commands
- Ribbon and status bar
- Editor extension (CM6)
- Markdown post-processor and code block

## main.ts wiring

```ts
export default class TodoPlugin extends Plugin {
  settings: TodoSettings;

  async onload() {
    await this.loadSettings();
    this.registerView(VIEW_TYPE_TODO, (leaf) => new TodoView(leaf, this));
    this.addCommand({ id: "open-list", name: "Open todo list", callback: () => this.activateView() });
    this.addSettingTab(new TodoSettingTab(this.app, this));
  }

  async activateView() {
    const { workspace } = this.app;
    let leaf = workspace.getLeavesOfType(VIEW_TYPE_TODO)[0];
    if (!leaf) {
      leaf = workspace.getRightLeaf(false)!;
      await leaf.setViewState({ type: VIEW_TYPE_TODO, active: true });
    }
    await workspace.revealLeaf(leaf);
  }

  async loadSettings() {
    this.settings = Object.assign({}, DEFAULT_SETTINGS, await this.loadData());
  }
  async saveSettings() { await this.saveData(this.settings); }
}
```

`await this.loadSettings()` must come before anything that reads settings. An un-awaited load
leaves `settings` undefined when the settings tab or a view first renders.

## Custom view (sidebar)

```ts
export const VIEW_TYPE_TODO = "todo-list";

export class TodoView extends ItemView {
  constructor(leaf: WorkspaceLeaf, private plugin: TodoPlugin) { super(leaf); }
  getViewType() { return VIEW_TYPE_TODO; }
  getDisplayText() { return "Todo list"; }
  getIcon() { return "check-square"; }

  async onOpen() {
    this.registerEvent(this.app.metadataCache.on("changed", () => this.requestRender()));
    this.registerEvent(this.app.vault.on("delete", () => this.requestRender()));
    this.registerEvent(this.app.vault.on("rename", () => this.requestRender()));
    this.render();
  }

  private requestRender = debounce(() => this.render(), 300, true);

  private render() {
    const root = this.contentEl;
    root.empty();
    const list = root.createEl("ul", { cls: "todo-list" });
    for (const row of collectTodos(this.app)) {
      const li = list.createEl("li");
      li.createEl("a", { text: row.file.basename, cls: "internal-link" })
        .addEventListener("click", () => this.app.workspace.getLeaf().openFile(row.file));
      if (row.due) li.createSpan({ text: row.due, cls: "todo-list-due" });
    }
  }
}
```

- Events registered on the view (`this.registerEvent` inside the view) die with the view.
- Debounce re-renders. `changed` fires per keystroke-save.
- Listeners on elements you create inside `contentEl` are freed when you `empty()` it. Use
  `registerDomEvent` only for `document`, `window` or elements you don't own.
- Keep data collection (`collectTodos`) a pure function of `App` in its own module so it can be
  prototyped by eval and tested.

## Modal and suggest modal

```ts
class ConfirmRename extends Modal {
  constructor(app: App, private onOk: () => void) { super(app); }
  onOpen() {
    this.setTitle("Rename notes");
    new Setting(this.contentEl).addButton((b) =>
      b.setButtonText("Rename").setCta().onClick(() => { this.close(); this.onOk(); }));
  }
  onClose() { this.contentEl.empty(); }
}
```

Pick-from-list: `SuggestModal<T>` or `FuzzySuggestModal<T>` (`getItems`, `getItemText`,
`onChooseItem`). Autocomplete on a text input: `AbstractInputSuggest` (1.4.10+), not a copied
`TextInputSuggest`. Yes/no confirmation on 1.13+: `ConfirmationModal`.

## Settings tab

Classic `display()`, required while `minAppVersion` < 1.13.0:

```ts
display(): void {
  const { containerEl } = this;
  containerEl.empty();
  new Setting(containerEl)
    .setName("Due date field")
    .setDesc("Frontmatter key that holds the due date.")
    .addText((t) => t.setValue(this.plugin.settings.dueField).onChange(async (v) => {
      this.plugin.settings.dueField = v;
      await this.plugin.saveSettings();
    }));
}
```

- Every `Setting` must be attached to `containerEl` (or a child created from it with
  `containerEl.createDiv()`). A bare global `createDiv()` makes a detached element, so the tab
  renders blank with no error.
- Headings: `new Setting(containerEl).setName("Display").setHeading()`, only with two or more
  sections, never "General" or "… settings". `SettingGroup` (1.11+) groups related settings.

Declarative settings (1.13.0+): implement `getSettingDefinitions()` so settings appear in
settings search, and call `this.update()` instead of re-running `display()`. Keep `display()` too
while `minAppVersion` is below 1.13. Persist everything inside `plugin.settings`, because
auto-persist overwrites sibling `saveData` keys. The lint rules under `settings-tab/*` enforce
this. Check the exact types (`SettingDefinition*`) in `obsidian.d.ts` before writing.

Since 1.13 the settings UI can open in its own window. Inside settings callbacks
`activeDocument` is that window. Reach the main UI via
`this.app.workspace.containerEl.ownerDocument`.

## Commands

```ts
this.addCommand({
  id: "wrap-callout",
  name: "Wrap selection in callout",
  editorCheckCallback: (checking, editor) => {
    const sel = editor.getSelection();
    if (!sel) return false;
    if (!checking) editor.replaceSelection(toCallout(sel));
    return true;
  },
});
```

| Callback | When |
|---|---|
| `callback` | Always available |
| `checkCallback(checking)` | Available only in some states |
| `editorCallback(editor, ctx)` | Needs an editor |
| `editorCheckCallback(checking, editor, ctx)` | Needs an editor and a condition |

No `hotkeys`. Id and name must not contain the plugin id, plugin name, or the word "command".
Keep ids stable across releases, since users bind hotkeys to them.

## Ribbon and status bar

```ts
this.addRibbonIcon("check-square", "Open todo list", () => this.activateView());
const status = this.addStatusBarItem();   // desktop only; returns an element
status.setText("3 todos");
```

Icons are Lucide names. Status bar items don't exist on mobile, so guard any logic that depends
on them.

## Editor extension (CM6)

```ts
const highlight = ViewPlugin.fromClass(class {
  decorations: DecorationSet;
  constructor(view: EditorView) { this.decorations = build(view); }
  update(u: ViewUpdate) { if (u.docChanged || u.viewportChanged) this.decorations = build(u.view); }
}, { decorations: (v) => v.decorations });

this.registerEditorExtension([highlight]);
```

Only decorate `view.visibleRanges`. To toggle at runtime, keep one array, register it once,
mutate it, then call `this.app.workspace.updateOptions()`.

## Markdown post-processor and code block

```ts
this.registerMarkdownCodeBlockProcessor("todo-summary", (source, el, ctx) => {
  el.createEl("p", { text: summarise(source) });
});

this.registerMarkdownPostProcessor((el, ctx) => {
  el.querySelectorAll("span.due").forEach((s) => s.addClass("todo-due"));
});
```

Post-processors run per rendered section, not per document. When rendering Markdown yourself,
pass a `Component` you own (the view or a `MarkdownRenderChild`) to `MarkdownRenderer.render`,
never the plugin. Otherwise children leak until the plugin unloads.

Docs: <https://docs.obsidian.md/Plugins/Editor/Markdown+post+processing>
