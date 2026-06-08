---
name: tailwind-shadcn
description: shadcn/ui + Tailwind CSS discipline. Auto-detects the framework — React projects use the official shadcn CLI + Base UI primitives; SolidJS projects use the shadcn-solid CLI + Kobalte primitives (Zaidan components). Use proactively whenever building new UI or refactoring existing UI — component styling, class/className props, CSS variables, design tokens, responsive layouts, component variants, or any time you're about to reach for an arbitrary Tailwind value. Triggers on: "style this component", "add a color", "make this responsive", "refactor the styling", "add a variant", any use of class/className or Tailwind classes, or whenever you are about to reach for an arbitrary Tailwind value like [24px].
allowed-tools: Bash(npx shadcn@latest *), Bash(pnpm dlx shadcn@latest *), Bash(bunx --bun shadcn@latest *), Bash(npx shadcn-solid@latest *), Bash(pnpm dlx shadcn-solid@latest *), Bash(bunx --bun shadcn-solid@latest *)
---

# Tailwind + shadcn/ui

Apply these patterns whether you're building new UI from scratch or refactoring existing code — the rules are the same either way.

## Detect framework first

This skill serves two ecosystems. Before doing anything, read `package.json` to detect which one you're in:

- **`react` / `react-dom` present → React mode.** Primitive layer is **Base UI**; the CLI is the official **`npx shadcn@latest`**. Gather project context by running:
  ```json
  npx shadcn@latest info --json
  ```
  Key fields: `tailwindVersion` (v4 uses `@theme inline`; v3 uses `tailwind.config.js`), `tailwindCssFile` (always edit this file for CSS variables — never create a new one), `aliases` (use the actual alias prefix for imports, never hardcode), `base` (the primitive layer).
- **`solid-js` present → Solid mode.** Primitive layer is **Kobalte** (Zaidan components); the CLI is **`npx shadcn-solid@latest`** (much thinner — no `info`/`search`/`docs`). Gather project context by reading `ui.config.json` + `package.json`. See [references/cli-shadcn-solid.md](./references/cli-shadcn-solid.md) and [references/primitives-solid.md](./references/primitives-solid.md).

## Framework mode

| Concern               | React mode                                                         | Solid mode                                                                                         |
| --------------------- | ------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------- |
| Primitive layer       | Base UI — see "Primitive layer APIs" below                         | Kobalte (Zaidan) — see [references/primitives-solid.md](./references/primitives-solid.md)          |
| CLI                   | `npx shadcn@latest` — see [references/cli.md](./references/cli.md) | `npx shadcn-solid@latest` — see [references/cli-shadcn-solid.md](./references/cli-shadcn-solid.md) |
| JSX class attribute   | `className`                                                        | `class`                                                                                            |
| Component composition | `asChild` + `Slot`                                                 | `<Dynamic>` / Kobalte polymorphic `as`                                                             |
| Lists                 | `.map(... key={})`                                                 | `<For each>` / `<Index>`                                                                           |

**Normalization note (Solid mode):** the styling discipline below and the shared references (`theming.md`, `responsive.md`, `cva.md`) use React (`className`) syntax in their examples. In Solid mode, **read every `className` as `class`**, read `.map(...)`+`key` as `<For each>`, and ignore React-only constructs (`React.*` types, `asChild`, `Slot`). See [references/primitives-solid.md](./references/primitives-solid.md) for the Solid equivalents. The Tailwind/token discipline itself is identical in both modes.

---

## Principles

1. **Reusability first — exhaust existing options before writing any inline/custom component.** Work down this ladder in order:
   1. **Search what the project already has.** Scan the UI components directory (`resolvedPaths.ui` in React mode; the components dir in Solid mode) for an existing component that fits.
   2. **Extend what exists.** If an existing component nearly fits and it makes semantic sense, reuse it. Add a new `cva` variant **only when the styling will be reused in more than one place** — for a one-off, prefer composition or a `className` override. Don't reach for a variant by default; overusing `cva` adds complexity. See [references/cva.md](./references/cva.md).
   3. **Bring in an available component.** Check the registry before writing custom UI — `npx shadcn@latest search` (React); browse the shadcn-solid catalog (Solid, no `search`). Check community registries too.
   4. **Only then write custom code**, and consult the component docs first (`npx shadcn@latest docs <component>` in React mode) for the recommended reusable pattern.
2. **Compose, don't reinvent.** Settings page = Tabs + Card + form controls. Dashboard = Sidebar + Card + Chart + Table.
3. **Use built-in variants before custom styles.** `variant="outline"`, `size="sm"`, etc.
4. **Use semantic colors.** `bg-primary`, `text-muted-foreground` — never raw values like `bg-blue-500`.

