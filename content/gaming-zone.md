
## Problem

A PC café runs on one question all day: which station is free, and for how long. Most venues answer it with a whiteboard and a cash drawer. That works until a booked slot collides with a walk-in, until a player insists they only played an hour, until the internet drops mid-shift and the counter has no idea what any station is doing.

The billing problem is worse than the scheduling one. A café charges for time, so its revenue is a measurement — and if that measurement lives in someone's memory, every dispute is a coin flip and the owner never learns what the venue actually earned. Chain owners have the same problem multiplied: no way to compare Tuesday at one branch against Tuesday at another.

Gaming Zone is a real-time resource scheduling and billing system for physical venues: players book and pay from their phone, staff run the floor from a dashboard, an agent on each gaming PC enforces and records the truth, and the server reconciles all three into a billing record that survives being argued with.

## What I built

**Player app** — discover nearby stores, browse live station availability, book a slot, pay, check in, watch a running session timer, top up and spend wallet credits, redeem campaign offers, get realtime notifications, and dispute a charge they disagree with.

**Admin dashboard** — the same Flutter codebase, a different product. Live floor grid, session start/end/extend/pause for walk-ins, station and system-type management, pricing rules, billing and payments, credit issuance, campaigns, dispute resolution, and store analytics — every screen scoped to the one store that admin belongs to.

**Multi-tenant backend** — 20 domain modules over Bun and ElysiaJS covering auth, stores, systems, pricing, bookings, sessions, billing, payments, credits, campaigns, notifications, disputes and analytics. Postgres via Drizzle, Redis for cache and locks, FCM for push, WebSockets for live floor state.

**PC agent protocol** — each station authenticates with its own key, heartbeats every 30–60 seconds, receives lock/unlock commands, and keeps running its local countdown when the network is gone.

72 screens, mapped to routes and to the endpoints behind them. 308 Dart files, 97 TypeScript files.

## How it was done

**Two clients, one codebase, one shared model layer.** Player and Admin are separate products that share `lib/models/` and every piece of core infrastructure. The rule that keeps that honest is a fixed layering — `Widget → Notifier → Repository → ApiClient → Backend` — and no API call is allowed above the notifier line. Riverpod without codegen, hand-written `fromJson`, no `freezed`, no `build_runner`: deliberate, because a 300-file app that generates half its own source is an app you can't read.

**`store_id` is the isolation boundary, and it's enforced by where the ID comes from.** Player routes nest under `/stores/:storeId/...` so the tenant is in the URL. Admin routes take the store from the admin's JWT — an admin cannot address another store's data because they cannot express it. Agent routes take it from the URL but must match the hash of the station's own key. Global identity tables stay unscoped so one player account works across every venue.

**Timing is hybrid, because the network is not a dependency you get to assume.** The agent owns the countdown, locks the machine when time expires, and buffers events while offline. The server owns the official record: billing, pricing, credits, conflict resolution. On reconnect the agent posts a batch of buffered logs, the server appends them as `session_logs` and reconciles session state from any end events in the batch. Partial failures don't abort the batch. A station that lost its internet for two hours still bills correctly.

**Money is append-only.** `billing_ledger`, `credit_ledger` and `admin_overrides` have no `updated_at` and are never updated or deleted — a correction is a new row. Credit balance is derived by summing the ledger, never stored as a mutable number. An admin correcting a charge must supply a reason, enforced as `OVERRIDE_REQUIRES_REASON`, and it writes an audit row alongside the new charge. Deactivation is `is_active = false` rather than deletion, so last quarter's revenue still reconciles after a station is retired.

**Race-prone operations are serialized in Redis.** Booking a slot and claiming a system both go through `SET lock:{key} PX ttl NX`, so two players tapping the same station at the same moment produce one booking and one clean `BOOKING_OVERLAP` rather than two half-valid rows. The cache layer next to it fails soft on purpose — a Redis error degrades speed, never correctness.

**Pricing is a rule engine shared with billing.** Rules match on day of week, time window, member tier, and optionally system type; highest priority wins; the winning multiplier applies to the system type's base rate. The engine that previews a price before booking is the same one that computes the final charge, and it writes `applied_rule_id` and `applied_multiplier` into the ledger — so any charge can be explained months later by pointing at the rule that produced it.

**The bug worth telling.** A production-readiness audit found the app's flagship promise — see how much session time is left — was quietly fake. The backend returned no end time for an active session, and the frontend covered for it in two different places in two different ways: the detail screen fell back to `startedAt + 2 hours`, so a one-hour booking displayed "2h left" and a walk-in counted down into negative numbers while the player was still playing; the home banner returned zero instead, so the same live session simultaneously read "0:00 remaining" in one place and "1:58 remaining" in another.

The fix was not a better guess. The backend gained a real `expectedEndAt`, the frontend's fabricated fallback was deleted, and the field was made nullable and threaded through honestly — a countdown when the end time is genuinely known, "2h 14m elapsed · open session" when it isn't. Wiring it surfaced a second dead path: the extend-session WebSocket event had never included the new end time, so the app's live-update-on-extend code had been correct and receiving nothing for its whole life. A guessed default had been hiding a missing field, and a missing field had been hiding a dead code path. Fabricated data doesn't degrade gracefully; it just lies quietly until someone measures it.

## Stack

**App** — Flutter, `flutter_riverpod` (no codegen), `go_router`, `http` behind a shared `ApiClient`, hand-written models, custom responsive primitives and theme tokens, Firebase FCM.
**Backend** — Bun, ElysiaJS, PostgreSQL, Drizzle ORM, Redis (cache, distributed locks, rate limiting), WebSockets, Firebase FCM, Nodemailer.
**Agent** — per-station key auth, heartbeat + command channel, local countdown, offline log buffer.

