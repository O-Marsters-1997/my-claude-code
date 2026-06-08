# CVA and cn — Variant Patterns

> **Solid mode:** the `cva`/`cn` distinction and `tailwind-merge` conflict resolution are identical in Solid — `cn()` works the same. The *component-authoring* examples below are React (`React.*` types, `Slot`/`asChild`, destructured `className`, `{...props}`); for the Solid equivalents (`splitProps`, `props.class`, `<Dynamic>`) see [primitives-solid.md](./primitives-solid.md). Read `className` as `class` throughout.

## The Core Distinction

| Tool | Where it lives | What it does |
|------|---------------|--------------|
| `cva` | Component definition | Declares the full variant matrix — what props the component accepts and which Tailwind classes each combination resolves to |
| `cn` | Call sites and component internals | Merges class strings, resolves Tailwind conflicts (last wins), handles conditional application |

They solve different problems and always work together. `cva` tells you *what* a component can be; `cn` handles *how* those classes get combined with external overrides.

---

## `cva` — Class Variance Authority

`cva` creates a function that takes variant props and returns a resolved className string. Use it whenever a component has more than one visual mode.

```tsx
import { cva, type VariantProps } from "class-variance-authority"

const buttonVariants = cva(
  // Base classes — always applied
  "inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default:     "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline:     "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
        secondary:   "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost:       "hover:bg-accent hover:text-accent-foreground",
        link:        "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm:      "h-9 rounded-md px-3",
        lg:      "h-11 rounded-md px-8",
        icon:    "size-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

// Surface the variant props on the component's TypeScript interface
interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

function Button({ className, variant, size, asChild = false, ...props }: ButtonProps) {
  const Comp = asChild ? Slot : "button"
  return (
    <Comp
      className={cn(buttonVariants({ variant, size }), className)}
      {...props}
    />
  )
}
```

### Compound variants

When a combination of two variants needs its own class set beyond what each individually provides:

```tsx
const alertVariants = cva("relative rounded-lg border p-4", {
  variants: {
    variant: {
      default:     "bg-background text-foreground",
      destructive: "border-destructive/50 text-destructive",
    },
    size: {
      default: "text-sm",
      lg:      "text-base",
    },
  },
  compoundVariants: [
    {
      variant: "destructive",
      size: "lg",
      class: "font-semibold",  // only when both destructive AND lg
    },
  ],
  defaultVariants: {
    variant: "default",
    size: "default",
  },
})
```

### When to use `cva`

- The component has 2+ visual variants controlled by props
- You want TypeScript to enforce valid prop values
- You're adding a new variant to an existing shadcn component

### When NOT to use `cva`

- A one-off component that will never have variants — just use a plain `cn()` call or a className string
- You only need conditional classes at the call site — that's `cn`'s job

---

## `cn` — Class Merging Utility

`cn` is a thin wrapper around `clsx` + `tailwind-merge`. It does two things:
1. Conditional class application (via `clsx`): `cn("base", isActive && "bg-primary")` 
2. Tailwind conflict resolution (via `tailwind-merge`): `cn("p-4", "p-2")` → `"p-2"` (last wins)

```ts
// Standard implementation in lib/utils.ts
import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

### Patterns

**Conditional classes:**
```tsx
// Instead of:
<div className={`flex ${isActive ? "bg-primary text-primary-foreground" : "bg-muted"}`}>

// Use:
<div className={cn("flex", isActive ? "bg-primary text-primary-foreground" : "bg-muted")}>
```

**Multiple conditions:**
```tsx
<div
  className={cn(
    "flex items-center rounded-md px-3 py-2 text-sm",
    isActive && "bg-accent text-accent-foreground",
    isDisabled && "opacity-50 pointer-events-none",
    className,  // always pass through the caller's className last
  )}
>
```

**Merging with cva output:**
```tsx
// The className prop always goes last so callers can override
<button className={cn(buttonVariants({ variant, size }), className)} />
```

**Conflict resolution (tailwind-merge):**
```tsx
cn("px-4", "px-2")           // → "px-2"     (later wins)
cn("text-sm", "text-lg")     // → "text-lg"
cn("bg-red-500", "bg-blue-500") // → "bg-blue-500"
cn("p-4", "pt-2")            // → "p-4 pt-2" (different axis, both kept)
```

### When to use `cn`

- Any conditional class application
- Combining the output of `cva` with a caller-provided `className`
- Merging several class strings that might conflict
- Any time you'd otherwise write a template literal in `className`

---

## Adding a Variant to an Existing Component

When a visual pattern is used in more than one place and no built-in variant covers it, add it to the component source:

```tsx
// Before: scattered one-off className overrides across the codebase
<Badge className="bg-success text-success-foreground">Saved</Badge>

// After: add 'success' to Badge's variant map
const badgeVariants = cva("...", {
  variants: {
    variant: {
      default:     "...",
      secondary:   "...",
      destructive: "...",
      outline:     "...",
      success:     "bg-success text-success-foreground",  // ← add here
    },
  },
})

// Then use cleanly everywhere:
<Badge variant="success">Saved</Badge>
```

The semantic color token (`--success`) must exist in the theme first. See SKILL.md → Adding New Theme Tokens.
