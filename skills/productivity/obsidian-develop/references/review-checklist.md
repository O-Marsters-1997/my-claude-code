# Review checklist

Run before calling a change done and before every release. Sources, which win when they disagree
with this file:

- Plugin guidelines: <https://docs.obsidian.md/Plugins/Releasing/Plugin+guidelines>
- Submission requirements: <https://docs.obsidian.md/Plugins/Releasing/Submission+requirements+for+plugins>
- Manifest: <https://docs.obsidian.md/Reference/Manifest>
- ESLint plugin: <https://github.com/obsidianmd/eslint-plugin>

## Contents

- Lint
- Code
- UI
- Manifest
- Repo and release

## Lint

`npm i -D eslint eslint-plugin-obsidianmd` (needs ESLint 9, typescript-eslint 8, and `obsidian`
installed, or `no-unsupported-api` reports nothing). Flat config:

```js
import obsidianmd from "eslint-plugin-obsidianmd";
import { defineConfig, globalIgnores } from "eslint/config";

export default defineConfig([
  globalIgnores(["node_modules", "dist", "main.js", "esbuild.config.mjs", "version-bump.mjs"]),
  ...obsidianmd.configs.recommended,
  { rules: { "obsidianmd/ui/sentence-case": "error" } },
]);
```

Run `npx eslint .` from the root so `manifest.json` and `LICENSE` are validated. `ui/sentence-case`
is off in `recommended`, so turn it on. Zero errors, zero warnings.

High-signal rules (all `obsidianmd/`): `detach-leaves`, `no-view-references-in-plugin`,
`no-plugin-as-component`, `no-tfile-tfolder-cast`, `no-static-styles-assignment`,
`prefer-create-el`, `no-forbidden-elements`, `no-nodejs-modules`, `no-unsupported-api`,
`prefer-file-manager-trash-file`, `vault/iterate`, `hardcoded-config-path`,
`prefer-window-timers`, `prefer-active-doc`, `regex-lookbehind`, `commands/*`, `settings-tab/*`,
`validate-manifest`, `validate-license`, `no-sample-code`, `sample-names`. Full list in the
plugin README.

## Code

- [ ] `this.app`, never global `app`
- [ ] Every listener, interval and extension registered via `this.register*`. No leaf detaching in `onunload`
- [ ] No view references stored on the plugin
- [ ] Heavy startup work in `onLayoutReady`
- [ ] Active file edits via `Editor`, background edits via `vault.process`, frontmatter via `processFrontMatter`
- [ ] Deletes via `fileManager.trashFile`. Lookups via `getFileByPath`. User paths through `normalizePath`
- [ ] No `workspace.activeLeaf`
- [ ] `instanceof TFile` / `TFolder`, no casts, no `any`
- [ ] `requestUrl` for network
- [ ] No Node or Electron APIs, or `isDesktopOnly: true`
- [ ] Console logs only errors by default
- [ ] `const`/`let`, async/await
- [ ] Sample code and names (`MyPlugin`, `SampleSettingTab`) gone
- [ ] `main.ts` only wires modules together

## UI

- [ ] No `innerHTML`, `outerHTML`, `insertAdjacentHTML`
- [ ] No inline styles. Classes in `styles.css`, using Obsidian CSS variables. No `!important`, no `:has`
- [ ] Sentence case everywhere
- [ ] Settings headings via `setHeading()`, only if more than one section, never "General", "Settings", the plugin name, or "… settings"
- [ ] Commands: no default hotkeys, correct callback type, id and name free of the plugin id or name and the word "command"
- [ ] Works on mobile, or is `isDesktopOnly`

## Manifest

- [ ] `id`: lowercase and hyphens, doesn't contain `obsidian`, doesn't end in `plugin`, unique, matches folder name
- [ ] `name`: no "Obsidian", "Plugin", or bare core-feature names
- [ ] `description`: starts with a verb, no "This plugin", at most 250 characters, ends with a period, no emoji
- [ ] `version` is x.y.z and `minAppVersion` is the true minimum (covers every `@since` you use)
- [ ] `isDesktopOnly` set correctly. `fundingUrl` only for donation services
- [ ] `versions.json` maps each version to its `minAppVersion`

## Repo and release

- [ ] `README.md`, `LICENSE`, `manifest.json` at repo root
- [ ] GitHub release tag equals `manifest.version` exactly (no `v` prefix)
- [ ] Release assets: `main.js`, `manifest.json`, `styles.css` (if any)
- [ ] Submission is through <https://community.obsidian.md>, which reviews automatically and reads `manifest.json` from the default branch
