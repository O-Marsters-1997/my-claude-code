# Responsive Design — Breakpoints and Layout

> **Solid mode:** breakpoints and the flex/grid discipline apply unchanged. In the examples, read `className` as `class` and `.map(...)`+`key` as `<For each>` — see [primitives-solid.md](./primitives-solid.md).

## Breakpoints

Tailwind's default breakpoints (mobile-first — each modifier means "at this width and above"):

| Modifier | Min-width | Typical target |
|----------|-----------|----------------|
| *(none)* | 0px       | All phones     |
| `sm:`    | 640px     | Large phones, small landscape |
| `md:`    | 768px     | Tablets        |
| `lg:`    | 1024px    | Laptops        |
| `xl:`    | 1280px    | Desktops       |
| `2xl:`   | 1536px    | Wide monitors  |

Write the **base (smallest) style first**, then override upward. Never write desktop-first and try to undo with `max-md:`.

```tsx
// Correct — mobile first
<div className="flex flex-col gap-4 md:flex-row md:gap-6 lg:gap-8">

// Wrong — desktop first, trying to undo at smaller screens
<div className="flex flex-row max-md:flex-col">
```

---

## Flex vs Grid — Decision Guide

**Default to `flex`.** It covers the majority of UI layout needs with less cognitive overhead.

### Use `flex` for:
- Navbars, toolbars, button groups
- Card footers and headers
- Form rows (label + input)
- Any layout that's essentially a line of things, possibly with wrapping
- Vertical stacks of content

```tsx
// Vertical stack
<div className="flex flex-col gap-4">

// Horizontal row, wraps on overflow
<div className="flex flex-wrap gap-3">

// Space between (nav pattern)
<nav className="flex items-center justify-between px-6 py-4">

// Centred content
<div className="flex items-center justify-center min-h-screen">
```

### Use `grid` only when:
- You need **rows AND columns** to both carry meaning simultaneously
- The layout is a genuine two-dimensional grid (calendar, image gallery, dashboard tile layout)
- You need subgrid alignment across siblings

```tsx
// Calendar — both axes matter
<div className="grid grid-cols-7 gap-px">

// Dashboard tiles — 2D
<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">

// Data table with pinned columns — genuine 2D
<div className="grid grid-cols-[200px_1fr_1fr_100px]">
```

### Common mistake: grid used as a substitute for flex

```tsx
// Bad — this is 1D, should be flex
<div className="grid grid-cols-2 gap-4">
  <label>Name</label>
  <input />
</div>

// Good
<div className="flex items-center gap-4">
  <label>Name</label>
  <input className="flex-1" />
</div>
```

---

## Common Responsive Patterns

### Sidebar layout

```tsx
<div className="flex flex-col lg:flex-row min-h-screen">
  <aside className="w-full lg:w-64 shrink-0">...</aside>
  <main className="flex-1 min-w-0 p-6">...</main>
</div>
```

### Card grid

```tsx
<div className="flex flex-wrap gap-4">
  {items.map(item => (
    <div key={item.id} className="w-full sm:w-[calc(50%-0.5rem)] lg:w-[calc(33.333%-0.667rem)]">
      <Card>...</Card>
    </div>
  ))}
</div>
```

Or if the items are genuinely equal-column tiles, grid is appropriate:

```tsx
<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
  {items.map(item => <Card key={item.id}>...</Card>)}
</div>
```

### Hero section

```tsx
<section className="flex flex-col items-center gap-6 px-4 py-16 text-center md:py-24 lg:flex-row lg:text-left lg:gap-12">
  <div className="flex-1">...</div>
  <div className="w-full max-w-md lg:max-w-lg">...</div>
</section>
```

### Form layout

```tsx
// Single column on mobile, two columns on tablet+
<div className="flex flex-col gap-4 md:grid md:grid-cols-2 md:gap-6">
  <Field>...</Field>
  <Field>...</Field>
  <Field className="md:col-span-2">...</Field>  {/* full-width row */}
</div>
```

Note: grid is acceptable here because both rows and columns matter for the form structure.

### Responsive text

```tsx
<h1 className="text-2xl font-bold sm:text-3xl lg:text-4xl">
```

### Responsive padding/container

```tsx
<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
```

---

## Hiding and Showing Elements

```tsx
// Hidden on mobile, visible from md
<div className="hidden md:block">

// Visible on mobile only
<div className="block md:hidden">

// Flex on large screens only
<div className="hidden lg:flex items-center gap-4">
```

Avoid `invisible` for responsive hiding — it hides the element but still occupies space. Use `hidden` / `block` / `flex`.
