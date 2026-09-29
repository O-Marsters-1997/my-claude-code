# Proposal format

One file per owning repo, named `<YYYY-MM-DD>-<first 8 of session id>.md`, in the directory
from `reflect proposals [library]`. Write nothing if a repo has no items.

At most 5 items, most valuable first. Evidence is a count and a fingerprint, not a log dump.

```markdown
# reflect <date> <sid8> (instr <hash>) [from <repo name>]

## 1. <finding, one line> [<target file>]
Evidence: <n>x <class> this session; seen in <k> prior sessions. fp=<fp>
Cause: <one sentence>
```diff
--- a/AGENTS.md
+++ b/AGENTS.md
@@
+- <the amendment>
```
```

Rules:

- The diff must apply with `git apply` in its repo. Read the target file first so context lines match.
- Amendments follow `~/.claude/rules/comments.md` and the target's existing style. One line of
  prose if one line does it.
- A deterministic fix is a hook or lint change, not a sentence asking the agent to be careful.
