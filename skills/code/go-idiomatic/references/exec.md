# Go Subprocesses (`os/exec`)

For shelling out to an external CLI (`git`, `gh`, a linter, an agent binary). The rules hold for
a single inline call too; the wrapper package only appears once there are several callers.

---

## Argv, Never a Shell

```go
cmd := exec.CommandContext(ctx, "git", "-C", dir, "log", "--format=%H", ref)
```

Not `exec.Command("sh", "-c", "git log "+ref)`: a `ref` containing `; rm -rf ~` becomes code,
quoting is now your problem, and the error names `sh` instead of `git`. Use a shell only when
the user configures a shell command line, and name that in the type (`ShellCommand string`).

Never pass secrets as arguments. Every user on the machine can read argv through `ps`. Pass them
in `cmd.Env` (appended to `os.Environ()`) or on stdin.

## Context, a Named Timeout, `WaitDelay`

```go
const ghTimeout = 30 * time.Second

ctx, cancel := context.WithTimeout(ctx, ghTimeout)
defer cancel()
cmd := exec.CommandContext(ctx, "gh", args...)
cmd.WaitDelay = 5 * time.Second
```

The caller's `ctx` usually lives as long as the program, so without a timeout a `gh` stalled on
the network blocks its caller forever. `CommandContext` kills the process on cancel, but `Wait`
still blocks while a grandchild holds the stdout or stderr pipe open; `WaitDelay` bounds that.
For a graceful stop, set `cmd.Cancel` to send `os.Interrupt`; `WaitDelay` then escalates to kill.

A CLI waiting on a prompt hangs until the timeout. Leave `Stdin` nil and switch prompts off with
the tool's own variable (`GIT_TERMINAL_PROMPT=0`, `GH_PROMPT_DISABLED=1`).

A process meant to outlive the call (a daemon, a detached agent) uses `exec.Command` and
`Start`, not `CommandContext`: cancelling the context would kill it. Its owner decides who waits
on it and who signals it.

## Stderr Goes in the Error

```go
var stderr bytes.Buffer
cmd.Stderr = &stderr
out, err := cmd.Output()
if err != nil {
	return nil, fmt.Errorf("%s: %w: %s", cmd, err, bytes.TrimSpace(stderr.Bytes()))
}
```

Without this the error reads `exit status 1` and the CLI's explanation is lost. Don't use
`CombinedOutput` when you parse stdout: one warning on stderr corrupts the JSON. Truncate the
command line in the message when callers pass large arguments such as issue bodies.

## Exit Codes Are Classified Once

Some CLIs answer through their exit status: `git merge-base --is-ancestor`, `git diff --quiet`,
`grep`. Map the codes in one function beside the command:

```go
func IsAncestor(ctx context.Context, dir, a, b string) (bool, error) {
	err := command.Run(ctx, dir, "git", "merge-base", "--is-ancestor", a, b)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return false, nil
	}
	return err == nil, err
}
```

Never `if err != nil { return false }` at the call site: that turns exit 128 ("not a repo",
"bad ref") into "no". Give any other meaningful code a sentinel (`references/errors.md`). After
a timeout the error reads `signal: killed`, so check `ctx.Err()` first and return that instead.

## One Wrapper Package

Once a third package shells out, extract `internal/command` with `Run` and `Output`. They build
the command, set `Dir`, capture stderr and wrap the error as above. Each CLI then gets a thin
package on top (`internal/git`, `internal/gh`) that owns its timeout and exit-code mapping. Any
other `exec.Command` in the codebase is a review flag, except the long-lived process case above.

## Vendor CLI JSON

```go
const prFields = "number,title,headRefName,state"

type rawPR struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	HeadRefName string `json:"headRefName"`
	State       string `json:"state"`
}
```

Keep the `--json` field list as a const directly above the struct it fills, so adding a field is
one hunk, not two edits in different files. Decode with plain `json.Unmarshal`: unknown fields
are ignored, which is what you want from a CLI that adds fields between releases.

`--limit N` truncates silently. Paginate, or treat a full page as a signal: when
`len(prs) == prLimit`, return an error or log at Warn. Never read a full page as "all of them".

## Testing

Test callers against a consumer-owned interface (`references/testing.md`). Test the CLI package
against the real binary when it is cheap: `git init` in `t.TempDir()`. When it isn't, put a fake
on `PATH` by re-executing the test binary:

```go
func TestMain(m *testing.M) {
	if os.Getenv("FAKE_GH") == "1" {
		os.Exit(fakeGH(os.Args[1:], os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

func installFakeGH(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	if err := os.Symlink(os.Args[0], filepath.Join(dir, "gh")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_GH", "1")
}
```

`fakeGH(args, stdout, stderr) int` serves fixtures from `testdata/` recorded from the real CLI,
so the field names are the vendor's, not your guess at them. For an end-to-end suite, build the
same function as its own `main` package instead.
