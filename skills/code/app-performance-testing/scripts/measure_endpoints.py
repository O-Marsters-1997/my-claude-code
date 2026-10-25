#!/usr/bin/env python3
"""Measure single-user HTTP latency per endpoint, split into DNS, connect, TLS, TTFB and download.

Endpoints file (JSON list):
  [{"name": "companies", "method": "GET", "path": "/companies?sort=relevance",
    "record": "heavy", "headers": {}, "body": null}]

Path placeholders like {job.heavy} or {company.typical} are filled from the seed manifest's
known_ids. "record" ("typical"/"heavy") picks which seeded user's session cookie is sent, so
the same endpoint can be measured for a typical and a power user.

Only GET, HEAD and OPTIONS run unless --allow-writes is passed, and --allow-writes is refused
for any host other than localhost/127.0.0.1/::1.
"""

import argparse
import datetime as dt
import http.client
import json
import os
import re
import socket
import ssl
import statistics
import sys
import time
import urllib.parse

SAFE_METHODS = {"GET", "HEAD", "OPTIONS"}
LOCAL_HOSTS = {"localhost", "127.0.0.1", "::1"}
TIER_FOR_RECORD = {"heavy": "power", "typical": "typical"}


def utc_now():
    return dt.datetime.now(dt.timezone.utc).isoformat(timespec="milliseconds").replace("+00:00", "Z")


def fill_path(path, known_ids):
    def sub(m):
        entity, record = m.group(1), m.group(2)
        try:
            return known_ids[entity][record]
        except KeyError:
            raise SystemExit(f"placeholder {{{entity}.{record}}} not in manifest known_ids")

    return re.sub(r"\{(\w+)\.(typical|heavy)\}", sub, path)


def session_cookie(manifest, record, cookie_name):
    tier = TIER_FOR_RECORD.get(record)
    if not manifest or not tier:
        return None
    for u in manifest.get("users", []):
        if u.get("tier") == tier:
            return f"{cookie_name}={u['session_cookie']}"
    return None


def one_request(base, method, path, headers, body, timeout):
    host, port = base.hostname, base.port or (443 if base.scheme == "https" else 80)
    t0 = time.perf_counter()
    addr = socket.getaddrinfo(host, port, type=socket.SOCK_STREAM)[0]
    t_dns = time.perf_counter()
    sock = socket.socket(addr[0], addr[1], addr[2])
    sock.settimeout(timeout)
    sock.connect(addr[4])
    t_conn = time.perf_counter()
    if base.scheme == "https":
        sock = ssl.create_default_context().wrap_socket(sock, server_hostname=host)
    t_tls = time.perf_counter()

    conn = http.client.HTTPConnection(host, port, timeout=timeout)
    conn.sock = sock
    full_path = (base.path.rstrip("/") + path) or "/"
    data = json.dumps(body).encode() if body is not None else None
    hdrs = {"Accept": "application/json", **headers}
    if data is not None:
        hdrs.setdefault("Content-Type", "application/json")
    conn.request(method, full_path, body=data, headers=hdrs)
    resp = conn.getresponse()
    t_first = time.perf_counter()
    payload = resp.read()
    t_end = time.perf_counter()
    conn.close()
    ms = lambda a, b: round((b - a) * 1000, 2)
    return {
        "status": resp.status,
        "bytes": len(payload),
        "dns_ms": ms(t0, t_dns),
        "connect_ms": ms(t_dns, t_conn),
        "tls_ms": ms(t_conn, t_tls),
        "ttfb_ms": ms(t_tls, t_first),
        "download_ms": ms(t_first, t_end),
        "total_ms": ms(t0, t_end),
    }


def pct(values, p):
    s = sorted(values)
    k = (len(s) - 1) * p / 100
    lo, hi = int(k), min(int(k) + 1, len(s) - 1)
    return round(s[lo] + (s[hi] - s[lo]) * (k - lo), 2)


