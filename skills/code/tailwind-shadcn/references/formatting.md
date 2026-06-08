# Formatting & Linting

Run the project formatter on changed files after every edit. A non-zero exit is a blocking failure — do not declare the task done until it passes.

## Detection heuristic

Read the `scripts` block in `package.json`. Match script keys and devDependencies to identify the tool:

| Signal | Tool |
|---|---|
| `@biomejs/biome` in devDependencies, or a script body containing `biome` | **Biome** |
| `prettier` in devDependencies | **Prettier + ESLint** |
| `eslint` in devDependencies, no Prettier | **ESLint-only** |
| None of the above | No formatter — skip this step |

When multiple signals match, prefer: **Biome > Prettier+ESLint > ESLint-only**.

Common script keys to look for: `check`, `format`, `lint`, `fix`.

## Commands

Use the project's package runner (`npx` / `pnpm dlx` / `bun run`) from `packageManager` in project context.

### Biome

```bash
# Via script (preferred — respects project config)
bun run check        # or: npm run check / pnpm run check

# Direct (if no script defined)
bunx @biomejs/biome check --write <changed-files…>
```

### Prettier + ESLint

```bash
# Via scripts
npm run format && npm run lint

# Direct
npx prettier --write <changed-files…>
npx eslint --fix <changed-files…>
```

### ESLint-only

```bash
# Via script
npm run lint

# Direct
npx eslint --fix <changed-files…>
```

## Changed files only

Pass only the files touched in the current task — not the whole project. Derive the list from what was written or edited during the session.
