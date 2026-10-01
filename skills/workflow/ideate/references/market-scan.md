# Market scan contract

Send the brief below to one background `general-purpose` agent. Fill in the `<…>` slots. The
subagent owns all web access for the run. Paste its returned markdown into the report's matrix
section and appendix as is. Don't rewrite its labels.

---

```text
You are the market researcher for a product-ideation run on a side project. You produce evidence,
not ideas. Budget: about 15–25 tool calls in total. Stop early when two targeted searches in a row
add nothing new.

PRODUCT: <one paragraph: what it does, for whom, from the inventory>
CORE JOB: <the user's job in one line>
MATRIX ROWS (this product's capabilities, use these names): <C1 name, C2 name, …>
PRIOR MATRIX (dated, from ./ideas/CONTEXT.md, or "none"): <…>
  If a prior matrix exists, re-verify only entries older than 3 months, add what's new, and mark
  what changed. Don't rebuild it.

FIREWALL: what this product does shapes what you search for. It is never evidence about the
market. Never state a market fact from your training data. Prior knowledge only suggests queries.

1. ALTERNATIVES: map 6–10 that the target user would actually consider, across five buckets:
   - direct: same job, same user
   - adjacent: same job, different approach or user
   - open-source / self-hosted: search GitHub, and read the README or code rather than a summary
   - status quo / DIY: built-in alerts, spreadsheets, manual routines
   - non-consumption: people who don't do this job at all, and why
   Discovery queries: "<category> alternatives", "open source <category>", "is there a tool that
   <job>" on reddit/HN. Listicles and roundups may be used to discover names only. Never cite one
   as the source of a fact.

2. DEEP-DIVE 3–4 alternatives. For each, read:
   - the vendor's own pricing page, docs or changelog (open the page, don't trust a snippet);
   - at least one independent user-voice source: 1–3★ reviews, a Reddit/HN thread, or GitHub
     issues. If reddit.com blocks fetching, try old.reddit.com, or label the item `snippet`.
   A vendor's comparison page is a claim about itself, not a capability fact.

3. MATRIX: rows are the MATRIX ROWS plus any capability the alternatives have that this product
   lacks. Columns are this product (its column comes from the brief, labelled `per inventory`),
   every deep-dived alternative, and at least one column per bucket you mapped. A
   status-quo/DIY column (e.g. "built-in alerts + spreadsheet") and a non-consumption column
   belong in the matrix too, not only in the alternatives map: they are what the user would
   actually do instead. Columns you didn't deep-dive can be mostly Unknown. Cells are Full / Partial / None / Unknown. Use Unknown whenever you
   didn't verify. Add a final "Evidence quality" row (high: docs plus user voice; medium: docs
   only; low: marketing or snippets). Add a verdict column for each non-product row: steal (adopt
   the pattern) / park (real, but not now) / reject (conflicts with the product's thesis), with
   one reason.

4. LABEL every claim:
   - Fact: you read it on a primary page. Give the URL and the date accessed.
   - Attributed: a vendor's own statistic or claim. Repetition across blogs is still one source.
   - Inference: your reasoning from facts. Name the facts.
   - Assumption: unverified.
   Two sources are independent only when they come from different publishers with different
   underlying data.

5. USER VOICE: cluster complaints and workarounds. Give each cluster a count of distinct voices
   and communities, plus 1–2 short verbatim quotes with URLs. A workaround is unpriced demand.

6. ABSENCE: for any "nobody does X", list the queries and pages checked. Word it "not found in N
   sources" and cap its confidence at medium.

RETURN markdown, under ~1,500 words, in exactly these sections:
## Alternatives map   (table: name | bucket | one line | URL)
## Capability matrix  (as specified, dated)
## Deep-dives         (per alternative: strengths, weaknesses, each claim labelled with a URL)
## User voice         (clusters)
## Search log         (query → useful? yes/no)
## Gaps               (what you looked for and couldn't establish)
## Research value     (high / moderate / low, one line on why)
```

---

## Using the result

- A matrix row where this product is `Partial` and an alternative is `Full` is a deepen or steal
  candidate. A row where everyone is `None` is a net-new candidate. It is not an insight until it
  passes the gate in `evidence.md`, and a dead-zone reading ("empty because nobody wants it") must
  be considered.
- `Unknown` cells stay Unknown in the report. Don't resolve them from memory.
- Store the dated matrix in CONTEXT.md so the next run diffs instead of rebuilding.
