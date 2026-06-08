# Solid + Kobalte — Authoring & Bundle Discipline (Solid mode)

> Solid mode only. The styling discipline (tokens, `cva`/`cn`, responsive) is shared and lives in [theming.md](./theming.md), [cva.md](./cva.md), [responsive.md](./responsive.md) — those examples use React syntax; this file gives the Solid equivalents. The primitive layer is **Kobalte**; components come from **Zaidan** (Kobalte-only, no corvu).

---

## Authoring idioms — React → Solid

| React (in the shared references) | Solid equivalent |
|---|---|
| `className="..."` | `class="..."` |
| `className={cn(...)}` | `class={cn(...)}` |
| `function C({ className, ...props })` (destructure) | `splitProps(props, ["class"])` — **never destructure props**, it breaks reactivity |
| `{...props}` spread | `{...others}` (the rest from `splitProps`) — spreading is fine, destructuring is not |
| `asChild` + `<Slot>` | Kobalte polymorphic **`as` prop**, or `<Dynamic component={...}>` from `solid-js/web` |
| `items.map(x => <li key={x.id}>)` | `<For each={items}>{x => <li>...</li>}</For>` (or `<Index>` when the value, not identity, changes) |
| `React.ComponentProps<"button">` | `ComponentProps<"button">` from `solid-js` |
| `React.ButtonHTMLAttributes<...>` | `JSX.ButtonHTMLAttributes<HTMLButtonElement>` from `solid-js` |

Key Solid rules:
- **Don't destructure props** in the function signature or you lose reactivity. Use `splitProps` to separate `class`/variant props from the rest.
- **`class` is reactive** and merges fine with `cn()`. Solid also has a `classList={{ active: isActive() }}` helper, but prefer `cn()` for consistency with the shared discipline.
- Icons: use **`lucide-solid`**, never `lucide-react`.

---

## cva + cn in Solid

The `cva`/`cn` *concept* is identical (see [cva.md](./cva.md)) — only the component shell differs. Solid Button equivalent of the cva.md example:

```tsx
import { splitProps, type ComponentProps } from "solid-js"
import { Dynamic } from "solid-js/web"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "~/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default:     "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline:     "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
      },
      size: { default: "h-10 px-4 py-2", sm: "h-9 rounded-md px-3", icon: "size-10" },
    },
    defaultVariants: { variant: "default", size: "default" },
  },
)

type ButtonProps = ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & { as?: string }

function Button(props: ButtonProps) {
  const [local, others] = splitProps(props, ["class", "variant", "size", "as"])
  return (
    <Dynamic
      component={local.as ?? "button"}
      class={cn(buttonVariants({ variant: local.variant, size: local.size }), local.class)}
      {...others}
    />
  )
}
```

Note: `splitProps` replaces destructuring; `class` (not `className`); `<Dynamic>` gives the `asChild`-style polymorphism. Everything inside `cva(...)` — the Tailwind classes, semantic tokens, variant matrix — is byte-for-byte the same as React.

---

## Kobalte primitive APIs

Kobalte is the headless primitive layer (the Solid analogue of Radix/Base UI). General patterns:

- **Compound components** with a `.Root` and parts, e.g. `Dialog.Root` / `Dialog.Trigger` / `Dialog.Content`.
- **Polymorphism** via the `as` prop on most primitives (`<Dialog.Trigger as={Button}>`), not `asChild`.
- **Overlays** (Dialog, Popover, Tooltip, Select) manage their own portal + stacking — don't add manual `z-index` (matches the shared Critical Rules).
- Always read the component's page on the Kobalte docs (kobalte.dev) before wiring it — the controlled/uncontrolled props and slot names differ per component.

**Import per subpath, never the barrel** (this matters for tree-shaking — see below):

```tsx
import { Select } from "@kobalte/core/select"      // ✅ per-component subpath
import * as Kobalte from "@kobalte/core"            // ❌ barrel — pulls everything
```

---

## Bundle-weight discipline (the Solid-only rule)

Solid's value is a tiny runtime. Kobalte/corvu-style primitive libraries erode the **bundle-size** pillar (not the runtime pillar) when adopted wholesale, because some components pull a large shared internal graph that isn't fully tree-shaken. Zaidan mitigates this by dropping corvu — but the discipline still applies. **Weight should be proportional to interaction complexity.**

Decision rule for every component you add:

```
Adding a component →
  Is it a styled wrapper around native elements          → hand-roll it, ZERO Kobalte
  (button, badge, card, breadcrumb, separator, input,        (just Tailwind + cva + cn)
   label, alert, skeleton, table)?
  Does it need focus trapping / ARIA / keyboard nav /     → use Kobalte; the weight is earned
  positioning / typeahead (dialog, popover, select,
  combobox, menu, tooltip, datepicker, tabs)?
  Simple component the source built on Kobalte anyway     → measure; trim the import or hand-roll
  (e.g. a pin-input)?                                        if it pulls weight with no a11y payoff
```

Practices:
- **Hand-roll the trivial tier.** A breadcrumb is `<nav>/<ol>/<li>` — Kobalte buys it nothing. Take Zaidan's *styling* (classes + cva variants) and drop the primitive import.
- **Per-subpath imports** (`@kobalte/core/select`), never the barrel.
- **Lazy-load heavy interactive components** (datepicker, combobox) with `lazy()` + `<Suspense>` so they stay out of the initial bundle.
- **Measure, gate per-component.** Add `rollup-plugin-visualizer` to the Vite config and set a bundle budget. Before keeping a Kobalte-backed component, look at what it actually pulls; if a simple component drags a heavy graph for no a11y benefit, hand-roll it.
- **Own-your-code means you can trim.** Because Zaidan/shadcn-solid are copy-paste, the Kobalte coupling is local and editable — auditing a generated file is minutes, not a rewrite.
