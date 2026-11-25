## What it is

OctopusHub is a FastAPI system that holds scraped video and image content, cleans up the metadata, and distributes it across many domains. Ingest uses configurable JSON mappings. Storage is Cloudflare R2. Distribution goes to web servers over SSH/SFTP, with Celery and Redis running the jobs. The admin is a React app deployed on Cloudflare Pages. It's the system ImageHub was later split out of.

This is a team codebase with several contributors, and most of the commits aren't mine. I joined in 2026 and worked on the video side.

## What I did

- **Trim job.** Distributed videos pile up in several quality tiers and eat disk. I added an admin job that keeps only the best N tiers of a video. It starts as a dry run, and it only demands 20 GB of free disk on the server, because trimming frees space instead of using it.
- **Nested playlist folders.** Some ZIPs put the master playlist in the root with the quality folders underneath, and others nest them deeper. Distribution and trimming both assumed a flat layout, so I fixed the path handling for both, including where the timeline thumbnails sit.
- **Creator images.** Host the creator image from the ZIP and store its URL, backfill it for content that was already imported, and set one by hand on a single item.

The code lives in a client's GitHub org, so there's no link. The screenshot is a local build with mock data.
