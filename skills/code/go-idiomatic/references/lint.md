# Lint Reference

A golangci-lint (v2) config that enforces the mechanical rules in `SKILL.md`, so they are checked
rather than remembered. Merge it into the repo's `.golangci.yml`.

```yaml
version: "2"
linters:
  enable:
    - errcheck
    - errorlint
    - forbidigo
    - sloglint
    - thelper
    - tparallel
    - usetesting
  settings:
    errcheck:
      check-blank: true
      exclude-functions:
        - (io.ReadCloser).Close
    forbidigo:
      analyze-types: true
      forbid:
        - pattern: ^log\.(Print|Fatal|Panic)
          msg: inject a *slog.Logger; package log is for main only
        - pattern: ^reflect\.DeepEqual$
          msg: use cmp.Diff from github.com/google/go-cmp
    sloglint:
      kv-only: true
      no-global: all
      context: scope
      static-msg: true
      key-naming-case: snake
  exclusions:
    rules:
      - path: ^cmd/
        text: package log
        linters:
          - forbidigo
      - path-except: _test\.go
        text: cmp\.Diff
        linters:
          - forbidigo
```

| Linter      | Enforces                                                                                       |
| ----------- | ---------------------------------------------------------------------------------------------- |
| `errcheck`  | No unchecked errors, including `_ =` discards. A bare `Close` is allowed only on an `io.ReadCloser`, such as `resp.Body`. |
| `errorlint` | Errors matched with `errors.Is`/`errors.As`, not `==` or type assertions; `%w` when wrapping.  |
| `forbidigo` | No package `log` outside `cmd/`; no `reflect.DeepEqual` in tests.                              |
| `sloglint`  | No global logger; `*Context` variants when a `ctx` is in scope; constant messages; snake_case keys. |
| `thelper`   | Test helpers call `t.Helper()` and take `t` first.                                             |
| `tparallel` | `t.Parallel()` used consistently between a test and its subtests.                              |
| `usetesting`| `t.TempDir`, `t.Setenv`, `t.Context` instead of the `os` equivalents.                          |

`paralleltest` is left out on purpose: it demands `t.Parallel()` everywhere, and parallelism is
opt-in.
