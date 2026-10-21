---
name: repo-skill-candidates
description: >
  Audits one codebase for the recurring, multi-step patterns worth turning into repo-level
  Claude Code skills, and writes a short evidence-backed report to skill-candidates-report.md.
  Evidence comes only from the code (source, layout, tests, CI, scripts, docs, existing .claude
  setup), so it works on a new or inherited repo with no history. Triages every pattern to
  skill, CLAUDE.md line, linter config or script, and drops what existing skills already cover.
  Use when the user says "find skill candidates", "what in this repo should be a skill",
  "audit this repo for skill opportunities", "which patterns here are worth capturing as
  skills", or "repo-skill-candidates". Report only. Not for writing or drafting a skill (use
  skill-creator), reviewing code quality (use code-review or architecture-review), or
  suggesting skills from session history or the user's own habits (use reflect).
---

# Repo Skill Candidates

You answer one question: **what does this repo keep doing that a skill would help Claude do
right?** You produce a report and nothing else.

## Hard limits

- **Report only.** The single file you write is `skill-candidates-report.md` at the repo root.
  No SKILL.md drafts, templates, frontmatter sketches or scaffolding. No edits to code, config,
  CLAUDE.md or anything else.
- **Code is the only evidence.** Source, directory structure, tests, CI config, build and deploy
  scripts, docs, and the existing `.claude/` setup. Never read session transcripts, conversation
  history, `~/.claude/projects/`, memory files or reflect ledgers as evidence.
- **Ownership rule.** If a pattern's evidence is in the code, it is yours. If it would only show
  up in conversations or corrections, it is out of scope: leave it to reflect.
- **No generic advice.** Every candidate cites file paths in this repo. "Projects like this
  usually benefit from X" is not a candidate.

## Step 1: Survey

Establish the shape of the repo before hunting for patterns.

1. Run `git ls-files | wc -l` and count files by extension to find the languages.
2. Read the manifests (`package.json`, `go.mod`, `pyproject.toml`, `Cargo.toml`, `Gemfile`,
   `pom.xml` and similar) for frameworks and tooling.
3. Map the top two levels of the directory tree.
4. Locate tests, CI (`.github/workflows/`, `.gitlab-ci.yml`, `.circleci/`), build and deploy
   scripts (`Makefile`, `justfile`, `scripts/`, `bin/`, `Dockerfile`, IaC), migrations,
   codegen inputs, and docs (`README`, `CONTRIBUTING`, `docs/`).
5. Note linter and formatter configs already in place: they decide later whether a style
   pattern is already enforced.

**Sampling.** On a large repo, sample rather than read everything: up to about 50 files per
language, spread across top-level directories and weighted towards the most-changed
(`git log --format= --name-only | sort | uniq -c | sort -rn | head -100`). Read every CI file,
script and doc file in full; they are small and dense with procedure. Record exactly what you
sampled and what you skipped: the report must say so.

When a repo has several large languages or services, you may fan out one `Explore` agent per
area with the categories in `references/pattern-catalogue.md`, asking each for patterns with
counts and paths. Merge and verify their findings yourself before scoring.

## Step 2: Find recurring patterns

Work through `references/pattern-catalogue.md`. You are looking for **multi-step procedures that
repeat**, not single facts. The tell is a change that has to touch several places in a fixed
order, e.g. a new endpoint needs a handler, a route registration, a schema, a test and a docs
entry.

For each pattern record:

- **What it is**, in one line, as a task ("add a new background job").
- **The steps**, as observed across instances: which files change, in what order.
- **Prevalence**: how many instances, counted with grep or glob, not estimated. Say how you
  counted.
- **Evidence**: 2–3 concrete paths to good instances.
- **Inconsistency**: instances that skip a step or do it differently (a route with no test,
  a migration with no down step, a registry entry missing). These are the strongest signal
  that a skill would pay off; cite them.

Reconstructing a procedure from git history (`git log --stat` on commits that added a
similar file) is fair: commits are code. Commit messages are evidence; PR discussion is not.

## Step 3: Coverage check

Read what already guides Claude in this repo:

- `.claude/skills/*/SKILL.md`, `.claude/rules/`, `.claude/commands/`, `.claude/agents/`
- `CLAUDE.md` and `AGENTS.md` at the root and in subdirectories
- `~/.claude/skills/*/SKILL.md` if accessible (frontmatter descriptions are enough unless one
  looks close to a candidate)

These are read for **coverage**, not as evidence of patterns. Some may have come from reflect;
treat them like any other existing guidance.

For each candidate:

