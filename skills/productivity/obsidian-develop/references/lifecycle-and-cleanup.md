# Lifecycle and cleanup

A plugin is reloaded many times a session during development and by users toggling it. Anything
not torn down on unload doubles up on the next load: duplicate listeners, intervals, DOM nodes.
The cure is to route every resource through `Component`'s register methods so `onunload` has
nothing to do.

Docs: <https://docs.obsidian.md/Plugins/Getting+started/Anatomy+of+a+plugin>

## Order of events

1. `onload()` runs when the plugin is enabled, often during app start before the layout exists.
2. `workspace.onLayoutReady(cb)` runs `cb` once the layout is restored, immediately if it already is.
3. `onunload()` runs on disable, reload and app quit.
4. `onUserEnable()` (1.7.2+) runs only when the user enables the plugin by hand. Open your view
   here, not in `onload`, so it doesn't steal a leaf on every app start.

`onload` should register things and return. Scanning the vault, opening views, or subscribing to
vault `create` belongs in `onLayoutReady`. Vault `create` fires once per file during startup
indexing otherwise.

## What to register with

| Resource | Register with |
|---|---|
| App, vault, workspace or metadata event | `this.registerEvent(this.app.vault.on(...))` |
| DOM listener on `document`, `window` or foreign elements | `this.registerDomEvent(el, "click", fn)` |
| Interval | `this.registerInterval(window.setInterval(fn, ms))` |
| Command, ribbon, status bar, settings tab | `addCommand`, `addRibbonIcon`, `addStatusBarItem`, `addSettingTab` (auto-removed) |
| View type | `registerView` |
| CM6 extension | `registerEditorExtension` |
| Post-processor, code block | `registerMarkdownPostProcessor`, `registerMarkdownCodeBlockProcessor` |
| Anything else (observer, third-party listener) | `this.register(() => observer.disconnect())` |
| A child object with its own lifetime | `this.addChild(component)` |

Use `window.setTimeout` and `window.setInterval`, not bare `setTimeout`. That keeps the
`number` return type and passes `obsidianmd/prefer-window-timers`.

## Don'ts

- **Don't detach leaves in `onunload`.** Obsidian keeps the leaf and restores it as a placeholder
  when the plugin comes back. Detaching loses the user's layout.
- **Don't keep `this.view = new MyView()`.** Views are created by the workspace and can be
  recreated or deferred. Look them up with `getLeavesOfType`.
- **Don't pass the plugin as the `Component` to `MarkdownRenderer.render`.** The rendered children
  then live until the plugin unloads. Pass the view or a `MarkdownRenderChild`.
- **Don't attach DOM to `document.body` without removing it.** Wrap in `this.register(() => el.remove())`.
- **Don't do async work in `onunload` that must finish.** It is not awaited.

## Views have their own lifecycle

`ItemView` is a `Component`. Events registered with `this.registerEvent` inside a view are
cleaned up when the view closes (`onClose`), independent of the plugin. Register view-specific
listeners on the view, plugin-wide ones on the plugin.

## Multi-window

Pop-out windows have their own `document`. Use `activeDocument` and `activeWindow` for "wherever
the user is", `el.doc` and `el.win` for "the window this element lives in", and
`el.instanceOf(HTMLElement)` rather than `instanceof` for DOM checks across windows.

## Checking for leaks in the live app

After implementing, reload the plugin three times via the backend, then exercise the feature
once and confirm each effect happens once. Count listeners where you can:

```js
(() => { const p = app.plugins.plugins["<id>"]; return { children: p._children?.length, events: p._events?.length }; })()
```

`_children` and `_events` are private `Component` fields. Use them for inspection only. Counts
that grow across reloads mean something escapes the register methods. Also check that the DOM
selectors you added return one match, not one per reload.