---

## Critical Rules

These rules are always enforced.

### Colors

- **Never** use raw Tailwind palette values (`text-blue-500`, `bg-emerald-600`) in application code.
- **Never** hardcode hex/rgb/oklch values inline in className strings.
- **Use semantic tokens** for all color decisions: `bg-primary`, `text-destructive`, `border-border`.
- For status indicators with no matching token, use `Badge` variants first. If a reusable semantic token is needed (e.g. `--success`), add it to the theme.

### Spacing & layout

- **No `space-x-*` or `space-y-*`.** Use `flex` + `gap-*` instead. `space-y-4` → `flex flex-col gap-4`.
- **`size-*` when width = height.** `size-10` not `w-10 h-10`.
- **`truncate`** not `overflow-hidden text-ellipsis whitespace-nowrap`.

### Component styling

- **`className` is for layout, not overrides.** Acceptable: `max-w-md mx-auto mt-4`. Not acceptable: `bg-blue-100 text-blue-900 font-bold`.
- **Built-in variants first.** Exhaust `variant="outline"`, `variant="destructive"`, `size="sm"` before adding new ones.
- **Add a new variant sparingly** — only when a styling pattern repeats (used in more than one place) and no built-in fits. Edit the component source using `cva`. Not every component needs variants; a one-off doesn't warrant one. See [references/cva.md](./references/cva.md).

### Class merging

- **`cn()` for all conditional or merged classes.** No `${isActive ? 'bg-primary' : 'bg-muted'}` in className.
- **No manual `z-index`** on overlay components (Dialog, Sheet, Popover, Tooltip, Drawer). They manage their own stacking.
- **No manual `dark:` overrides.** Semantic tokens handle theme switching via CSS variables.

---

## CSS Architecture

### Variables at `:root`

All color tokens are CSS custom properties on `:root` in the global CSS file (`tailwindCssFile` from project context). Never scatter variable definitions across component files.

```css
:root {
  --background: oklch(1 0 0);
  --foreground: oklch(0.145 0 0);
  --primary: oklch(0.205 0 0);
  --primary-foreground: oklch(0.985 0 0);
  --muted: oklch(0.97 0 0);
  --muted-foreground: oklch(0.556 0 0);
  /* ... see references/theming.md for the full token set */
}
```

**OKLCH:** `oklch(lightness chroma hue)` — perceptually uniform, so equal chroma values look equally saturated across hues. Prefer it over HSL.

### Map variables to Tailwind

With **Tailwind v4**, register every CSS variable via `@theme inline`:

```css
@theme inline {
  --color-background: var(--background);
  --color-foreground: var(--foreground);
  --color-primary: var(--primary);
  --color-primary-foreground: var(--primary-foreground);
  /* ... see references/theming.md for the full mapping */
  --radius-sm: calc(var(--radius) - 4px);
  --radius-md: calc(var(--radius) - 2px);
  --radius-lg: var(--radius);
  --radius-xl: calc(var(--radius) + 4px);
}
```

### Apply base styles

```css
@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
  }
}
```

Do not double-wrap: inside `@layer` blocks, use utility classes (`bg-background`) or `var(--background)` directly — never `oklch(var(--background))`.

---

## Adding Theme Tokens

Before reaching for an arbitrary value or a raw Tailwind color:

```
Need a color or spacing unit →
  Does a semantic token cover it?            → use it
  Does a Tailwind default utility cover it?  → use it
  Is this a one-off pixel nudge?             → arbitrary is fine: [3px]
  Will this value appear in more than once?  → add a token
```

See [references/theming.md](./references/theming.md) for the full custom color and spacing token workflow, naming conventions, and the complete token table.

---

## Components & CLI

> Run all CLI commands with the project's package runner (`npx` / `pnpm dlx` / `bunx --bun`) — from `packageManager` in project context.

### React mode

```bash
npx shadcn@latest add button card dialog
npx shadcn@latest add @magicui/shimmer-button   # from a community registry
```

**Preview before installing** with `--dry-run` / `--diff` — see [references/cli.md](./references/cli.md). Full registry surface available: `search`, `docs`, `info`.

### Solid mode

```bash
npx shadcn-solid@latest add button card dialog
```

The shadcn-solid CLI is thin — **only `init`, `add`, and `diff`**. There is **no** `info`, `search`, `docs`, `--dry-run`, or `--diff`/`--view` flag. Browse component docs at zaidan.carere.dev / shadcn-solid.com in the browser instead, and preview changes via the `diff` command or by reading the copy-paste source. See [references/cli-shadcn-solid.md](./references/cli-shadcn-solid.md).

