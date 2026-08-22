---
name: code-review
description: >
  Use when the user wants to review a branch, a PR, work-in-progress changes, or asks to
  "review since X". Reviews changes since a fixed point (commit, branch, tag, or merge-base)
  along two parallel axes — Standards (does this follow the repo's documented coding standards?)
  and Spec (does this match what the originating issue asked for?) — reported side by side. On
  Greptile-enabled repos, adds a third axis that pulls the PR bot's findings forward and
  diagnoses each as fix-or-decline before the PR is raised. For whole-system design assessment
  rather than a diff, use architecture-review.
---

Review of the diff between `HEAD` and a fixed point the user supplies, along two axes:

- **Standards** — does the code conform to this repo's documented coding standards?
- **Spec** — does the code faithfully implement the originating issue / spec?

Both axes run as **parallel sub-agents** so they don't pollute each other's context, then this skill aggregates their findings.

On repos where Greptile is enabled, a third **Greptile** axis runs after those two — see step 5.

## Process

### 1. Pin the fixed point

Whatever the user said is the fixed point — a commit SHA, branch name, tag, `main`, `HEAD~5`, etc. If they didn't specify one, ask for it.

Capture the diff command once: `git diff <fixed-point>...HEAD` (three-dot, so the comparison is against the merge-base). Also note the list of commits via `git log <fixed-point>..HEAD --oneline`.

Before going further, confirm the fixed point resolves (`git rev-parse <fixed-point>`) and the diff is non-empty. A bad ref or empty diff should fail here — not inside two parallel sub-agents.

### 2. Identify the spec source

Look for the originating spec, in this order:

1. Issue references in the commit messages or branch name (`gro2-218`, `#123`, `Closes #45`) — fetch Linear issues with the Linear MCP tools, GitHub issues with `gh issue view`.
2. A path the user passed as an argument.
3. A spec file under `docs/`, `specs/`, or `.scratch/` matching the branch name or feature.
4. If nothing is found, ask the user where the spec is. If they say there isn't one, the **Spec** sub-agent will skip and report "no spec available".

### 3. Identify the standards sources

Anything in the repo that documents how code should be written, such as `CODING_STANDARDS.md` or `CONTRIBUTING.md`.

On top of whatever the repo documents, the Standards axis always carries the **smell baseline** below — a fixed set of Fowler code smells (_Refactoring_, ch.3) that applies even when a repo documents nothing. Two rules bind it:

- **The repo overrides.** A documented repo standard always wins; where it endorses something the baseline would flag, suppress the smell.
- **Always a judgement call.** Each smell is a labelled heuristic ("possible Feature Envy"), never a hard violation — and, like any standard here, skip anything tooling already enforces.

Each smell reads *what it is* → *how to fix*; match it against the diff:

- **Mysterious Name** — a function, variable, or type whose name doesn't reveal what it does or holds. → rename it; if no honest name comes, the design's murky.
- **Duplicated Code** — the same logic shape appears in more than one hunk or file in the change. → extract the shared shape, call it from both.
- **Feature Envy** — a method that reaches into another object's data more than its own. → move the method onto the data it envies.
- **Data Clumps** — the same few fields or params keep travelling together (a type wanting to be born). → bundle them into one type, pass that.
- **Primitive Obsession** — a primitive or string standing in for a domain concept that deserves its own type. → give the concept its own small type.
- **Repeated Switches** — the same `switch`/`if`-cascade on the same type recurs across the change. → replace with polymorphism, or one map both sites share.
- **Shotgun Surgery** — one logical change forces scattered edits across many files in the diff. → gather what changes together into one module.
- **Divergent Change** — one file or module is edited for several unrelated reasons. → split so each module changes for one reason.
- **Speculative Generality** — abstraction, parameters, or hooks added for needs the spec doesn't have. → delete it; inline back until a real need shows.
- **Message Chains** — long `a.b().c().d()` navigation the caller shouldn't depend on. → hide the walk behind one method on the first object.
- **Middle Man** — a class or function that mostly just delegates onward. → cut it, call the real target direct.
- **Refused Bequest** — a subclass or implementer that ignores or overrides most of what it inherits. → drop the inheritance, use composition.

### 4. Spawn both sub-agents in parallel

**Standards sub-agent prompt** — include:

- The full diff command and commit list.
- The list of standards-source files you found in step 3, **plus the smell baseline from step 3** pasted in full — the sub-agent has no other access to it.
- The brief: "Report — per file/hunk where relevant — (a) every place the diff violates a documented standard: cite the standard (file + the rule); and (b) any baseline smell you spot: name it and quote the hunk. Distinguish hard violations from judgement calls — documented-standard breaches can be hard, but baseline smells are always judgement calls, and a documented repo standard overrides the baseline. Skip anything tooling enforces. Under 400 words."

**Spec sub-agent prompt** — include:

- The diff command and commit list.
- The path or fetched contents of the spec.
- The brief: "Report: (a) requirements the spec asked for that are missing or partial; (b) behaviour in the diff that wasn't asked for (scope creep); (c) requirements that look implemented but where the implementation looks wrong. Quote the spec line for each finding. Under 400 words."

If the spec is missing, skip the Spec sub-agent and note this in the final report.

### 5. Greptile axis (gated)

Greptile reviews these repos on the PR, so its findings normally land after the push. `greptile review` runs the same review locally against the branch's base, which pulls them forward. This axis runs that review and diagnoses each finding.

**Gate.** Run this axis only when every check passes. On any failure, skip it, state the reason in one line, and carry on — a skipped Greptile axis must never block Standards or Spec:

1. `git rev-parse --show-toplevel` resolves to a path under `~/Documents/plain/`.
2. `git remote get-url origin` contains `support-app` or `services`. Match on the remote, not the directory name, so worktrees (`support-app-gro2-219-…`) are covered.
3. `command -v greptile` succeeds.
4. `greptile whoami` succeeds. If it reports not signed in, the skip line is: "Greptile axis skipped — not signed in, run `greptile login`."

A CLI error from the review itself (repo not enabled, index missing) is also a skip-with-reason. Never run `greptile init` — that writes to the repo, and enabling Greptile is the user's call.

**Sequencing.** Spawn this sub-agent only once the Standards and Spec sub-agents have both returned. It needs their findings to dedupe against, and its verdicts lean on what they already established.

**Greptile sub-agent prompt** — include:

- The fixed point, the diff command, and the commit list.
- The command to run: `greptile review --agent`, adding `-b <fixed-point>` when the fixed point is a branch. For a SHA or tag, omit `-b` and let Greptile use the repo's default base. `--agent` is Greptile's own plain-output-for-agents mode, so no JSON parsing is needed.
- The Standards and Spec findings in full, for dedupe.
- The brief: "Run the review, then per finding report: (a) the finding quoted, with file and line; (b) a verdict — **fix** or **decline**; (c) the reasoning, grounded in the code you read rather than in Greptile's wording; (d) for a fix, the specific change to make; for a decline, why it's a false positive or an accepted trade-off. Read the code before judging any finding — do not take Greptile at face value. Where a finding is already covered by the Standards or Spec findings above, mark it a duplicate and don't restate it. Do not edit any files. Under 500 words."

### 6. Aggregate

Present the reports under `## Standards`, `## Spec`, and — when it ran — `## Greptile` headings, verbatim or lightly cleaned. Do **not** merge or rerank findings — the axes are deliberately separate (see _Why two axes_).

End with a one-line summary: total findings per axis, and the worst issue _within each axis_ (if any). Don't pick a single winner across axes — that's the reranking the separation exists to prevent.

## Why two axes

A change can pass one axis and fail the other:

- Code that follows every standard but implements the wrong thing → **Standards pass, Spec fail.**
- Code that does exactly what the issue asked but breaks the project's conventions → **Spec pass, Standards fail.**

Reporting them separately stops one axis from masking the other.

## Why the Greptile axis

It is the PR bot's review pulled forward, so its findings get triaged on a local branch instead of as review comments after the push. It runs last because a finding the Standards or Spec axis already raised needs no second verdict, and because a diagnosis is better with those two reports in hand.
