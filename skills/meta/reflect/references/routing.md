# Routing an amendment

| The fix is | Target |
| --- | --- |
| A navigation pointer, or a convention nearly every task needs | `AGENTS.md` |
| A coding standard that takes judgement | the review skill's standards, never `AGENTS.md` |
| Wiring, order of steps, or procedure for one task | the skill that owns the task (`SKILL.md` or its `references/`) |
| Something a script can check or enforce | a hook, or a linter rule, never prose |
| Agent-specific behaviour | the agent's file under `agents/` |
| Cross-project preference | `rules/*.md` |

Standards belong to review, which carries the least context; `AGENTS.md` is loaded into every
agent, so it holds navigation pointers. Prefer the narrowest target. A single skill's problem goes in that skill, not in `AGENTS.md`.

## AGENTS.md

If the repo has no `AGENTS.md`, the issue asks to create it and add `@AGENTS.md` to `CLAUDE.md`
(creating that too if needed). `AGENTS.md` is canonical; `CLAUDE.md` stays a pointer.

## Local or library

Two owners:

- **Local**: files inside the repo being worked on. The issue goes to this repo.
- **Library**: skills, agents, rules and hooks that live in the skills library (`library:` line of
  `reflect status`). The issue goes to the library's repo and names the originating repo.

To classify a file under `~/.claude/` or a repo's `.claude/skills/`: resolve symlinks
(`readlink -f`). If it lands inside the library, it is a library file at that path. If it is
an installed copy (from `npx skills add`), find the same skill name under the library's
`skills/*/<name>/` and target that. If neither matches, treat it as local.
