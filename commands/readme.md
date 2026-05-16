Generate or refresh a README.md for the project at the path in $ARGUMENTS, or the current working directory if no path is given.

Follow the `readme-generator` skill. The README should answer three questions a stranger would have in under 30 seconds: *what is this, how do I install it, how do I use it.*

**Required sections:**
- Title + italic one-line tagline (the "why")
- 2–4 sentence "what it does" — problem, audience, what's different
- Installation — actual commands in fenced code blocks with language hints
- Usage — minimal, runnable examples (as few as possible while showing the scope of the tool — often one, sometimes two)

**Optional sections** (only if the project genuinely has something to say): Prerequisites, Configuration, Development, Testing, API/Reference, Contributing, Future development, Limitations, License.

**Process:**
1. Discover project type (`package.json`, `pyproject.toml`, `Cargo.toml`, `go.mod`, etc.) to determine install/usage shape.
2. Find the "why" from `description` fields, an existing README's opening paragraph, or the main entry point docstring. **If none of those give a real description (or it's a generic stub), ask the user for one sentence.** Don't invent one.
3. **Ask the user one short open question:** "Is there anything specific about this project a reader needs to know that the code won't tell me? E.g. a non-obvious gotcha, why it exists rather than \<obvious alternative\>, who it's not for, a known limitation." Incorporate the answer using the mapping table in the skill (contrast sentences in "what it does", callouts under Usage, dedicated Limitations section, etc.) — don't dump it into a generic "notes" section.
4. If a README already exists, ask whether to *refresh*, *rewrite*, or *augment* — and back it up to `README.md.bak` before writing.
5. Show the draft before writing. Adjust based on feedback.
6. Write to `README.md`.

GitHub-flavoured markdown. Sentence case headings. No emojis, no badge soup, no roadmap/acknowledgments unless the user asks. Code blocks always get a language hint.
