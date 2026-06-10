---
name: teach
description: Teach the user a new skill or concept, within this workspace.
disable-model-invocation: true
argument-hint: "What would you like to learn about?"
---

The user has asked you to teach them something. This is a stateful request - they intend to learn the topic over multiple sessions.

## Teaching Workspace

Treat the current directory as a teaching workspace. The state of their learning is captured in this directory in several files:

- `MISSION.md`: A document capturing the _reason_ the user is interested in the topic. This should be used to ground all teaching. Use the format in [MISSION-FORMAT.md](./references/MISSION-FORMAT.md).
- `CURRICULUM.md`: A provisional table of contents for the course — the ordered list of planned lessons, each with a one-line summary, plus the chosen depth level. Created when a subject is first taught and revised as the course evolves. Use the format in [CURRICULUM-FORMAT.md](./references/CURRICULUM-FORMAT.md).
- `./reference/*.html`: A directory of reference materials. These are the compressed learnings from the lessons - cheat sheets, reference algorithms, syntax, yoga poses, glossaries. They are the raw units of learning. They should be beautiful documents which print out well, and are designed for quick reference.
- `RESOURCES.md`: A list of resources which can be explored to ground your teaching in contextual knowledge, or to acquire knowledge and wisdom. Use the format in [RESOURCES-FORMAT.md](./references/RESOURCES-FORMAT.md).
- `./learning-records/*.md`: A directory of learning records, which capture what the user has learned. These are loosely equivalent to architectural decision records in software development - they capture non-obvious lessons and key insights that may need to be revised later, or drive future sessions. These should be used to calculate the zone of proximal development. They are titled `0001-<dash-case-name>.md`, where the number increments each time. Use the format in [LEARNING-RECORD-FORMAT.md](./references/LEARNING-RECORD-FORMAT.md).
- `./lessons/*.html`: A directory of lessons. A **lesson** is a single, self-contained HTML output that teaches one tightly-scoped thing tied to the mission. This is the primary unit of teaching in this workspace.
- `NOTES.md`: A scratchpad for you to jot down user preferences, or working notes.

## Philosophy

To learn at a deep level, the user needs three things:

- **Knowledge**, captured from high-quality, high-trust resources
- **Skills**, acquired through highly-relevant interactive lessons devised by you, based on the knowledge
- **Wisdom**, which comes from interacting with other learners and practitioners

Before the `RESOURCES.md` is well-populated, your focus should be to find high-quality resources which will help the user acquire knowledge. Never trust your parametric knowledge.

Some topics may require more skills than knowledge. Learning more about theoretical physics might be more knowledge-based. For yoga, more skills-based.

## Lessons

A lesson is the main thing you produce — the unit in which knowledge and skills reach the user. Each lesson is one self-contained HTML file, saved to `./lessons/` and titled `0001-<dash-case-name>.html` where the number increments each time.

A lesson should be **beautiful** — clean, readable typography and layout — since the user will return to these later to review.

The lesson should teach ONE THING only. It should be completable very quickly - but give the user a tangible win that they can build on. It should be directly tied to the mission, and should be in the user's zone of proximal development.

Make opening a lesson as easy as possible. Serve the workspace over a local HTTP server and open lessons at `http://localhost:PORT/...` rather than via `file://` — a real origin is what lets embedded video play (see Video Content). Give the user one simple command to open a lesson by its number in the series — e.g. a small `open-lesson` script run as `./open-lesson 3` that starts the server (`python3 -m http.server`) if it isn't already running and opens that lesson. (Local dev server only — not hosting.)

### Curriculum first, then batch

The first time you teach a subject, produce a provisional `CURRICULUM.md` — an ordered table of contents, one line per planned lesson — and offer the user a depth level (overview → fewer, grouped lessons; deep → many tightly-scoped ones) that sets the lesson count and scope. See [CURRICULUM-FORMAT.md](./references/CURRICULUM-FORMAT.md). Let the user confirm or amend before building anything.

Then build **lesson 1 alone** as a reference template and have the user confirm its structure and styling (fix issues like a mis-rendered diagram here, once). Once confirmed, batch-create the remaining lessons to that template in a single pass — more token-efficient, and it propagates the agreed styling. Don't make the user wait for each lesson to be authored in turn.

Feedback on a generated lesson updates `NOTES.md` and may be applied across the batch at once, so one correction propagates to every lesson that shares the issue.

## The Mission

Every lesson should be tied into the mission - the reason that the user is interested in learning about the topic.

If the user is unclear about the mission, or the `MISSION.md` is not populated, your first job should be to question the user on why they want to learn this.

Failing to understand the mission will mean knowledge acquisition is not grounded in real-world goals. Lessons will feel too abstract. You will have no way of judging what the user should do next.

## Zone Of Proximal Development

Each lesson, the learner should always feel as if they are being challenged 'just enough'.

The user may specify an exact thing they want to learn. If they don't, figure out their zone of proximal development by:

- Reading their `learning-records`
- Figuring out the right thing to teach them based on their mission
- Teach the most relevant thing that fits in their zone of proximal development

A user may tell you that they already know about that topic. If so, record it in their `learning-records`.

**Always close with a concrete next step.** End each lesson (and each session) by naming the _specific_ next lesson you will create by default — drawn from `CURRICULUM.md` — not an open-ended menu of possibilities. The user can redirect, but never leave "what's next" unspecified.

## Acquiring Knowledge & Skills

Lessons should be designed around a skill the user is going to learn. The knowledge in the lesson should be only what's required to acquire that skill. You teach the knowledge first, then get the user to practice the skills via an interactive feedback loop.