- **Fully covered**: an existing skill or rule already tells Claude the procedure. Move it to
  "Already covered" with the file that covers it.
- **Partially covered**: an existing skill handles the area but misses steps the code shows.
  Keep it, marked "extend `<skill>`", and name the missing steps.
- **Not covered**: new candidate.

If existing guidance contradicts what the code actually does, note it under Limitations. Do not
fix it.

## Step 4: Triage

This is the step that matters most. For every pattern, pick exactly one home:

| Verdict | Fits when |
|---|---|
| **Skill** | Multi-step, needs judgement or repo context (choosing names, deciding which variant applies, knowing the gotchas), and Claude would plausibly get it wrong or incomplete without guidance. |
| **CLAUDE.md / rules line** | A fact or convention that fits in one or two sentences ("errors are wrapped with `pkg/errs.Wrap`", "tests live beside the code"). |
| **Linter / formatter config** | A style rule a tool can enforce deterministically (import order, naming case, quote style, banned calls). |
| **Script or hook** | A fixed sequence with no judgement (regenerate code after editing the schema, run three commands in order). Note when an existing script already does it. |

Tests for a Skill verdict, all of which should hold:

1. It takes three or more steps, or touches three or more files.
2. At least one step needs a decision a script cannot make.
3. Writing it as a rule would take more than a short paragraph.

If a pattern is mostly deterministic with one judgement call, prefer script plus a rules line
and say so. A repo whose only patterns are formatting conventions should yield few or no skill
recommendations: that is a correct result, not a failure. Never pad.

Only Skill verdicts are recommendations. Everything else goes in "Belongs elsewhere", one line
each, so nothing is silently dropped.

## Step 5: Score

Score each Skill-verdict candidate 1–3 on four axes:

| Axis | 1 | 2 | 3 |
|---|---|---|---|
| **Prevalence** | 2–3 instances | 4–10 | more than 10, or a whole layer |
| **Complexity** | 3 steps, few gotchas | 4–6 steps or one real gotcha | 7+ steps or several gotchas |
| **Error-proneness** | no inconsistency found | one or two divergent instances | repeated omissions or divergence |
| **Future frequency** | rare (yearly-ish) | monthly-ish | most feature work touches it |

Total out of 12. Judge future frequency from the code: how fast the relevant directory grows in
git history, how central the layer is.

Drop anything scoring below 7, or anything you could not back with at least two real instances.
State your confidence (high / medium / low) on each kept candidate; low-confidence ones go in
"Other candidates" at most, never in the top recommendations.

## Step 6: Report

Write `skill-candidates-report.md` at the repo root in this structure. Lead with what would make
the biggest difference. Top recommendations: 2–3. Full recommended list (top plus other): about
7 at most.

```markdown
# Skill candidates: <repo name>

_Generated <date> by repo-skill-candidates. Evidence from code only._

## Repo summary
- Stack: <languages, frameworks, test tooling, CI>
- Sampled: <what was read, per language, and how it was chosen>
- Existing guidance: <skills, CLAUDE.md, AGENTS.md, rules, commands found>

## Top recommendations

### 1. `<proposed-skill-name>` (new | extends `<skill>`), score N/12, confidence <level>
- **Purpose:** <one line>
- **Trigger:** <the task that would invoke it, as a user would phrase it>
- **Evidence:** <prevalence count and how counted>; `<path>`, `<path>`, `<path>`
- **Inconsistencies found:** <paths that skip or diverge, or "none">
- **Why a skill, not a rule or script:** <the judgement it needs>
- **Score:** prevalence N, complexity N, error-proneness N, frequency N

## Other candidates
<same fields, two or three lines each>

## Belongs elsewhere
- **CLAUDE.md:** <pattern>: <one-line convention> (`<path>`)
- **Linter:** <pattern>: <tool and rule> (`<path>`)
- **Script/hook:** <pattern>: <what it would automate> (`<path>`)

## Already covered
- <pattern>: covered by `<skill or file>`

## Limitations
- <what was not sampled, low-confidence areas, guidance that contradicts the code>
```

Leave a section out only by writing "None found." under its heading, so the reader knows it was
checked.

Then summarise in chat: the top 2–3 candidates with one line each and their scores, the count
in each other section, and the report path. Do not offer to write the skills; the user takes
that step deliberately.

## Before you finish

- `git status` shows only `skill-candidates-report.md` as new or changed by you.
- Every recommendation cites real paths that exist (`ls` them).
- Every prevalence number came from a count you ran.
- No candidate rests on session history, memory or anything outside the repo and its
  `.claude/` setup.
