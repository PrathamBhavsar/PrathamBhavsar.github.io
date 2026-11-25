## Why

I run a lot of coding agents at once, on my PC and on a Linux VPS. The IDE I was using for it was an Electron app that sat at roughly 2 GB of RAM with only a few agents going. Kestrel is my replacement: a terminal workspace where each agent gets its own git worktree, and a small host daemon does the actual work.

It's a personal tool, one user and a couple of machines, so the repo is private.

## What's in it

A host daemon runs agents (claude, codex, agy and others) inside pseudo-terminals. A TUI sits on top of it, and the daemon can live on another machine, so a VPS feels like a local session. A CLI mirrors the commands of the tool it replaces, which means my existing scripts and skills keep working. The screen shows each agent's state (working, idle, blocked, done), context use, cost, and the diff it has made so far.

The workspace is 25 Rust crates, around 51,000 lines. Terminal handling, the PTY daemon, git, diffs, a resource monitor, cost tracking and the RPC layer are each their own crate.

## Decisions I'd defend

- **Agents get less than I do.** Each agent talks to a second socket that exposes fewer methods: no secrets, no settings. Their environments are scrubbed of my own tokens. Sockets are `0600` and the home directory `0700`.
- **A permission prompt on screen goes to a human.** It's never answered automatically.
- **Idle agents can hibernate** and wake when a message arrives, so only the ones doing work cost memory.
- **Nothing merges without the gate.** One command runs fmt, clippy, tests, end-to-end checks and a parity check against the original tool, and the merge command refuses if it hasn't passed for that commit.

## Where it stands

Started on November 16, 500 commits since. It's been run on macOS against real agents. Windows and a remote host are the next things to prove, and the README says so.
