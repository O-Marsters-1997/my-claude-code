---
name: code-review
description: >
  Use when the user wants to review a branch, a PR, work-in-progress changes, or asks to
  "review since X". Reviews changes since a fixed point (commit, branch, tag, or merge-base)
  along two parallel axes — Standards (does this follow the repo's documented coding standards?)
  and Contract (does this match the ticket and the feature's frozen contract?) — reported side by side.
  Also defines the Journey axis, which /fleet reconcile applies to a completed slice. For
  whole-system design assessment rather than a diff, use architecture-review.
---

Review of the diff between `HEAD` and a fixed point the user supplies, along two axes (a third, Journey, checks a delivery slice rather than a diff; see _Journey axis_):

- **Standards** — does the code conform to this repo's documented coding standards?
- **Contract** — does the code meet the ticket's acceptance criteria, implement every stub the ticket lists, and leave the contract's signatures untouched?

Both diff axes run as **parallel sub-agents** so they don't pollute each other's context, then this skill aggregates their findings.

## Process

### 1. Pin the fixed point

Whatever the user said is the fixed point — a commit SHA, branch name, tag, `main`, `HEAD~5`, etc. If they didn't specify one, ask for it.

Capture the diff command once: `git diff <fixed-point>...HEAD` (three-dot, so the comparison is against the merge-base). Also note the list of commits via `git log <fixed-point>..HEAD --oneline`.

Before going further, confirm the fixed point resolves (`git rev-parse <fixed-point>`) and the diff is non-empty. A bad ref or empty diff should fail here — not inside two parallel sub-agents.

### 2. Identify the contract and spec source

Look for the originating ticket and contract, in this order:

1. Issue references in the commit messages or branch name (`#123`, `Closes #45`) — resolve with `gh issue view`.
2. A path the user passed as an argument.
3. A spec file under `docs/`, `specs/`, or `.scratch/` matching the branch name or feature.
4. If nothing is found, ask the user where the spec is. If they say there isn't one, the **Contract** sub-agent will skip and report "no spec available".

Then look for the feature contract: take the ticket's feature label and read `docs/plans/<feature>.md`, which lists the stubs and their signatures. Get the feature branch from `gh pr view --json baseRefName`, falling back to the branch the user named as the fixed point. The ticket's stub list is in its issue body. `docs/plans/` is gitignored, so a fresh checkout may lack the file: say so in the report rather than silently dropping the contract check. If the ticket has no feature label or the file is missing (ad hoc work, sweep, non-code), there is no contract and the Contract axis falls back to plain spec review: acceptance criteria only, as below.

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

**Contract sub-agent prompt** — include:

- The diff command and commit list.
- The path or fetched contents of the ticket (or other spec), and of the contract file when one exists, plus the feature branch name.
- The brief: "Report: (a) acceptance criteria or requirements that are missing, partial, or look implemented but wrong; (b) behaviour in the diff that wasn't asked for (scope creep). Quote the ticket line for each finding. With a contract also report: (c) every stub the ticket lists that the diff does not implement (still a stub body, or absent); (d) signature diff: compare each stubbed declaration (name, parameters, return types, visibility) in the diff against the same declaration on the feature branch (`git show <feature-branch>:<path>`), and flag every changed signature that is not tied to a `contract-change` issue (match `gh issue list --label contract-change --state all` against the issue numbers in the commits and PR body, or an issue naming the changed symbol). A signature edit with no such issue is a violation, not a judgement call. Under 400 words."

If neither a ticket nor a spec is available, skip the Contract sub-agent and note this in the final report.

### 5. Aggregate

Present the reports under `## Standards` and `## Contract` headings (Contract covers acceptance criteria, stub coverage and the signature diff; without a contract, acceptance criteria only), verbatim or lightly cleaned. Do **not** merge or rerank findings — the axes are deliberately separate (see _Why separate axes_).

End with a one-line summary: total findings per axis, and the worst issue _within each axis_ (if any). Don't pick a single winner across axes — that's the reranking the separation exists to prevent.

## Journey axis

A third axis, separate from Standards and Contract. It does not review a diff: it runs once per
delivery slice, on integrated code, when `/fleet reconcile` finds the slice's last ticket merged.
Run it as its own sub-agent, never inside the two above, so a passing diff cannot mask a broken journey.

Input is one PRD journey (title, surface, steps, final `→ <observable state>`) and the evidence
gathered for it: the output of running its steps. Missing or empty evidence counts as needs human
eyes, never pass. Report, per journey:

- **Pass**: the evidence shows the `→` end state. Quote the evidence line that shows it.
- **Fail**: the evidence contradicts the end state or stops before it. Quote the step where it diverged.
- **Needs human eyes**: the end state is visual UI, or the evidence cannot show it. List what the user
  should look at. Never guess pass or fail from output that cannot show the state.

## Why separate axes

A change can pass one axis and fail another:

- Code that follows every standard but implements the wrong thing → **Standards pass, Contract fail.**
- Code that does exactly what the issue asked but breaks the project's conventions → **Contract pass, Standards fail.**

Reporting them separately stops one axis from masking another.
