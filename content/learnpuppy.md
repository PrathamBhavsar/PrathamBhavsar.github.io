## What it is

LearnPuppy is a study tool I built for myself. You give it a topic, it asks a few questions to see what you already know, writes a plan, and teaches the first lesson. Whatever you learn goes into a knowledge graph that sticks around between sessions, and a spaced-repetition scheduler (FSRS) brings things back for review just before you'd forget them.

It has no AI of its own. It drives the `claude`, `codex` and `agy` command-line tools that are already logged in on your machine, so there's no API key to hand over and nothing leaves the computer except what those tools send.

## How it's put together

`learnpuppy serve` starts a Rust server (axum) on `127.0.0.1`. The React app is compiled into the same binary, so the install is one file. You get a one-time launch URL, the server swaps that token for an HttpOnly cookie, and it rejects any request with a foreign `Host` or `Origin`. Each workspace (a course, a subject) has its own isolated store.

The server generates an OpenAPI spec, and the TypeScript client types are generated from it. If I change an endpoint and forget the UI, `tsc` fails instead of the page breaking at runtime.

## Things I'm glad I did

- **Several Claude profiles act as separate rate-limit buckets.** When one runs dry the request fails over to the next, and that path has a test.
- **The gate is one command.** `make gate` runs fmt, clippy, tests, and checks the OpenAPI file hasn't drifted. Playwright runs both against mocks and against the real server, and there are perf budgets with a longer soak run.
- **I deleted two UIs.** The first versions were a Dioxus desktop app and a Ratatui terminal UI. I dropped both for the web app and wrote down why in an ADR. The TUI is still on a branch.

## What isn't proven

The README lists the gaps. `agy` was out of quota when I verified, so it's untested as the teacher, though failing over away from it is tested. The graph legend's four node kinds are cosmetic for now because the wire format only carries the source kind. Review-all was checked against a scripted engine, not the real CLIs.
