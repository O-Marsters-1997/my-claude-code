# Priority Catalogue

Use this catalogue when eliciting priorities from the user. Present a curated short-list
(8–10 options) relevant to the use case, and ask them to pick 3–5 and rank them.

Each priority below includes:
- **What it means in practice** — so you can explain it to the user in plain terms
- **When it's most important** — context that helps the user decide quickly

---

## Community Support

**What it means**: Size and health of the user community — GitHub stars, active forums,
Stack Overflow coverage, Discord/Slack, conference talks. Larger communities mean more
tutorials, more answered questions, and faster help when something goes wrong.

**Most important when**: The team is new to this domain, or the tool will be maintained
long-term by people who didn't pick it.

---

## Maintenance Activity

**What it means**: How actively the project is developed — recent releases, open issue
response time, number of core maintainers, bus factor. A tool with a single maintainer
who last released two years ago carries real risk.

**Most important when**: The tool is a dependency that will need to track upstream changes
(security patches, language version compatibility, platform support).

---

## Avoiding Vendor Lock-in

**What it means**: Whether switching away is feasible — open data formats, standard protocols,
no proprietary SDKs required for core functionality. Low lock-in means you can migrate if
pricing changes, the vendor shuts down, or requirements shift.

**Most important when**: The tool handles data persistence, auth, or any function that
would be expensive to rewrite if the vendor changes terms.

---

## Extensibility / Composability

**What it means**: How easy it is to add behaviour — plugin systems, middleware hooks,
composable APIs, escape hatches for custom logic. An extensible tool bends to your
requirements; a rigid one forces you to work around it.

**Most important when**: Your use case is non-standard or will evolve, and you expect to
need custom integrations or override default behaviour.

---

## Ecosystem Fit

**What it means**: How well the tool integrates with the existing codebase — official SDKs
in the project's language, idiomatic patterns, compatible with existing frameworks and
infrastructure. Poor ecosystem fit means glue code and translation layers.

**Most important when**: The project already has strong conventions (e.g. a specific ORM,
a testing framework, a deployment platform) and deviation adds friction.

---

## License

**What it means**: The legal terms under which the software can be used — MIT/Apache (very
permissive), GPL (copyleft — anything linking to it may need to be open-source), LGPL
(copyleft but more permissive for libraries), BSL/commercial (source-available but with
restrictions). Some licenses are incompatible with closed-source commercial products.

**Most important when**: The project is commercial, will be distributed, or has legal/
compliance requirements.

---

## Performance

**What it means**: Throughput, latency, resource usage (CPU, memory, network). Whether the
tool can handle the expected load without becoming the bottleneck.

**Most important when**: The use case is on a hot path (request handling, real-time
processing) or resource-constrained (edge, embedded, high-scale).

---

## Learning Curve

**What it means**: Time to become productive — quality of documentation, availability of
tutorials, simplicity of the API, familiarity of the mental model. A steep learning curve
is fine for a single expert; it's a tax on a whole team.

**Most important when**: Multiple people will work with the tool, onboarding speed matters,
or you want to minimise the risk that one person becomes the sole expert.

---

## Cost

**What it means**: Total cost of adoption — licensing fees, managed service pricing, compute
costs, egress fees. Includes the hidden cost of operational work if self-hosted.

**Most important when**: Budget is constrained, the tool will handle significant volume
(where per-request pricing compounds), or there's a free/open-source alternative that
covers the requirements.

---

## Operational Burden

**What it means**: The ongoing work required to run the tool in production — upgrades,
backups, failover, monitoring, tuning. Managed services reduce operational burden at the
cost of control (and usually money); self-hosted gives control at the cost of ops work.

**Most important when**: The team is small, has no dedicated ops/SRE, or is optimising for
shipping speed over operational control.

---

## Maturity / Production-Proven

**What it means**: Whether the tool is used in production at scale by well-known organisations.
A 1.0 release does not equal production-proven. Battle-tested tools have known failure modes;
new tools have unknown ones.

**Most important when**: The use case is critical-path, the failure mode is costly, or the
team lacks the capacity to debug novel edge cases.
