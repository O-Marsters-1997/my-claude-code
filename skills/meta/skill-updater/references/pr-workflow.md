# PR Workflow

Commands and conventions for opening a draft PR after a skill update. All git commands use `-C <source-repo>` so they work regardless of cwd.

## Branch creation

```bash
git -C <source-repo> fetch origin
git -C <source-repo> checkout -b skill-update/<name>-<YYYYMMDD> origin/main
```

If the branch name already exists (collision), append `-2`, `-3`, etc.

## Staging and commit

Stage only the files that were actually changed:

```bash
git -C <source-repo> add <skill-dir>/SKILL.md <skill-dir>/evals/evals.json
# if references/ files were added/changed:
git -C <source-repo> add <skill-dir>/references/
```

Commit using conventional commits:

```
chore(skills/<name>): <one-line summary of the change>
```

## PR template

```markdown
## Summary
<one sentence: what changed in the skill>

## Motivation
<why: incorrect guidance / user feedback / new pattern / gap uncovered in practice>

## Intended impact
<how Claude will behave differently when the skill fires>

## Evidence
<the scenario, then the old skill's result and the new skill's result on it, side by side, and the rerun count if the edit was reworked>
```

Four sections, one short paragraph each (Evidence may be a small table). No boilerplate footer or checklists.

## Opening the PR

```bash
gh pr create \
  --draft \
  --base main \
  --title "chore(skills/<name>): <summary>" \
  --body "$(cat <<'EOF'
## Summary
...

## Motivation
...

## Intended impact
...

## Evidence
...
EOF
)"
```

Run with `cwd` set to `<source-repo>`, or pass `-R <owner>/<repo>` if needed.

## Failure modes

Stop and surface a clear error — do not auto-recover — if:

- **Dirty working tree**: `git status` shows uncommitted changes in `<source-repo>`. Ask the user to resolve before proceeding.
- **No gh auth**: `gh auth status` fails. Prompt the user to run `gh auth login`.
- **Push rejected**: report the exact error; never force-push.
- **Branch already exists on remote**: append a suffix (`-2`, `-3`) and retry once; if still blocked, ask the user.

Never use `--no-verify` and never amend a commit that has already been pushed.
