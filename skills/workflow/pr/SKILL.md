---
name: pr
description: "Write a pull request's title and body and open it as a draft. Use whenever a PR is about to be opened or its description written: 'open a PR', 'raise a PR', 'create the pull request', 'write the PR body', 'gh pr create', or when a ticket or /implement run finishes with a PR for review. Produces a short, visual-first body (Summary, Evidence, Merge Danger, Closes lines) built for a reviewer deciding whether to merge agent-written work. Not for /fleet reconcile's feature-branch PR into main, which only lists tickets already reviewed on their own PRs."
metadata:
  credits:
    - skill: pr
      author: Matt Pocock
      url: "https://github.com/mattpocock/skills/blob/main/skills/engineering/pr/SKILL.md"
    - skill: show-me
      author: Dex Horthy
      organisation: Humanlayer
      url: "https://github.com/humanlayer/skills/blob/main/plugins/show-me/skills/show-me/SKILL.md"
---

The reader is the person who merges agent-written work. They want to decide
whether to trust and merge it without reading the diff first, and in as little
time as possible. Every line you add costs them reading time, so the body is
the smallest thing that answers: what changed, does it work, how risky is it.

## Template

```markdown
## Summary

<visual, where relevant>

<one line of prose>

## Evidence

- **Before:** <one line or screenshot>
- **After:** <one line or screenshot>

## Merge Danger

**Door:** <one-way | two-way>
**Blast radius:** <one phrase>

**Look at:** <the file or decision that most deserves human attention>

Closes #<N>

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

Omit any section that has nothing real to say rather than filling it. No
preamble, no restating the title, no lists of files touched.

## Summary

Lead with a visual when the change has a shape worth seeing, then one line of
prose stating the intent. A trivial change gets the prose line alone.

- **UI changes:** screenshots of the changed screen. This beats any diagram.
- **Everything else:** at most one or two of the views below, the smallest that
  makes the point. Keep only the calls, files, states and boundaries the
  reviewer needs.

Pseudocode for logic:

```text
on(save)
  if content is unchanged
    return cached result
  write new content
```

Call tree for runtime flow, component tree for UI structure, shallow file tree
for a broad refactor, Mermaid for interaction between parts:

```text
submitForm
  createSession
    persistPrompt
    launchAgent
```

A `diff` sketch when the surrounding shape already exists and the point is what
changed. Match the sketch to the topic (call tree, file tree, component tree,
control flow):

```diff
 submitForm
   createSession
     persistPrompt
+    expandSkillMention
     launchAgent
```

## Evidence

Show proof the change works, one line per side. Screenshots for visual
changes; otherwise the exact test, command or skill run with its result before
and after. For skill or prose-only changes, a one-line note of the invocation
you ran and what it produced is enough.

If nothing could be run, write `Not verified: <why>` in place of the two lines.
Never describe a check you did not run. A reviewer who trusts false evidence
merges a broken change, which is worse than one who knows to check.

## Merge Danger

- **Door:** two-way if a revert fully undoes it; one-way if it destroys data,
  migrates state, publishes something, or changes a contract others depend on.
- **Blast radius:** one phrase naming who or what breaks if this is wrong
  (e.g. `fleet dispatch only`, `every skill's frontmatter`, `none`).

Drop the whole section for docs, typos and other changes with no runtime
effect.

## Look at

One optional line pointing the reviewer at the part that most needs a human:
a judgement call, an unusual pattern, a spot you are least sure of. Leave it
out when nothing stands out.

## Closes

One `Closes #<N>` line per ticket the PR resolves, so GitHub closes them on
merge. A bundle covering several tickets repeats the line for each.

## Opening the PR

1. Match the title to the repo's recent convention
   (`gh pr list --state merged --limit 5 --json title`).
2. Open it as a draft:

   ```bash
   gh pr create --draft --title "<title>" --body "<body>"
   ```

   Add `--base <branch>` only when you were given a base; otherwise use the
   repo default.
3. Print the PR URL.
