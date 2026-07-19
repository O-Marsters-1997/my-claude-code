# Interview Question Bank

Pick 3–5 questions from this list that are genuinely unresolved after your Phase 1 characterisation.
Good questions to pick: ones where different answers would meaningfully change which ideas are
worth generating. Skip questions where you can make a confident assumption from what you observed.

Group them into 1–2 `AskUserQuestion` calls. Don't ask questions serially when they can be batched.

---

## About the user and their goals

- Who is the target user for this product? (If multiple user types, which is the primary one?)
- What does success look like in 12 months — more users, deeper usage by existing users, a specific
  capability unlocked, revenue, something else?
- Is there a persona or type of user you actively *don't* want to build for? (Constraints are as
  useful as goals.)

## About the competitive context

- Are there 1–2 tools you admire in this space — not necessarily direct competitors, but ones that
  do something right you'd want to emulate? What specifically do you admire?
- Are there tools you've explicitly decided *not* to be? What's the thing those tools do that you're
  running away from?
- Is there a category of feature request you keep getting that you've so far declined? Why?

## About strategy and constraints

- Is the goal to go wide (more use cases, more users) or deep (serve existing users better)?
- Are there parts of the current architecture you'd prefer ideas *not* to touch — things that are
  fragile, being replaced, or out of scope?
- What's the most painful thing users complain about today that isn't yet on the roadmap?
- Are there integration points (APIs, ecosystems, platforms) you want to be closer to, or have
  already started moving toward?

## About resources and risk appetite

- How would you describe the team's risk appetite right now — are you looking for safe bets that
  compound existing strengths, or are you willing to bet on something more ambitious?
- Is there a time horizon constraint — something that needs to be shipped within a given window?

---

## How to ask

Ask **one question at a time** via `AskUserQuestion`. Never batch. Each answer may shift what
the next best question is — batching wastes that signal and makes the conversation feel like
a form rather than a dialogue.

Use structured options for yes/no or multiple-choice framings. Leave open-ended strategy
questions as free-form (the "Other" fallback works for this).

Don't ask more than 5 questions total. If you're tempted to ask more, make the assumption
explicit in the report instead.
