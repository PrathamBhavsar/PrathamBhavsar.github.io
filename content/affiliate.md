
## Problem

Affiliate marketing is a promise about money made by a computer. A partner sends a visitor, that visitor buys something days later on a different device, and everyone has to agree on who earned what — the merchant, the affiliate, and a payout system that will move real currency on the answer.

Every hard part is a data-integrity problem. The tracking script runs on someone else's website, so you control neither the browser nor the page. Conversions arrive by webhook from systems that retry, which means the same sale can arrive twice. Amounts arrive in whatever currency the merchant sells in. And the ingestion endpoint is public by necessity — you cannot require authentication from a stranger's checkout page.

The part that concentrates the mind: if the database is down when a conversion webhook arrives, the money event is gone. The merchant's system got its `200`, will never retry, and an affiliate is silently not paid. A dashboard that shows the wrong number is a bug. A conversion that was never recorded is unrecoverable.

## What I built

A full affiliate platform — 23 backend modules, 28 tables — covering the path from a click on a partner's link to money leaving the merchant's account.

- **Tracking** — an embeddable tracker script served from the API, click capture with referrer, user agent, landing page and GeoIP country, and referral-code management.
- **Conversions** — public ingestion endpoints and a webhook handler, conversion types, chargebacks, and customer records.
- **Commercials** — commission plans, per-affiliate assignments, payments, payout requests, and multi-currency handling.
- **Analytics** — dashboards by referrer, operating system, browser, country and device, with rollups, exportable reports and a backfill path.
- **Three roles, three products** — a self-service affiliate portal, an admin dashboard, and super-admin above it, plus webhooks, notifications and audit logs.

The frontend is Next.js App Router with route groups per role; the backend is Bun and Elysia over PostgreSQL with Drizzle.

## How it was done

**Conversions are written to disk before they are written to the database.** The conversion logger is deliberately independent of the DB: every conversion and click is appended to a log file first, then synced to Postgres with retry logic, with a spool for what can't be written yet and counters tracking primary, secondary, spooled, failed and replayed writes. Logs are compressed on rotation and there are backfill commands to replay them into the database. It means a database outage degrades the product to "analytics are behind" instead of "affiliates were not paid" — and because the log is the source of record, the recovery is a replay rather than an apology.

**Money is integers, all the way down.** Every amount is stored as cents and every percentage as basis points, so ten percent is `1000` and $100.50 is `10050`. There is no float in the ledger. Multi-currency converts to a USD base using live exchange rates fetched at startup, with hard-coded fallback rates for when the rate provider is unavailable — because a missing exchange rate must not become a missing commission.

**The tracker script is generated, not static.** It is served dynamically from the API with the API URL injected at runtime, and startup validation refuses to boot a production instance whose tracker would point at localhost. That check exists because the failure it prevents is invisible: a tracker deployed to a merchant's live site, quietly posting clicks to nowhere, discovered only when someone asks why a campaign has no traffic.

**Public endpoints are separated from authenticated ones at the routing level.** Tracking ingestion has to accept requests from anonymous browsers on third-party sites, so it lives outside the JWT-guarded API surface entirely and gets its own rate limiter — one of four tuned classes covering standard traffic, auth, webhooks and registration.

**The frontend's auth has three layers that must stay in sync, and the interesting work is in the seams.** A token manager wraps storage and notifies subscribers on every change; an API client singleton holds a session version that increments on each token change; a React context handles login, logout and global 401s. The request cache is keyed by that session version, so an entry from a previous session can never be served to a new one — the class of bug where a user logs out, logs in as someone else, and briefly sees the previous account's data.

**Logout is a hard reload, on purpose.** A client-side navigation would leave module-level state — session version, request cache, in-flight refresh promise — alive in memory. A full reload guarantees the next session starts clean. And tokens are cleared *before* the logout call is sent, so a request already in flight cannot slip through carrying credentials that are supposed to be dead.

**Not every 403 means log the user out.** The API returns a role-specific forbidden error naming the role it saw. If that role doesn't match the one in local storage, it's a stale-JWT symptom, not an authorisation failure — so the client reloads to pick up the correct token *without clearing the session*. Treating it as a logout would be the obvious implementation and it would boot a user with perfectly valid credentials out of the dashboard at random. Distinguishing "your token is old" from "you may not do this" is the whole difference.

**Small things that stop dashboards lying.** GET requests are cached for five seconds per session, with paginated endpoints excluded and every mutation clearing the cache wholesale — enough to collapse duplicate fetches, too short to show anyone a stale figure. Dashboard pages fetch exactly once on mount via a ref guard, so React StrictMode's double-invoke doesn't double-count. Manual refresh uses a separate flag so re-fetching updates numbers in place instead of flashing the whole page back to a skeleton.

## Stack

**Frontend** — Next.js 16 App Router, React, TypeScript, Radix, client-side JWT with refresh-and-retry.
**Backend** — Bun, ElysiaJS, PostgreSQL, Drizzle ORM, JWT with role middleware, tiered rate limiting, GeoIP, live FX rates, file-based conversion logging with spool and replay, auto-generated Swagger.

