# shadcn-solid CLI — Quick Reference (Solid mode)

> Solid mode only. For the React `shadcn` CLI see [cli.md](./cli.md).
> Always use the project's package runner: `npx shadcn-solid@latest`, `pnpm dlx shadcn-solid@latest`, or `bunx --bun shadcn-solid@latest`. Check `packageManager` from `package.json`.

The Solid port adopts shadcn's **copy-paste / own-your-code** convention — the CLI writes editable component source into your repo, it is not a runtime dependency you import. Because you own the file, you can (and should) inspect and trim what it generates. See [primitives-solid.md](./primitives-solid.md) for the bundle-weight discipline that decides what to keep.

---

## The CLI is thin — know what's missing

The shadcn-solid CLI implements **only three verbs**:

| Command | Purpose |
|---|---|
| `init` | Initialise the project: install deps, add the `cn` util, configure Tailwind + CSS variables, create `ui.config.json`. Flags: `-c/--cwd <path>`. |
| `add [component]` | Add components (interactive picker, or name them). Flags: `-o/--overwrite`, `-c/--cwd <path>`, `-a/--all`. |
| `diff [component]` | Check installed components against upstream for updates. |

**There is NO equivalent of the React CLI's:** `info` / `info --json`, `search`, `docs`, `--dry-run`, `--diff` flag, or `--view`. Do not invent these — they will error. Use the substitutes below.

```bash
npx shadcn-solid@latest init
npx shadcn-solid@latest add button card dialog
npx shadcn-solid@latest add            # interactive picker
npx shadcn-solid@latest diff button    # see what changed upstream
```

---

## Substitutes for the missing verbs

| Need (React verb) | Solid substitute |
|---|---|
| Project context (`info --json`) | Read `ui.config.json` (paths, aliases, Tailwind/CSS file, base color) and `package.json` (`tailwindcss` version, installed deps). |
| Discover components (`search`) | Browse the docs sites in the browser — **zaidan.carere.dev** and shadcn-solid.com. |
| Component docs (`docs`) | Read the same docs sites; fetch the page for the component's API and examples before writing. |
| Preview before write (`--dry-run`/`--diff`) | Run `add` then review the generated file (it's copy-paste — nothing is hidden in `node_modules`); for an existing component use `diff` to see upstream changes before overwriting. |

---

## Components: Zaidan (standardized source)

This project standardizes on **Zaidan** (`carere/zaidan`, zaidan.carere.dev): a copy-paste component set built on **Kobalte only — no corvu**, chosen specifically to avoid the bundle bloat that affects solid-ui and stock shadcn-solid. Two delivery paths — verify which applies in this repo:

1. **shadcn-solid CLI registry** — if `ui.config.json` is pointed at a Zaidan-compatible registry, `npx shadcn-solid@latest add <component>` pulls Zaidan source directly.
2. **Manual copy-paste** — otherwise, copy the component source from zaidan.carere.dev into the project's UI directory (the path in `ui.config.json`), then fix imports/aliases.

Either way, the result is editable source you own. **After adding, always:**
- Confirm it uses `class` (not `className`) and imports only `@kobalte/core/*` (or plain elements) — see [primitives-solid.md](./primitives-solid.md).
- Apply the **bundle-weight check**: if a simple component pulls a heavy Kobalte graph for no accessibility benefit, hand-roll it instead.
- Replace any `lucide-react` icon imports with `lucide-solid`.

---

## init checklist

After `init`, confirm it produced a coherent setup:
- `ui.config.json` exists with correct component path + alias.
- `cn` util added (clsx + tailwind-merge) — matches the shared [cva.md](./cva.md) implementation.
- Tailwind CSS variables / `@theme inline` wired per [theming.md](./theming.md).
- Tailwind v4 → `@theme inline`; v3 → `tailwind.config.js` (check the `tailwindcss` version in `package.json`).
