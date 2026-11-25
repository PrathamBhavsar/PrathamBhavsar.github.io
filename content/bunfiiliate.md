## Why I wrote it

Affiliate software mostly comes in two kinds: SaaS that takes a cut of your sales, and old monoliths that fall over when click volume goes up. I'd already built a private affiliate platform in TypeScript for a client, and I wanted an open-source version where the two hard parts were done properly: taking clicks fast and getting the money right.

## The click path never waits on the database

Requests to the tracker are checked against an in-memory snapshot of programs and referral codes. Postgres `LISTEN/NOTIFY` tells the tracker to reload it when config changes. A valid click is appended to a segmented write-ahead log on disk, with CRC framing and a group-commit fsync, and the response goes out. A separate task batches the log into Postgres with `INSERT ... UNNEST ... ON CONFLICT DO NOTHING`. If the database is down, clicks still land in the log and get replayed. If buffers fill up, the tracker returns 503 and sheds load instead of growing memory.

My benchmark run, on one laptop with the load generator and Postgres sharing the machine, so treat it as a floor:

- fsync on every group: about 1,500 requests a second, p50 around 13 ms
- fsync off: 9,000 to 28,000 requests a second, p50 under 1.5 ms

The docs say what the fsync setting costs you if the host crashes.

## Money is a ledger

Every movement posts to an append-only double-entry ledger. Amounts are integer cents. Database triggers reject updates and deletes on ledger rows, and deferrable constraint triggers check that each transaction sums to zero. Affiliate balances are derived from posted entries, never stored and edited. Commission is calculated in basis points with round-half-up in `i128`, and the pure math lives in a crate with no I/O so it can be property-tested.

## The rest

Workspaces, programs, affiliates and referral codes, with commission plan overrides, first-touch and last-touch attribution, bot and velocity filtering, payout batches with an approval step, signed outbound webhooks through a transactional outbox, and Stripe and Paddle webhook verification. Sessions are cookie JWTs with rotating refresh tokens, and reusing an old refresh token revokes the whole family. It ships as two binaries, `tracker` and `api`, behind Caddy.
