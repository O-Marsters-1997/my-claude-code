# Implementation review gate

Load the `code-review` skill and run its independent Standards and Contract axes at medium level,
or low for a small fix. Spawn its reviewers from the implementation agent, not from a built-in
research child that cannot delegate. Pass the worktree path and fixed `Base:`; before commit,
review tracked staged and unstaged changes against the base plus every untracked file. After a
commit, compare `git -C <worktree> diff origin/<base>...HEAD`. If the effective diff is empty,
stop and report; never fall back to `HEAD~1`.

Run the review in the foreground and read its result when it returns. Apply findings and rerun
review when a fix materially changes the diff. An unavailable reviewer is a blocker, not a passed
review. Never wait by polling or sleeping (`sleep`, `timeout`, `tail -f`, `perl -e 'sleep …'`).
