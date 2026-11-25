## The problem

The content for the hubs comes from scrapers and video converters, and those run on a handful of home PCs, not on the servers. Someone has to decide what runs where, notice when a machine drops off, and hand finished work to the right hub (OctopusHub for video, ImageHub for images). Before Beacon that was a lot of scripts and checking by hand.

## What it is

Beacon is the control plane. A Go server holds the queue, the data model and the admin UI. A small agent on each PC pulls jobs and runs workers. Workers are Python programs, and they never talk to the server: they talk only to the agent on `127.0.0.1`, through a small SDK that uses nothing outside the standard library. That keeps a scraper's code simple and means a worker can't do anything the agent doesn't allow.

The admin is React with TanStack Router and Query, and it's embedded in the server binary. It has a pipeline editor (React Flow) where you wire a scrape step to a convert step to a delivery step, and a live view of jobs, nodes and sources over server-sent events. The server is about 150 Go files and 19 migrations, using Echo, sqlc and pgx, with no ORM.

## Rules I wrote down and kept to

- **Docs first.** The `docs/` folder is the source of truth, one file each for the API, the queue, the agent, workers and the data model. If the code and a doc disagree, one of them gets fixed on purpose.
- **Soft deletes everywhere.** There's no hard delete through the API, and a test checks that every read of a soft-delete table filters deleted rows.
- **The agent runs a closed list of commands.** No shell, no arbitrary code.
- **Every write by a user goes to the audit log.**
- **Migrations expand, then contract.** A column is never dropped in the same deploy that stops using it.
- **v2 stays v2.** Features marked for later aren't built early, even when the doc sketches them.

## Deploy

Beacon listens on loopback behind nginx, and agents reach it over Tailscale. Docker runs only the Beacon binary. Postgres, TLS and Tailscale are the host's own services. Agent releases are built and signed in CI, and a new machine joins with one command.

The code isn't public. The screenshot is the local build with mock data and made-up source names.
