# Project setup

Start from the official sample plugin rather than writing config by hand. It is kept current with
the build, lint and TypeScript settings reviewers expect.

## Scaffold

```bash
git clone --depth 1 https://github.com/obsidianmd/obsidian-sample-plugin <id>
cd <id> && rm -rf .git && git init && npm install
```

Then:

1. Edit `manifest.json`: `id` (matches the folder name), `name`, `description`, `author`,
   `minAppVersion`, `isDesktopOnly`. Rules: `review-checklist.md`, Manifest.
2. Rename `MyPlugin`, `MyPluginSettings`, `SampleSettingTab`, and delete the sample command,
   modal, ribbon icon and interval. `obsidianmd/no-sample-code` and `sample-names` flag leftovers.
3. Upgrade `eslint-plugin-obsidianmd` to latest. The template can lag.
4. Turn on `obsidianmd/ui/sentence-case` in `eslint.config.mts`.

The scaffold already has strict TypeScript (`strict`, `noUncheckedIndexedAccess`,
`noImplicitReturns`), esbuild bundling `src/main.ts` to CommonJS `main.js` with `obsidian`,
`electron`, `@codemirror/*` and `@lezer/*` external, and `npm run dev` / `build` / `lint` /
`version` scripts. Don't add a second `main.ts` at the root.

Reference: <https://github.com/obsidianmd/obsidian-sample-plugin>

## Layout

```
src/
  main.ts           # onload wiring only
  settings.ts       # interface, DEFAULT_SETTINGS, setting tab
  views/…           # one file per view
  commands/…        # command builders
  lib/…             # pure logic, no Obsidian UI, easy to test
styles.css
```

## Getting the build into a vault

Pick one:

- **Develop inside the vault**: clone into `<vault>/.obsidian/plugins/<id>/` and run `npm run dev`.
  Simplest. Use a test vault.
- **Develop elsewhere**: build, then `scripts/deploy.sh <vault-path>`, which copies `main.js`,
  `manifest.json` and `styles.css` and reloads via the CLI. Never `rm -rf` the plugin folder in
  the vault, because it holds the user's `data.json`.
- **Symlink the folder** into `.obsidian/plugins/`. Directory symlinks work with hot-reload, file
  symlinks don't.

### Hot reload

Install pjeby/hot-reload (<https://github.com/pjeby/hot-reload>) in the test vault. It reloads a
plugin about 750 ms after `main.js` or `styles.css` changes, for any plugin folder containing
`.git` or a `.hotreload` file. With it plus `npm run dev`, step 5's "reload" is automatic. Still
read the console after each change.

## Test vault

Create a small vault with fixture notes that exercise the feature: tags, frontmatter, nested
folders, a note with no frontmatter, a renamed note. Commit it alongside the plugin if useful.
Use `vault=<name>` with the CLI so commands never hit the user's main vault.

## Versioning and release

```bash
npm version patch   # runs version-bump.mjs: syncs manifest.json, appends versions.json
git push --follow-tags
```

Create a GitHub release whose tag equals the version exactly, and attach `main.js`,
`manifest.json`, `styles.css`. The sample plugin's `.github/workflows` builds and drafts this on
tag push. Raise `minAppVersion` whenever you adopt an API with a newer `@since`.

Release docs: <https://docs.obsidian.md/Plugins/Releasing/Release+your+plugin+with+GitHub+Actions>
