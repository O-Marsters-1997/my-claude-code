# shadcn CLI — Quick Reference

> Always use the project's package runner: `npx shadcn@latest`, `pnpm dlx shadcn@latest`, or `bunx --bun shadcn@latest`. Check `packageManager` from `npx shadcn@latest info`.

---

## Adding components

```bash
npx shadcn@latest add button card dialog
npx shadcn@latest add @magicui/shimmer-button     # community registry
npx shadcn@latest add owner/repo/item             # GitHub registry
```

After adding, always read the added files and verify composition before moving on.

---

## Previewing before install

Use `--dry-run`, `--diff`, or `--view` to inspect changes without writing files.

```bash
npx shadcn@latest add button --dry-run            # list all files that would be written
npx shadcn@latest add button --diff               # show diffs (first 5 files)
npx shadcn@latest add button --diff button.tsx    # diff a specific file
npx shadcn@latest add button --view button.tsx    # show full file content
npx shadcn@latest add button --diff globals.css   # preview CSS variable changes
```

Use `--diff` before overwriting an existing component to understand what upstream changed.

---

## Searching registries

```bash
npx shadcn@latest search                          # list all components in the default registry
npx shadcn@latest search @shadcn -q "sidebar"    # search the shadcn registry
npx shadcn@latest search @tailark -q "stats"     # search a community registry
npx shadcn@latest search owner/repo -q "login"   # search a GitHub registry
```

---

## Docs and examples

```bash
npx shadcn@latest docs button dialog select
```

Returns resolved URLs for documentation, examples, and API references. Fetch the URLs to get the actual content. **Run this and fetch the URLs before working with any component** — don't guess the API.

---

## Project and component info

```bash
npx shadcn@latest info          # project config and installed components
npx shadcn@latest info --json   # machine-readable output
```

Key fields to check:

| Field | Use |
|---|---|
| `tailwindVersion` | `"v4"` → `@theme inline`; `"v3"` → `tailwind.config.js` |
| `tailwindCssFile` | The global CSS file — always edit this for CSS variables |
| `aliases` | Use the actual alias prefix for imports, never hardcode |
| `base` | `"base"` (Base UI) — affects component APIs and available props |
| `packageManager` | Use for any non-shadcn installs (`pnpm add date-fns` etc.) |
| `isRSC` | When `true`, hooks/event handlers need `"use client"` |
| `resolvedPaths.ui` | File-system path where UI components live |

```bash
npx shadcn@latest view @shadcn/button             # browse registry item without installing
```
