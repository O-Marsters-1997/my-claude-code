# Reading a Jam recording

Feedback can arrive as a `jam.dev/c/<id>` link plus the user's bullet points. The recording is
evidence for the points, never a source of points: a problem you spot that the user didn't
raise goes in the report as an aside, not into P1–Pn.

Read it with the Jam CLI (`jam`, signed in via `jam auth login`). Never run `jam delete`: it
destroys the user's recording with no undo.

## Read once, up front

```bash
jam get jam <id>                                   # type, duration, title, description
jam get transcript <id>                            # the user's narration, timestamped (video, mic on)
jam get comments <id>                              # comments carry playbackRelativeTimestamp
jam get metadata <id>                              # state the app attached with jam.metadata()
jam --json get user-events <id> --limit 500        # clicks and typing, with target text
jam get console <id> --level error,warn
jam --json get network <id> --headers --limit 500  # every request, with response headers
```

Description, transcript and timestamped comments state what the user expected. Prefer them over
anything you read off the screen.

## Map each point to its moment

- The point carries a timestamp (`1:07`): use it, ±5 s.
- It doesn't: find the moment from the clicks' target text and the requests that follow them.
  Say which you chose in the play-back.

At each moment, look at what the user saw:

```bash
jam get frames <id> --at <ms>,<ms> --size large --out "$(mktemp -d)"   # video
jam get screenshots <id> --out "$(mktemp -d)"                          # screenshot Jam
```

Then Read the images. Pull frames only at the moments that matter, not across the whole video.

## Follow a request to the server

Jam logs most requests twice, at start and finish. The same `X-Request-Id` on both means one
request. For a failed or slow request whose response headers carry `X-Request-Id`, find the
server's side in the logs. Where the repo ships logs to Grafana Loki and the Grafana MCP is
connected:

```
query_loki_logs  logql: {service="<api service>", env="prod"} |= "<request id>"
                 start/end: the Jam's createdAt minus its duration, to createdAt
```

These lines are **Prod** facts: errors, retries, timings the browser can't see. With no
`X-Request-Id` header or no log access, the server side is **Unverified**; say so rather than
guessing from timing.

## Labels

- **Jam**: what the recording proves: clicks, requests, console, metadata, and text read from a
  frame.
- **Prod**: what the server logged for a request from the recording.
- With no narration, comment or description, what the user objected to is an **assumption**.
  Put it to them in the play-back.
