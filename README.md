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
[`skills/product/workflow/README.md`](skills/product/workflow/README.md) for the map. Install the
whole suite — the skills call each other, so a partial install leaves broken handoffs:

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
| `to-roadmap` | portfolio | `./docs/approach.md` | `./roadmap.html` (cards are features) |
| `to-prd` | feature | one feature | `./docs/prd-<feature>.md` + a `[PRD]` issue |
| `to-plan` | feature | a PRD | `./plans/<feature>.md` |
| `to-tickets` | feature | a plan | GitHub issues |
| `ticket-tracker` | feature | GitHub issues | `status:*` label moves |

The suite calls three routines that are not stages — install them too:

```
for s in grill-with-docs source-synthesis design-an-interface; do
  npx skills add O-Marsters-1997/my-claude-code --skill "$s" -g -y
done
```

## Planning & Design

These skills help you think through problems before writing code.

- **grill-with-docs** — Get relentlessly interviewed about a plan or design until every branch of the decision tree is resolved.

  ```
  npx skills add O-Marsters-1997/skills --skill grill-with-docs
  ```

- **design-an-interface** — Generate multiple radically different interface designs for a module using parallel sub-agents.

  ```
  npx skills add O-Marsters-1997/skills --skill design-an-interface
  ```

- **request-refactor-plan** — Create a detailed refactor plan with tiny commits via user interview, then file it as a GitHub issue.

  ```
  npx skills add O-Marsters-1997/skills --skill request-refactor-plan
  ```

## Development

These skills help you write, refactor, and fix code.

- **tdd** — Test-driven development with a red-green-refactor loop. Builds features or fixes bugs one vertical slice at a time.

  ```
  npx skills add O-Marsters-1997/skills --skill tdd
  ```

- **triage-issue** — Investigate a bug by exploring the codebase, identify the root cause, and file a GitHub issue with a TDD-based fix plan.

  ```
  npx skills add O-Marsters-1997/skills --skill triage-issue
  ```

- **improve-codebase-architecture** — Explore a codebase for architectural improvement opportunities, focusing on deepening shallow modules and improving testability.

  ```
  npx skills add O-Marsters-1997/skills --skill improve-codebase-architecture
  ```

- **migrate-to-shoehorn** — Migrate test files from `as` type assertions to @total-typescript/shoehorn.

  ```
  npx skills add O-Marsters-1997/skills --skill migrate-to-shoehorn
  ```

- **scaffold-exercises** — Create exercise directory structures with sections, problems, solutions, and explainers.

  ```
  npx skills add O-Marsters-1997/skills --skill scaffold-exercises
  ```

## Tooling & Setup

- **setup-pre-commit** — Set up Husky pre-commit hooks with lint-staged, Prettier, type checking, and tests.

  ```
  npx skills add O-Marsters-1997/skills --skill setup-pre-commit
  ```

- **git-guardrails-claude-code** — Set up Claude Code hooks to block dangerous git commands (push, reset --hard, clean, etc.) before they execute.

  ```
  npx skills add O-Marsters-1997/skills --skill git-guardrails-claude-code
  ```

- **clean-comments** — Set up a `/clean-comments` command and a PreToolUse hook that blocks git commits when staged code files contain self-documenting comments.

  ```
  npx skills add O-Marsters-1997/skills --skill clean-comments
  ```

## Writing & Knowledge

- **write-a-skill** — Create new skills with proper structure, progressive disclosure, and bundled resources.

  ```
  npx skills add O-Marsters-1997/skills --skill write-a-skill
  ```

- **edit-article** — Edit and improve articles by restructuring sections, improving clarity, and tightening prose.

  ```
  npx skills add O-Marsters-1997/skills --skill edit-article
  ```

- **ubiquitous-language** — Extract a DDD-style ubiquitous language glossary from the current conversation.

  ```
  npx skills add O-Marsters-1997/skills --skill ubiquitous-language
  ```

- **obsidian-vault** — Search, create, and manage notes in an Obsidian vault with wikilinks and index notes.

  ```
  npx skills add O-Marsters-1997/skills --skill obsidian-vault
  ```
