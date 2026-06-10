# CURRICULUM.md Format

`CURRICULUM.md` lives at the workspace root. It is the provisional table of contents for the course — the ordered list of planned lessons, each with a one-line summary, plus the chosen depth level. It is created the first time a subject is taught and revised as the course evolves. Every planned lesson should trace back to [[MISSION.md]].

## Depth levels

When first proposing the curriculum, offer the user one of three depth levels. The level sets how many lessons there are and how tightly each is scoped:

- **Overview** — fewer, broader lessons that group related ideas together. A quick pass over the subject for someone who wants the shape of it fast.
- **Standard** — balanced scope. The default unless the user signals otherwise.
- **Deep** — many tightly-scoped lessons, each teaching one small thing. Slower but thorough; best when the user wants mastery.

Record the chosen level so future sessions stay consistent with it.

## Template

```md
# Curriculum: {Topic}

Depth: {Overview | Standard | Deep}

## Lessons
1. {dash-case-slug} — {one-line summary of the one thing this lesson teaches}
2. {dash-case-slug} — {…}
3. {…}

## Notes
- {Optional: open questions, sections deliberately deferred, or amendments the user requested}
```

## Rules

- **Provisional, not fixed.** The curriculum is a starting proposal. Let the user confirm or amend it before building any lessons, and revise it freely as the course progresses — a learning record or a shifted mission may reorder or replace planned lessons.
- **One line per lesson.** Each entry is a summary, not a lesson. If a summary needs a paragraph, the lesson is probably too broad — split it (or, at Overview depth, that breadth may be intentional).
- **Lesson 1 is the template.** Build lesson 1 alone first and confirm its structure and styling with the user. The rest of the lessons are then batch-created to match it, so styling fixes are made once and inherited.
- **Numbering tracks lessons.** Curriculum entry _n_ corresponds to lesson file `000n-slug.html`. Keep them aligned so the user can open a lesson by its number.
- **Keep it short.** If the curriculum runs long, the depth level is probably too fine — it should fit on a screen as a scannable map of the course.
```

