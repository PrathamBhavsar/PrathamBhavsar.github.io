
## Problem

A consumer AI app lives or dies on how it *feels* between the interesting moments. The generation is the product, but most of the session is browsing, scrolling galleries, opening a character, going back, opening another. If that connective tissue stutters, no amount of model quality saves it.

Two things make that hard on mobile. Media is everywhere — every screen is wall-to-wall generated imagery, and images that are technically cached still feel slow if they're decoded wrong. And the data is mutable in ways users expect to see instantly: generate an image and it must already be in the character's gallery when you navigate there, edit a character and the detail screen must not still show the old one.

Meanwhile the app is 75 screens deep, age-gated before first use, and has a paid credit economy. The interesting engineering here isn't any single feature. It's keeping a surface that large consistent.

## What I built

A Flutter app across iOS and Android — companion chat, character discovery and creation, image and video generation, galleries, group chats, following, a credit economy and subscriptions, all behind an age-verification gate.

- **Discover and chat** — browse characters, one-to-one and group conversations, voice, gifts, and a recent-chats list.
- **Character creation** — a guided builder with traits, personas, poses and voice, ending in avatar selection and publication.
- **Image and video studio** — text-to-image with model and LoRA selection, in-chat generation, history, galleries, privacy controls, favourites and voting.
- **Account layer** — guest mode, auth with OAuth callbacks and email verification, subscriptions, credit purchase, profile and following.

75 screens. 764 Dart files in the main app, with a parallel client build alongside it and a Next.js web companion covering chat, image and video generation.

## How it was done

**Caching is centralised, automatic, and deliberately minimal.** Two runtime-memory caches, both transparent to feature code: a TTL response cache driven by a Dio interceptor, and an image decode cache. Repositories were not modified to benefit from either. One policy file is the single source of truth for what is cached, for how long, and what each mutation invalidates, and the cache interceptor is registered *first* in the Dio chain so a hit short-circuits before auth and retry ever run.

The philosophy is the unusual part: the app leans toward fresh on every page load. The "stable per-user data" band is **30 seconds**, not the usual eight or ten minutes. Pull-to-refresh bypasses the cache entirely. Every mutation eagerly evicts what it touched. The cache exists to collapse back-to-back refetches during normal navigation, not to serve yesterday's data — only genuinely static config is held long, and media bytes are the deliberate exception.

**Caching a "processing" status is a bug, not a tradeoff.** In-progress generation, order status and the age-verification status poll are all excluded from caching by rule. The reasoning is written down because it isn't obvious in the moment: the poll re-issues the same GET, hits the cache, and can therefore *never observe the transition it exists to detect*. It would look like a hung generation, and the cache would be the last place anyone looked.

**The invalidation wildcard bug was worth a rule of its own.** Eviction patterns match one path segment at a time and require the same segment count, so a rule for `/foo/*` does **not** cover `/foo/{id}/bar`. That single subtlety produced a run of real bugs — group-chat rename and edit, gift sending, voice change, character edit and delete, avatar selection finalising creation, image privacy, vote and favourite — every one of them a nested path that was one rule pattern short of matching, and every one presenting as "the UI didn't update after I did the thing." The rule now is to write the real path out segment by segment against the pattern before assuming an existing wildcard covers it.

**Stale lists are fixed by refetching, never by patching the cache.** The recent-chats list reorders every time you send a message. The tempting fix is to reorder the cached list locally; the rule forbids it. Instead the list is short-TTL *and* every send or delete evicts it — so browsing away and back serves from memory, and the moment you actually chat it refetches once in the correct order. The server stays the source of truth and the client never drifts from it.

**Images were slow because they were decoded at source resolution.** Disk caching does not save you here: a 2000×3000 image in a 120px slot decodes to roughly a 24 MB bitmap, fills the in-memory image cache, forces eviction, and the next view re-reads from disk and decodes it all over again. That is why images feel sluggish *even when "cached"*. Decode size is now capped to display size, computed from the slot width and device pixel ratio.

**All media goes through one widget, and per-call-site overrides are banned.** Every network image, video, asset, GIF and AVIF renders through a single `AppMedia`, which delegates to the one place in the app that owns a video controller and the decode-capped image provider. It has exactly one loading state and one error state — and passing a custom placeholder or error widget at a call site is a rule violation, because that fragmentation is precisely what the rule exists to prevent. The one meaningful variation is semantic rather than visual: media representing a *named* thing a user recognises falls back to initials on a tinted gradient, while anonymous content falls back to a broken-image icon. A failed avatar shouldn't lose the character's identity.

**Ending a session has one entry point, wired to every path that ends one.** Clearing tokens locally is not enough to clear caches, because logout and account-deletion network calls are best-effort and can fail while the session genuinely ends anyway. So a single `clearAll()` wipes the response cache plus in-memory and on-disk image caches, and it is called from every path that terminates a session — voluntary logout, forced logout after refresh failure, account deletion, and the token-expiry fall-back-to-guest path that never touches the logout endpoint at all. Missing any one of them leaks one user's data into the next user's session.

**The age gate is a routing stack, not a dialog.** Verification has its own route stack outside the app shell, with success and error callbacks and a status poll for the pending-to-verified transition — the poll being one of the endpoints explicitly excluded from caching, for the reason above.

## Stack

Flutter (iOS + Android), Riverpod with `@riverpod` codegen, `go_router` with auth guard and shell routes, Dio behind repositories with a layered interceptor chain, Freezed and `json_serializable`, `flutter_secure_storage`, `cached_network_image` with a decode-capped provider, WebSockets via Reverb for realtime, Firebase. Web companion in Next.js.

