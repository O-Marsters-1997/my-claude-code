# Brief: `pr` skill (for skill-creator)

## Purpose

Write the title and body for a pull request and open it as a draft. Adapted
from Matt Pocock's `pr` skill
(https://github.com/mattpocock/skills/blob/main/skills/engineering/pr/SKILL.md),
which is itself built on Dex Horthy's show-me skill.

## Name and trigger

Name: `pr`. The description alone must make it trigger whenever a PR is to be
written or opened ("opening a PR", "write the PR body", "create the PR").
Other skills must not name it.

## Reader

The user, as merge gatekeeper, reviewing agent-written PRs. The description
must let them decide whether to trust and merge without reading the diff first.

## Hard rules

- Brevity comes first: no preamble, no filler.
- A section with nothing to say is omitted, not filled.
- Every non-prose element is the smallest one that makes the point.

## Body template

```markdown
## Summary

<visual first, where relevant>
<one line of prose>

## Evidence

- Before: <one line or screenshot>
  After: <one line or screenshot>

## Merge Danger

Door: one-way | two-way
Blast radius: <one phrase>

Look at: <file or decision that deserves human attention>   (optional)

Closes #N
Closes #M

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## Section rules

- **Summary.**
  - Visual first where relevant, then one line of prose.
  - UI changes get screenshots.
  - Otherwise a diagram only if needed. Use Pocock's menu: pseudocode, call
    tree, component tree, file tree, Mermaid, diff-sketch, full block. Choose
    the smallest, never all.
  - Trivial changes get prose only.
- **Evidence.**
  - A real before/after: test or command output, or a transcript of the skill
    or command actually run.
  - If nothing could be run, write "Not verified" and say why. Never invent
    proof.
  - One line per side. Screenshots are best for visual changes.
- **Merge Danger.**
  - Door and Blast radius, one line each.
  - Omit the whole section for trivial PRs (docs, typos).
- **Review focus.** Optional "Look at:" one-liner.
- **Closes.** Always a `Closes #N` line per ticket; in fleet bundles, repeat
  each ticket's line.
- **Footer.** Always keep the Claude Code attribution footer.
- **Hard caps.** Summary prose: one line. Evidence: one line per side. Door and
  Blast radius: one line each.

## Action

- Write title and body, then `gh pr create --draft`.
- Add `--base` only when told one (fleet prompts).
- Follow the repo's existing title convention.

## Out of scope

Out-of-scope/follow-up sections, a `GLOSSARY.md` dependency (the repo has
none), merging.

## Credit

Include a `metadata.credits` block like Pocock's, crediting Matt Pocock and Dex
Horthy (show-me).

## Follow-up repo change (separate from the skill)

- `/implement` finishes by raising a PR for review but names no PR skill; the
  skill's own description makes it trigger.
- `/fleet reconcile`'s feature-branch PR into main does not use the skill: it
  is not a draft and only lists the already-reviewed tickets.
