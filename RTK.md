# RTK - Rust Token Killer

**Usage**: Token-optimized CLI proxy (60-90% savings on dev operations)

## Meta Commands (always use rtk directly)

```bash
rtk gain              # Show token savings analytics
rtk gain --history    # Show command usage history with savings
rtk discover          # Analyze Claude Code history for missed opportunities
rtk proxy <cmd>       # Execute raw command without filtering (for debugging)
```

## Git Push — Always Use rtk proxy

`rtk git push` checks GitHub CI status before pushing and blocks on failure or when
a branch has never been pushed (status "none"). This creates a bootstrap deadlock on
new branches and prevents force-pushing through a failing CI.

**Rule: always push via `rtk proxy`:**

```bash
rtk proxy "git push -u origin <branch>"   # initial push / new branch
rtk proxy "git push"                       # subsequent pushes
```

This applies whenever raising a PR or pushing commits. The `rtk proxy` wrapper still
tracks usage but skips the CI gate entirely.

## Installation Verification

```bash
rtk --version         # Should show: rtk X.Y.Z
rtk gain              # Should work (not "command not found")
which rtk             # Verify correct binary
```

⚠️ **Name collision**: If `rtk gain` fails, you may have reachingforthejack/rtk (Rust Type Kit) installed instead.

## Hook-Based Usage

All other commands are automatically rewritten by the Claude Code hook.
Example: `git status` → `rtk git status` (transparent, 0 tokens overhead)

Refer to CLAUDE.md for full command reference.
