# Portfolio over SSH

The same projects as the website, as a terminal app: `ssh <host>`. No shell, no exec,
no forwarding, no password or key. Go + [Wish](https://github.com/charmbracelet/wish) +
Bubble Tea.

Content comes from `../projects.json` and `../content/*.md`, copied into `data/` at build time
so the site and the terminal never drift.

## Run locally

```bash
./sync.sh && go run . -addr :2222      # then: ssh -p 2222 localhost
```

## Build for a Linux server

```bash
./sync.sh && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/portfolio-ssh-linux .
```

## Deploy

Needs a server whose port 22 is free (move your own sshd to another port first) and a DNS
`A` record. Copy `bin/portfolio-ssh-linux` and `deploy/*` to the server and run
`deploy/install.sh` as root. The host key lives in `/opt/portfolio-ssh/host_ed25519`; keep it,
or visitors get a "host key changed" warning after every redeploy.

Limits: one new connection per IP every 2s (burst 5), 10 min idle timeout, 30 min per session,
128 MB memory cap in the unit.