def severity(ttfb_p50):
    if ttfb_p50 > 1000:
        return "critical"
    if ttfb_p50 > 500:
        return "high"
    if ttfb_p50 > 200:
        return "medium"
    return "ok"


def summarise(samples):
    out = {}
    for key in ("ttfb_ms", "total_ms", "download_ms", "connect_ms", "tls_ms", "dns_ms"):
        vals = [s[key] for s in samples]
        out[key] = {"p50": pct(vals, 50), "p95": pct(vals, 95), "max": max(vals), "mean": round(statistics.mean(vals), 2)}
    out["bytes"] = samples[-1]["bytes"]
    out["statuses"] = sorted({s["status"] for s in samples})
    out["severity"] = severity(out["ttfb_ms"]["p50"])
    return out


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("endpoints", help="JSON file with the endpoint list")
    ap.add_argument("--base-url", required=True)
    ap.add_argument("--warmup", type=int, default=3)
    ap.add_argument("--samples", type=int, default=10)
    ap.add_argument("--timeout", type=float, default=30)
    ap.add_argument("--manifest", help="seed manifest with known_ids and seeded session cookies")
    ap.add_argument("--cookie-name", default="session_id")
    ap.add_argument("--cookie-env", help="env var holding a full Cookie header, used when no manifest user applies")
    ap.add_argument("--allow-writes", action="store_true", help="permit non-GET methods (local hosts only)")
    ap.add_argument("-o", "--output", help="write JSON here instead of stdout")
    args = ap.parse_args()

    base = urllib.parse.urlparse(args.base_url)
    if args.allow_writes and base.hostname not in LOCAL_HOSTS:
        raise SystemExit(f"--allow-writes is only allowed against a local host, not {base.hostname}")

    with open(args.endpoints) as f:
        endpoints = json.load(f)
    manifest = None
    if args.manifest:
        with open(args.manifest) as f:
            manifest = json.load(f)
    known_ids = (manifest or {}).get("known_ids", {})

    for e in endpoints:
        method = e.get("method", "GET").upper()
        if method not in SAFE_METHODS and not args.allow_writes:
            raise SystemExit(f"refusing {method} {e['path']}: pass --allow-writes (local only) to send writes")

    results = []
    for e in endpoints:
        method = e.get("method", "GET").upper()
        path = fill_path(e["path"], known_ids)
        headers = dict(e.get("headers", {}))
        cookie = session_cookie(manifest, e.get("record"), args.cookie_name)
        if cookie is None and args.cookie_env:
            cookie = os.environ.get(args.cookie_env)
        if cookie:
            headers["Cookie"] = cookie
        print(f"{method} {e['path']} [{e.get('record', '-')}]", file=sys.stderr)
        error = None
        samples = []
        started = utc_now()
        try:
            for _ in range(args.warmup):
                one_request(base, method, path, headers, e.get("body"), args.timeout)
            for _ in range(args.samples):
                samples.append(one_request(base, method, path, headers, e.get("body"), args.timeout))
        except (OSError, http.client.HTTPException) as exc:
            error = f"{type(exc).__name__}: {exc}"
        results.append({
            "name": e.get("name", e["path"]),
            "method": method,
            "path": e["path"],
            "record": e.get("record"),
            "window_start_utc": started,
            "window_end_utc": utc_now(),
            "error": error,
            "summary": summarise(samples) if samples else None,
            "samples": samples,
        })

    doc = {
        "base_url": args.base_url,
        "warmup": args.warmup,
        "samples": args.samples,
        "seed": {k: (manifest or {}).get(k) for k in ("scale", "seed", "seeded_at")},
        "measured_at_utc": utc_now(),
        "endpoints": results,
    }
    text = json.dumps(doc, indent=2)
    if args.output:
        with open(args.output, "w") as f:
            f.write(text + "\n")
    else:
        print(text)


if __name__ == "__main__":
    main()
