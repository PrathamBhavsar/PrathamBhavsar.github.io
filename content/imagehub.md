## What it does

ImageHub is the backend behind a set of image sites. Comics and galleries arrive as scraped ZIP files. ImageHub imports them, uploads the pages to Cloudflare R2, and distributes them to content domains, either as files pushed over SFTP/rsync to a web server or as public R2 URLs. It handles four content types (video, comic, manga, gallery), and each type has its own rules for what gets distributed and how.

It's a Go port of a Python service, rebuilt around images instead of video. The whole thing is one Go binary with the React admin embedded using `go:embed`, and the job workers run inside the same process. Deploys are a GitHub Actions run that builds on the server over SSH and restarts a systemd unit.

## My part

The repo has about 60 commits and I wrote most of them. Among them:

- A ZIP upload endpoint for imports, and a bulk export endpoint so outside sites can sync content.
- Rewriting image URLs inside each distributed `data.json` so pages load from the right domain.
- A JSON patch API for content metadata, plus category and tag filters in the public API.
- Choosing which server disk each domain writes to, and public and private sites.

## The bugs worth telling

**Full disks.** A transfer to a full disk used to look fine on the surface and fail later. I made it fail fast, refresh disk usage on a timer, and show disks in the UI at all.

**Swallowed errors.** Some distribution failures were being caught and dropped, so content looked distributed when it wasn't. Transfer errors now bubble up, and the location a file was sent to is recorded so you can see it.

**Redirected domains.** Pages were published at a URL the domain redirected away from. Publishing now uses the URL the page actually lives at.

The code is in a client's GitHub org, so there's no public link. The screenshot is a local build running on mock data.
