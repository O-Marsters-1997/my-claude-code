# Theming — CSS Variables, OKLCH, and Token Architecture

> **Solid mode:** the CSS/token content here is framework-agnostic and applies unchanged. In the few `.tsx` usage examples, read `className` as `class` and skip `React.*` types — see [primitives-solid.md](./primitives-solid.md).

## Why OKLCH

OKLCH (`oklch(lightness chroma hue)`) is the preferred color format for design tokens because it is perceptually uniform: equal steps in lightness or chroma produce equal-looking changes across different hues. HSL does not have this property — `hsl(120 50% 50%)` (green) and `hsl(240 50% 50%)` (blue) look very different in perceived brightness despite identical lightness values.

OKLCH values:
- **Lightness**: 0 (black) → 1 (white)
- **Chroma**: 0 (fully grey) → ~0.37 (very saturated, gamut-dependent)
- **Hue**: 0–360 (red=0/360, yellow~90, green~145, cyan~200, blue~265, purple~300)

OKLCH is supported in all modern browsers. For tooling, the VS Code plugin "OKLCH Color Preview" and the [oklch.com](https://oklch.com) picker are useful for choosing values.

---

## Full Token Table

| CSS Variable | Tailwind utility | Purpose |
|---|---|---|
| `--background` | `bg-background`, `text-background` | Page background |
| `--foreground` | `text-foreground` | Default body text |
| `--card` | `bg-card` | Card surfaces |
| `--card-foreground` | `text-card-foreground` | Text on cards |
| `--surface` | `bg-surface` | Secondary surfaces (panels, sidebars) |
| `--surface-foreground` | `text-surface-foreground` | Text on secondary surfaces |
| `--primary` | `bg-primary` | Primary buttons, key actions |
| `--primary-foreground` | `text-primary-foreground` | Text on primary backgrounds |
| `--secondary` | `bg-secondary` | Secondary actions |
| `--secondary-foreground` | `text-secondary-foreground` | Text on secondary backgrounds |
| `--muted` | `bg-muted` | Disabled states, subtle fills |
| `--muted-foreground` | `text-muted-foreground` | Placeholder, subtext, captions |
| `--accent` | `bg-accent` | Hover highlights |
| `--accent-foreground` | `text-accent-foreground` | Text on hover highlights |
| `--destructive` | `bg-destructive`, `text-destructive` | Errors, delete actions |
| `--destructive-foreground` | `text-destructive-foreground` | Text on destructive backgrounds |
| `--border` | `border-border` | Default borders |
| `--input` | `border-input` | Input field borders |
| `--ring` | `ring-ring` | Focus rings |
| `--chart-1` … `--chart-5` | `bg-chart-1` etc. | Data visualization palette |
| `--radius` | `rounded-sm`, `rounded-md`, `rounded-lg`, `rounded-xl` | Global border radius scale |

---

## Complete globals.css Template (Tailwind v4)

```css
@import "tailwindcss";

:root {
  /* Surfaces */
  --background:          oklch(1 0 0);
  --foreground:          oklch(0.145 0 0);
  --card:                oklch(1 0 0);
  --card-foreground:     oklch(0.145 0 0);
  --surface:             oklch(0.97 0 0);
  --surface-foreground:  oklch(0.205 0 0);

  /* Interactive */
  --primary:             oklch(0.205 0 0);
  --primary-foreground:  oklch(0.985 0 0);
  --secondary:           oklch(0.97 0 0);
  --secondary-foreground: oklch(0.205 0 0);
  --accent:              oklch(0.97 0 0);
  --accent-foreground:   oklch(0.205 0 0);

  /* Feedback */
  --destructive:         oklch(0.577 0.245 27.325);
  --destructive-foreground: oklch(0.985 0 0);

  /* UI chrome */
  --muted:               oklch(0.97 0 0);
  --muted-foreground:    oklch(0.556 0 0);
  --border:              oklch(0.922 0 0);
  --input:               oklch(0.922 0 0);
  --ring:                oklch(0.708 0 0);

  /* Global radius */
  --radius:              0.625rem;

  /* Charts */
  --chart-1: oklch(0.646 0.222 41.116);
  --chart-2: oklch(0.6 0.118 184.704);
  --chart-3: oklch(0.398 0.07 227.392);
  --chart-4: oklch(0.828 0.189 84.429);
  --chart-5: oklch(0.769 0.188 70.08);
}

@theme inline {
  --color-background:          var(--background);
  --color-foreground:          var(--foreground);
  --color-card:                var(--card);
  --color-card-foreground:     var(--card-foreground);
  --color-surface:             var(--surface);
  --color-surface-foreground:  var(--surface-foreground);
  --color-primary:             var(--primary);
  --color-primary-foreground:  var(--primary-foreground);
  --color-secondary:           var(--secondary);
  --color-secondary-foreground: var(--secondary-foreground);
  --color-accent:              var(--accent);
  --color-accent-foreground:   var(--accent-foreground);
  --color-destructive:         var(--destructive);
  --color-destructive-foreground: var(--destructive-foreground);
  --color-muted:               var(--muted);
  --color-muted-foreground:    var(--muted-foreground);
  --color-border:              var(--border);
  --color-input:               var(--input);
  --color-ring:                var(--ring);
  --color-chart-1:             var(--chart-1);
  --color-chart-2:             var(--chart-2);
  --color-chart-3:             var(--chart-3);
  --color-chart-4:             var(--chart-4);
  --color-chart-5:             var(--chart-5);

  /* Radius scale */
  --radius-sm:  calc(var(--radius) - 4px);
  --radius-md:  calc(var(--radius) - 2px);
  --radius-lg:  var(--radius);
  --radius-xl:  calc(var(--radius) + 4px);
}

@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
  }
}
```

---

## Adding Custom Semantic Tokens

### Step 1 — Define in `:root`

Add to the existing `:root` block in `globals.css` (never create a new CSS file):

```css
:root {
  /* ... existing tokens ... */
  --success:             oklch(0.527 0.154 150);
  --success-foreground:  oklch(0.985 0 0);
  --warning:             oklch(0.769 0.188 70);
  --warning-foreground:  oklch(0.145 0 0);
  --info:                oklch(0.6 0.118 250);
  --info-foreground:     oklch(0.985 0 0);
}
```

### Step 2 — Register with Tailwind (v4)

Add to the `@theme inline` block:

```css
@theme inline {
  /* ... existing mappings ... */
  --color-success:             var(--success);
  --color-success-foreground:  var(--success-foreground);
  --color-warning:             var(--warning);
  --color-warning-foreground:  var(--warning-foreground);
  --color-info:                var(--info);
  --color-info-foreground:     var(--info-foreground);
}
```

### Step 3 — Use in components

```tsx
<div className="bg-success text-success-foreground rounded-md px-3 py-2">
  Changes saved
</div>

<Badge className="bg-warning text-warning-foreground">
  Pending
</Badge>
```

Or better — add a variant to the Badge component:

```tsx
// components/ui/badge.tsx
const badgeVariants = cva("...", {
  variants: {
    variant: {
      default:     "...",
      success:     "bg-success text-success-foreground",
      warning:     "bg-warning text-warning-foreground",
      info:        "bg-info text-info-foreground",
    },
  },
})
```

---

## Adding Custom Spacing Tokens

When the project repeatedly uses a spacing value that Tailwind doesn't have a step for, add it rather than using arbitrary values everywhere:

```css
@theme inline {
  --spacing-13: 3.25rem;  /* 52px — between gap-12 and gap-14 */
  --spacing-18: 4.5rem;   /* 72px — between gap-16 and gap-20 */
  --spacing-22: 5.5rem;
}
```

These become available as `gap-13`, `p-18`, `m-22`, `w-18`, etc.

---

## Tailwind v3 Alternative

When `tailwindVersion` is `"v3"` (check via `npx shadcn@latest info`):

```js
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        background:  "oklch(var(--background) / <alpha-value>)",
        foreground:  "oklch(var(--foreground) / <alpha-value>)",
        primary: {
          DEFAULT:    "oklch(var(--primary) / <alpha-value>)",
          foreground: "oklch(var(--primary-foreground) / <alpha-value>)",
        },
        success: {
          DEFAULT:    "oklch(var(--success) / <alpha-value>)",
          foreground: "oklch(var(--success-foreground) / <alpha-value>)",
        },
      },
    },
  },
}
```

Note: with v3, the CSS variable must store only the channel values, not the full `oklch(...)` wrapper:

```css
/* v3 pattern — channels only in the variable */
:root {
  --primary: 0.205 0 0;  /* lightness chroma hue — no oklch() wrapper */
}
```

With v4, store the full `oklch(...)` value in the CSS variable and reference it directly via `var()`.
