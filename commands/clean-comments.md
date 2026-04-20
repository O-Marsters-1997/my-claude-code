Remove self-documenting comments from the files provided as arguments, or from all staged files if no arguments are given.

A self-documenting comment is one that restates what the code already clearly shows — it adds no information beyond what a reader gets from the code itself.

**Remove** comments that:
- Describe what a function, variable, or type does when the name makes it obvious (`// increment counter` above `count++`)
- Restate the type or return value in prose (`// returns a string`)
- Echo a straightforward operation (`// open the file` above `os.Open(...)`)
- Label sections of code that are already clearly delimited (`// loop over items`)

**Keep** comments that:
- Explain *why* a decision was made, not *what* the code does
- Document non-obvious behaviour, edge cases, or gotchas
- Record constraints imposed by external systems, specs, or bugs
- Provide context a reader cannot derive from the code alone (e.g. `// workaround for upstream issue #123`)
- Are godoc / JSDoc / rustdoc public API documentation

**Process:**
1. If arguments are provided, treat them as file paths to clean
2. If no arguments, get staged files: `git diff --cached --name-only --diff-filter=ACM`
3. Filter to code files (`.go`, `.ts`, `.tsx`, `.js`, `.jsx`, `.py`, `.rs`)
4. For each file, read it, identify self-documenting comments, remove them, and write the file back
5. Re-stage any files that were modified: `git add <file>`
6. Report what was removed and from which files; if nothing needed removing, say so
