# Wireframe style — fidelity and layout for stage 5

Governs every wireframe this skill renders. [open-design.md](open-design.md) covers the *mechanics* of getting a page to a PNG; this file decides what that page is allowed to look like. Where they disagree, this file wins.

## The line

**A wireframe communicates structure and intent. The moment it looks like a working screen, it is a prototype and it is wrong.**

Fidelity is a claim about certainty. A polished screen says "this is decided" and invites feedback on the padding; a sketch says "this is the shape, argue with it" and invites feedback on the idea — which is the only feedback worth having at this stage. Over-rendering a direction *costs* you the critique you rendered it for.

## What good looks like

Picture a designer's own sketch of the change, and copy that: hand-drawn boxes on a white page, a hand-style font, empty space left empty, one blue arrow pointing from the thing that changed to a handwritten note explaining why — and **exactly one idea on the page**. Everything not part of the idea is a rectangle or a ruled line. It reads in three seconds and could not be mistaken for a build.

```
┌─────────────────────────────────────────────┐
│  ┌──────┐  ┌─────────────────────────────┐  │
│  │      │  │  Next steps                 │  │
│  │      │  │  ─────────────────────      │  │
│  │      │  │  ▓▓▓▓▓▓░░░░░░░░░░░░░░░      │  │
│  │      │  │                             │  │
│  │      │  │  ☑ Connect your channels  › │  │
│  │      │  │  ☐ Invite your team       › │  │
│  │      │  │  ☐ Set up workflows       › │  │
│  └──────┘  └─────────────────────────────┘  │
│      ↖                                      │
│       ╲__ moves out of the corner card,     │
│           onto the main page                │
└─────────────────────────────────────────────┘
              2A · first action broken
```

That's the whole grammar: boxes for regions, rules for text, one annotated arrow for the intent, a label naming the direction and the problem. Whatever renders it, the output should look like that — not like a screenshot of the shipped product with one thing moved.

## Layout

- **One direction per frame. One frame per screen.** Never a grid of directions on a single canvas. Nine directions is nine images (or nine pages), not one board with nine cells — a board squeezes every idea until none of them is legible, which is the failure this rule exists to prevent.
- **Give each frame room.** Roughly a full viewport per screen. Empty space is not waste; it's what makes the structure readable.
- **Don't repeat screens.** If two directions share a screen that doesn't change between them, draw it once. Redraw it only when the *difference* is the point — and then annotate the difference.
- **Label every frame** with its direction ID (matching the ID used in `findings.md`) and the problem it solves, in that order. An unlabelled frame is unciteable.

## Fidelity

- **Greyscale only.** Boxes, lines, and placeholder rules. One accent colour is permitted, and only for annotations — never for UI.
- **No brand.** No product typography, no real colour palette, no logos, no imagery. A single neutral font for everything; hand-style if the renderer has one.
- **Placeholder over polish.** Grey blocks for images, ruled lines for body text, plain rectangles for controls. Real copy only where the *words themselves* are the design decision (a button label, a headline that has to be argued about) — otherwise a line is enough.
- **Nothing that looks interactive.** No hover states, shadows, gradients, focus rings, live inputs, or anything that reads as a working component. If a viewer might try to click it, it's too finished.

## Annotation

Annotations carry the intent, and are what separate a wireframe from a picture of a layout. Use hand-style notes and arrows in the accent colour:

- Point at what changed and say why in a few words ("links to the main page", "collapses when done").
- Mark behaviour that a static frame can't show — what appears on scroll, what happens on submit, what's conditional.
- Keep them in the margin or overlaid on empty space, never crowding the structure.

## Grounding

Draw from the audited surface's real code and layout (see [open-design.md](open-design.md) step 1) so the frame reads as *this* product being modified, not a generic app. Structure comes from reality; **fidelity comes from this file.** Grounding the layout in the real screen is not licence to reproduce the real screen's finish.