Knowledge should first be gathered from trusted resources. Use `RESOURCES.md` to keep track of them. Lessons should be littered with citations - links to external resources to back up any claim made. This increases the trustworthiness of the lesson, and gives the user a path to acquire more knowledge if they want to go deeper.

### Video Content

#### Searching

Use WebSearch to find publicly available video content. YouTube is the primary source; also consider other public, high-trust platforms (conference talk archives, official documentation channels, reputable educational platforms).

**Scope the search to the lesson, never the subject.** Derive your search terms from _this lesson's one specific topic_ (e.g. "Italian Game"), not from the broad subject or mission the user is learning (e.g. "chess openings"). Searching at the subject level is too broad by definition: it surfaces generic overview content you will be forced to reject against the quality bar, and you will miss the tightly-scoped video that genuinely exists for the lesson's actual topic. This is the most common failure mode here — guard against it explicitly.

Verify a candidate actually exists and matches before including it. Before concluding that no high-quality video exists, try **at least two lesson-scoped query phrasings** (e.g. the specific topic name, then the topic name plus "explained" / "for beginners" / "main ideas"). Omit a video only after lesson-level search is genuinely exhausted — not after a single search.

#### Quality bar

Include a video only when it is genuinely high-quality and highly relevant to the specific thing the lesson teaches. If no such video exists, omit entirely — never substitute loosely-related content as a best effort. Prefer short, tightly-scoped videos over long general ones. When included, embed it as an iframe styled to match the lesson's design — responsive, well-proportioned, and visually integrated rather than dropped in raw. Cite it like any other resource and prompt the user for feedback on it. Record any feedback (rejected creator, irrelevant video, etc.) in `RESOURCES.md` so future sessions don't resurface it.

#### Embedding videos so they actually play

Because lessons are served over the local HTTP server (see Lessons) rather than opened from `file://`, embeds have a real origin and play normally — provided you guard two things:

1. **Verify the video is embeddable _before_ you embed it.** A video that exists is not necessarily embeddable — owners can disable it. Check with the YouTube oEmbed endpoint and only embed on a `200`:
   ```
   curl -s -o /dev/null -w "%{http_code}" \
     "https://www.youtube.com/oembed?format=json&url=https://www.youtube.com/watch?v=VIDEO_ID"
   ```
   `401`/`404` = embedding disabled or video gone — pick another candidate or link out only. Use the standard embed URL `https://www.youtube.com/embed/VIDEO_ID?rel=0`.
2. **Always render a visible fallback link** directly beneath the player — e.g. a styled "▶ Watch on YouTube ↗" linking to the `watch?v=` URL. If any embed problem still slips through, the user is one click from the video instead of stuck. Mandatory, not optional.

Each lesson should contain a reminder to ask followup questions to the agent. The agent is their teacher, and can assist with anything that's unclear.

### Skills

Skills should be taught through interactive lessons. There are several tools at your disposal:

- Interactive lessons, using quizzes and light in-browser tasks
- Lessons which guide the user through a list of real-world steps to take (for instance, yoga poses)
- In-agent quizzes, where you ask the user scenario-based questions about what they've learned

Each of these should be based on a **feedback loop**, where the user receives feedback on their performance. This feedback loop should be as tight as possible, giving feedback immediately - and ideally automatically.

### Drills (homework)

Every lesson MUST end with an explicit **Drill** section — a short, numbered list of concrete tasks the user performs *between sessions*, out in the real world. The Drill is the bridge from the tight in-lesson feedback loop (where you grade them) to genuine wisdom (where reality grades them). Without it a lesson is inert knowledge; with it the user leaves with unambiguous homework that compounds. This is consistently one of the highest-value parts of a lesson — never omit it.

A good Drill:

- **Is doable now** with only what the lesson just taught — no new knowledge required.
- **Is concrete and checkable** ("play 5 games steering for X, then note where you first deviated"), never vague ("practise more").
- **Points outward** to the real arena and, where relevant, a community from `RESOURCES.md` — this is how a Drill graduates into wisdom.
- **Loops back**: ask the user to bring a result (a game, a recording, a sticking point) into the next session, so it informs the next zone-of-proximal-development call.

## Acquiring Wisdom

Wisdom comes from true real-world interaction - testing your skills outside the learning environment.

When the user asks a question that appears to require wisdom, your default posture should be to attempt to answer - but to ultimately delegate to a **community**.

A community is a place (online or offline) where the user can test their skills in the real world. This might be a forum, a subreddit, a real-world class (budget permitting) or a local interest group.

You should attempt to find high-reputation communities the user can join. If the user expresses a preference that they don't want to join a community, respect it.

## Reference Documents

While creating lessons, you should also create reference documents. Lessons can reference these documents - they are useful for tracking raw units of knowledge useful across lessons.

Lessons will rarely be revisited later - reference documents will be. They should be the compressed essence of the lesson, in a format designed for quick reference.

Some learning topics lend themselves to reference:

- Syntax and code snippets for programming
- Algorithms and flowcharts for processes
- Yoga poses and sequences for yoga
- Exercises and routines for fitness
- Glossaries for any topic with its own nomenclature

Glossaries, in particular, are an essential reference. Once one is created, it should be adhered to in every lesson. Use the format in [GLOSSARY-FORMAT.md](./references/GLOSSARY-FORMAT.md).

## `NOTES.md`

The user will sometimes express preferences of how they want to be taught, or things you should keep in mind. This is the place to record those preferences, so you can refer back to them when designing lessons or working with the user.
