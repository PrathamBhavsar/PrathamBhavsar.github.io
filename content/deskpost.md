## What it is

Deskpost is a shared inbox for support tickets. Several projects feed one queue. Each ticket is a thread, the team can leave internal notes that customers never see, and there are canned replies, attachments and a live indicator of who else is looking at a ticket. Customers can use an embeddable widget or just send an email.

You host it yourself. A Rust API, a Next.js UI and one Postgres 16 database, started with Docker Compose or run as a systemd service behind Caddy. It's AGPL-3.0, and the site and docs are at the link on this page.

## Where it came from

The UI started as the frontend of an earlier ticket product of mine (SendDesk, with a Node backend). I kept the interface and rebuilt everything behind it in Rust: the API, the database layer, mail and file storage, split into `api`, `core`, `db`, `mail` and `storage` crates.

## Mail without AWS

I didn't want a self-hosted tool that needs a paid email service to be useful. Outgoing mail goes over plain SMTP. Incoming mail is either polled over IMAP or pushed to a signed webhook, depending on what your provider supports. There's a guide covering Gmail, Brevo, Resend and Mailgun settings, and a Mailpit container in the compose file so you can watch mail arrive while developing.

## Making it something other people can run

Most of the work after the code was the boring part. Every environment variable is listed with defaults in `.env.example`. There's a backup script, a smoke test to run after each deploy, a security checklist, and an operations doc for upgrades and rotating secrets. CI runs fmt, clippy, tests and `cargo deny`. Release builds ship with a one-line installer.