**After adding (either mode), always read the added files** and verify correct composition: check for missing sub-components, incorrect nesting, and violations of the Critical Rules above. In Solid mode, also apply the bundle-weight check from [references/primitives-solid.md](./references/primitives-solid.md) — trim or hand-roll components that pull Kobalte for no a11y benefit.

### Primitive layer APIs

**React mode — Base UI.** Key differences from Radix to keep in mind:

- **Composition:** use the `render` prop, not `asChild`. `<DialogTrigger render={<Button />}>Open</DialogTrigger>`
- **Select:** requires an `items` prop on the root; `defaultValue` for placeholder is `null` in the items array.
- **ToggleGroup / Accordion:** use `multiple` boolean instead of `type="multiple"`; `defaultValue` is always an array.
- **Slider:** accepts a plain number for a single thumb (`defaultValue={50}`), not an array.

**Solid mode — Kobalte.** The primitive APIs, composition (`as` prop), and the bundle-weight discipline that governs _when_ to reach for a Kobalte primitive at all are covered in [references/primitives-solid.md](./references/primitives-solid.md). Read it before adding any interactive Solid component.

---

## Workflow

1. **Detect framework & get project context** — read `package.json` (React vs Solid), then gather context per "Detect framework first" above (React: `npx shadcn@latest info --json`; Solid: read `ui.config.json` + `package.json`).
2. **Check installed components first** — before running `add`, check what's already in the UI directory (`resolvedPaths.ui` in React; the components dir from `ui.config.json` in Solid). Don't import components that haven't been added, and don't re-add ones already installed.
3. **Find the component** —
   - _React:_ `npx shadcn@latest search` to discover components (community registries too).
   - _Solid:_ no `search` — browse zaidan.carere.dev / shadcn-solid.com in the browser.
4. **Get docs and examples** —
   - _React:_ `npx shadcn@latest docs <component>` for URLs, then fetch; use `--dry-run` / `--diff` to preview.
   - _Solid:_ no `docs`/preview flags — read the docs site, and preview via `diff` or by reading the copy-paste source after `add`.
5. **Install** — `npx shadcn@latest add <component>` (React) or `npx shadcn-solid@latest add <component>` (Solid).
6. **Review added files** — read every added file; fix composition issues, icon imports, and alias paths. _Solid:_ also run the bundle-weight check ([references/primitives-solid.md](./references/primitives-solid.md)) — hand-roll or trim components that pull Kobalte without an a11y payoff.
7. **Apply styling rules** — use semantic tokens, `cn()`, the `cva` variant pattern, and mobile-first responsive styles per the Critical Rules above.

---

## CVA vs cn

See [references/cva.md](./references/cva.md) for the full guide.

**Quick rule:**

- `cva` lives in component definitions — it encodes what variants a component supports.
- `cn` lives at call sites — it merges incoming `className` with cva output and handles conditionals.

```tsx
const buttonVariants = cva("inline-flex items-center ...", {
  variants: { variant: { default: "bg-primary ...", outline: "border ..." } },
  defaultVariants: { variant: "default" },
});

function Button({ className, variant, ...props }) {
  return (
    <button className={cn(buttonVariants({ variant }), className)} {...props} />
  );
}
```

---

## Responsive Design

See [references/responsive.md](./references/responsive.md) for breakpoint reference and patterns.

**Key rules:**

- Write **base (mobile) styles first**, then layer breakpoint modifiers upward: `sm:`, `md:`, `lg:`.
- **Flex by default.** `flex flex-col` for vertical stacks, `flex` for horizontal.
- **Grid only when genuinely two-dimensional** — real rows _and_ columns (data tables, calendar grids). Never use grid as a substitute for flex on a single axis.

```tsx
// Good: flex column on mobile, row on desktop
<div className="flex flex-col gap-4 md:flex-row md:items-center">
```

---

## References

**Shared (both modes):**

- [references/theming.md](./references/theming.md) — Full OKLCH token table, `@theme inline` mapping, adding custom tokens
- [references/cva.md](./references/cva.md) — `cva` vs `cn`, variant patterns, compound variants
- [references/responsive.md](./references/responsive.md) — Breakpoints, flex/grid decision guide, common responsive patterns

**React mode:**

- [references/cli.md](./references/cli.md) — `npx shadcn@latest`: adding, previewing, searching registries, docs, project info

**Solid mode:**

- [references/primitives-solid.md](./references/primitives-solid.md) — Kobalte/Zaidan authoring idioms, primitive APIs, and the bundle-weight selective-adoption discipline
- [references/cli-shadcn-solid.md](./references/cli-shadcn-solid.md) — `npx shadcn-solid@latest`: the thin CLI surface, Zaidan delivery, browser-docs workflow
