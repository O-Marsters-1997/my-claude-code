# Agent Skills

A collection of agent skills that extend capabilities across planning, development, and tooling.

## CLI Reference

| Command | Purpose |
|---|---|
| `npx skills add <source>` | Install a skill |
| `npx skills list` | List all installed skills |
| `npx skills find [query]` | Search/browse available skills |
| `npx skills remove [skill]` | Remove an installed skill |
| `npx skills check` | Check installed skills for updates |
| `npx skills update` | Update all installed skills |
| `npx skills init [name]` | Create a new SKILL.md template |

Browse available skills at [skills.sh](https://skills.sh). Common flags: `-g` (global), `-y` (skip prompts).

To list all skills available in this repo:

```
npx skills add O-Marsters-1997/skills --list
```

## Product workflow suite

Eight skills that take work from a raw idea through to tickets on a board. See
[`skills/workflow/README.md`](skills/workflow/README.md) for the map. Install the whole suite at
once — the skills call each other, so a partial install leaves broken handoffs:

```
for s in artifact-scan ideate chat-to-approach to-roadmap to-prd to-plan to-tickets ticket-tracker; do
  npx skills add O-Marsters-1997/my-claude-code --skill "$s" -g -y
done
```

| Skill | Level | Takes | Produces |
|---|---|---|---|
| `artifact-scan` | — | nothing | a report + one routing recommendation |
| `ideate` | portfolio | the codebase | `./ideas/reports/YYYY-MM-DD-ideate.md` |
| `chat-to-approach` | portfolio | a pasted conversation | `./docs/approach.md` |
| `to-roadmap` | portfolio | `./docs/approach.md` | `./ideas/roadmap.html` (cards are features) |
| `to-prd` | feature | one feature | `./docs/prd-<feature>.md` + a `[PRD]` issue |
| `to-plan` | feature | a PRD | `./docs/plans/<feature>.md` |
| `to-tickets` | feature | a plan | GitHub issues |
| `ticket-tracker` | feature | GitHub issues | `status:*` label moves |

The suite calls three routines that are not stages. These sit at depth 2, so they install normally:

```
for s in grilling source-synthesis triage-issue; do
  npx skills add O-Marsters-1997/my-claude-code --skill "$s" -g -y
done
```

## Skill maintenance

`skill-updater` verifies and ships every change to a skill, and `skill-feedback-collector` turns feedback into issues it can take. Install both:

```
for s in skill-updater skill-feedback-collector; do
  npx skills add O-Marsters-1997/my-claude-code --skill "$s" -g -y
done
```

## Planning & Design

These skills help you think through problems before writing code.

- **grilling** — The interview loop itself: rounds of questions along the frontier of the design tree, each with a recommended answer. Model-invoked; the suite delegates here.

  ```
  npx skills add O-Marsters-1997/skills --skill grilling
  ```

- **grill-me** — Slash-command front door to `grilling`. `/grill-me`.

  ```
  npx skills add O-Marsters-1997/skills --skill grill-me
  ```

- **grill-me-with-docs** — Grilling that also challenges your plan against `CONTEXT.md`, sharpens terminology, and writes glossary entries and ADRs inline. `/grill-me-with-docs`.

  ```
  npx skills add O-Marsters-1997/skills --skill grill-me-with-docs
  ```

## Development

These skills help you write, refactor, and fix code.

- **code-review** — Review a diff since a fixed point on two axes in parallel — Standards (repo conventions plus a Fowler smell baseline) and Spec (does it match the originating issue) — plus a third Greptile axis on Greptile-enabled repos that triages the PR bot's findings before you push.

  ```
  npx skills add O-Marsters-1997/skills --skill code-review
  ```

- **tdd** — Test-driven development with a red-green-refactor loop. Builds features or fixes bugs one vertical slice at a time.

  ```
  npx skills add O-Marsters-1997/skills --skill tdd
  ```

- **triage-issue** — Investigate a bug by exploring the codebase, identify the root cause, and file a GitHub issue with a TDD-based fix plan.

  ```
  npx skills add O-Marsters-1997/skills --skill triage-issue
  ```

- **feedback** — Turn loose feedback about how the system behaves into verified tickets: split it into points, check each against the code and recorded intent, play back the findings until you agree (silence counts as agreement), then file one ticket per finding under a `feedback:<area>` label for `/implement` or `/fleet dispatch`.

  ```
  npx skills add O-Marsters-1997/my-claude-code --skill feedback
  ```

- **improve-codebase-architecture** — Explore a codebase for architectural improvement opportunities, focusing on deepening shallow modules and improving testability.

  ```
  npx skills add O-Marsters-1997/skills --skill improve-codebase-architecture
  ```

- **codebase-design** — Terms and principles for designing deep modules: module, interface, depth, seam, adapter, and the deletion test. Also covers how to deepen a cluster of modules and how to design an interface several ways in parallel. `improve-codebase-architecture` depends on it. Vendored from [mattpocock/skills](https://github.com/mattpocock/skills/tree/main/skills/engineering/codebase-design), without the interface-design question list.

  ```
  npx skills add O-Marsters-1997/my-claude-code --skill codebase-design
  ```

- **tailwind-design-system** — Tailwind v4 design system architecture: CSS-first `@theme` config, the brand/semantic/component token hierarchy, OKLCH colour, CVA variants, native dark mode, and the v3-to-v4 migration checklist. Vendored from [wshobson/agents](https://github.com/wshobson/agents/tree/main/plugins/frontend-mobile-development/skills/tailwind-design-system) (MIT); the description is rewritten as a trigger clause and scoped to token work so it does not collide with `tailwind-shadcn`.

  ```
  npx skills add O-Marsters-1997/my-claude-code --skill tailwind-design-system
  ```

- **scaffold-exercises** — Create exercise directory structures with sections, problems, solutions, and explainers.

  ```
  npx skills add O-Marsters-1997/skills --skill scaffold-exercises
  ```

## Tooling & Setup

- **settings** (`setup.sh`) — `~/.claude/settings.json` is generated, not edited. `setup.sh` merges the tracked `settings.shared.json` (permissions, hooks, statusLine, plugins, `autoMode`) with the untracked `~/.claude/settings.user.json` (`model`, `effortLevel`, `env`), so every launcher, including `claude -p` from scripts, gets both. Change either source and re-run `./setup.sh`; `setup.sh` also installs a git `post-merge` hook (`githooks/post-merge`) that re-runs it after every pull or merge on `main` in this repo. Edits Claude Code writes straight into `settings.json` (`/model`, `/permissions`) are dropped on the next run; `setup.sh` prints them and keeps the old file as `settings.json.bak`, so move any keepers into a source file. The merge is `jq '.[0] * .[1]'`, so a key in both files takes the user file's value and arrays are replaced, not combined.
- **clean-comments** — Set up a `/clean-comments` command and a PreToolUse hook that blocks git commits when staged code files contain self-documenting comments.

  ```
  npx skills add O-Marsters-1997/skills --skill clean-comments
  ```

- **reflect** — Run `/reflect` to review the current session from Claude Code's own transcripts: one Sonnet reviewer per agent that struggled (main, subagents, nested, forked and worktree agents, linked by exact ids) looks for hallucinations, repeated failures, churn, corrections and wasted calls, and the fixes you approve become GitHub issues. Nothing to switch on per repo. Machine setup: `./setup.sh --reflect` (needs Go) builds `~/.claude/bin/reflect`; a global SessionStart hook records instruction-file hashes for `/reflect metrics`. Correction detection is adapted from [claude-reflect](https://github.com/BayramAnnakov/claude-reflect) (MIT).

  ```
  npx skills add O-Marsters-1997/my-claude-code --skill reflect -g -y
  ```

- **subagent-creator** — Design the leanest Claude Code subagent that does the job. It grills you on whether the subagent earns its place and what its scope is, writes it from a template, and checks it against a short conformance list: an explicit tools allowlist, a `maxTurns` budget, worktree isolation for writers, an output contract and a `SCOPE_REQUEST` escape hatch. `expand <name>` widens an agent one step at a time, only when you ask and with evidence from a blocked run. `audit` runs the same check over existing agents.

  ```
  npx skills add O-Marsters-1997/my-claude-code --skill subagent-creator -g -y
  ```

- **require-go-skills** (hook, wired by `settings.shared.json`) — A global PreToolUse hook that blocks the first edit to a Go file until `go-idiomatic` has been loaded in the session, and also `testing-policy` for a `_test.go` in a project that ships that skill. It skips generated files, `vendor/` and paths with no `go.mod`, and fails open if it can't read the transcript. Skip it with `CC_GO_SKILLS_OFF=1`. Tests: `hooks/require-go-skills.test.sh`.
- **require-codegraph** (hook, wired by `settings.shared.json`) — A global PreToolUse hook on Grep, Glob and Bash. In a repo with a `.codegraph/` index it blocks code searches (Grep, Glob, `grep`/`rg`/`find` in Bash) until that agent has queried CodeGraph once, through the MCP tool or `codegraph explore`; after that searching is free. Searches of docs, templates, CSS and config pass, piped greps pass, Read is never gated, and each subagent needs its own query. Fails open if it cannot read the transcript. Skip it with `CC_CODEGRAPH_GATE_OFF=1`. Tests: `hooks/require-codegraph.test.sh`.
- **block-bash-inplace-edits** (hook, wired by `settings.shared.json`) — A global PreToolUse hook on Bash. It denies in-place file writes through the shell (`sed -i`, `perl -pi`, and inline `python`/`node`/`ruby` scripts that open a file for writing; it cannot see inside a script file) and points at Edit/Write, so the comment and skill gates on those tools are not bypassed. Inline interpreter scripts mentioning `/tmp` or the scratchpad, `gofmt -w` and other tool-owned rewrites pass. Skip it with `CC_INPLACE_GATE_OFF=1`. Tests: `hooks/block-bash-inplace-edits.test.sh`.
- **Explore** (agent) — Overrides the built-in Explore subagent, which skips `CLAUDE.md`, so it queries CodeGraph first in indexed repos. It leaves `tools` unset so it inherits MCP tools.

## Writing & Knowledge

- **unslop** — Strip AI tells from prose and put voice back in: puffery, AI vocabulary, em dashes, inline-header lists, hedging, passive voice. Vendored from [cursor/plugins](https://github.com/cursor/plugins/blob/main/pstack/skills/unslop/SKILL.md); the description is rewritten as a trigger clause so it loads before prose work instead of waiting to be invoked.

- **edit-article** — Edit and improve articles by restructuring sections, improving clarity, and tightening prose.

  ```
  npx skills add O-Marsters-1997/skills --skill edit-article
  ```

- **ubiquitous-language** — Extract a DDD-style ubiquitous language glossary from the current conversation.

  ```
  npx skills add O-Marsters-1997/skills --skill ubiquitous-language
  ```

