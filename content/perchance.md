
## Problem

A client had a working AI generation platform and an interface people had to forgive. The tools were there — image generation, character chat, a turn-based RPG, illustrated stories — but each one sat on its own page, in dense utilitarian chrome, with no thread running between them. You picked a model in one tool and picked it again in the next.

The brief was a full front-end replacement. It arrived as a ticket containing two waves of requirements pasted together that partly contradicted each other, describing a product whose backend was owned by another team and whose API shape nobody would confirm. Two obvious ways to fail: build to the ticket literally and ship the contradictions, or invent the API and produce a front end that would need rewriting the day the real one appeared.

So the first deliverable wasn't code. It was a reconciled statement of what the product actually is, written down as the canonical source, with the raw ticket demoted to an input and the four genuinely unresolved product decisions listed as open questions rather than quietly settled.

## What I built

A complete dark-mode SPA — every page in the sitemap built, no stubs — organised around one mechanic: **pick a base model and a LoRA once, and that choice follows you into every tool.**

- **Browse** — the entry point. Base models laid out rather than hidden in a dropdown, searchable LoRA reels by category, and a sticky dock carrying the loaded asset with quick-actions routing it into any of the four tools.
- **Image and video generators** — prompt composition with parameters kept simple by default behind an Advanced toggle, and staged progress per result tile.
- **Character chat** — a four-voice composer (character, you, narrator, image), a transcript that pins to the newest line until you scroll up, character JSON import/export with client-side validation, and a focus mode.
- **RPG and Story** — setup wizards opening into turn-based play and linear illustrated narrative, both illustrated by the loaded LoRA.
- **Credits economy** — balance, spend, recharge countdown, activity ledger, and an out-of-credits modal offering three ways forward ordered by cost to the user: wait for the free refill, watch an ad, or buy.
- **Source editor** — the platform's list-DSL generator page rebuilt, with a genuinely live preview.
- **A published design system and motion system**, both documented as specs with a live specimen route.

150 TypeScript files, 10 routes.

## How it was done

**The design system has a thesis, not a palette.** The product is a *ledger* — a registry, a printout. Generation is bookkeeping: every image, turn and credit is an entry with an address, a cost and a timestamp. That single idea produces the rules: flat, ruled, mono-labelled, `--radius: 0px` everywhere. It explicitly refuses the AI-product default of glassy rounded cards and glowing gradient heroes. Purple is present, but it is ink, not atmosphere — colour marks state and action, never decorates. Dark mode isn't a style choice either; the use scene is people judging generated imagery at night, so the chrome recedes to let the work be the brightest thing on screen.

**Motion is a system of roles, not a pile of transitions.** One idea — *everything arrives the way an image does* — since the product's subject is diffusion. Panels, rows and prices de-noise into existence: blur clearing, saturation rising, and a generated result lands the same way, slower and on purpose. It's enforced by a strict split: a component declares *what* is happening (`m-enter`, `m-pending`, `m-resolve` — thirteen semantic roles) and never a duration, curve or keyframe. That constraint is what keeps motion reviewable instead of drifting into sixty hand-tuned transitions nobody can audit. Hover is the deliberate exception: it draws a rule under the target rather than de-noising, because a hover is an acknowledgment, not an arrival. One global `prefers-reduced-motion` escape at the bottom of the file, never re-implemented per component.

**Honesty was a build rule, and it shaped the components.** The backend wasn't available, so nothing could actually generate — and the rule was that the UI must never pretend otherwise. Canned chat replies are never labelled AI-generated. The auth modal says plainly there's no account system behind it. The profile shows "Guest" when signed out rather than mocking a logged-in user. Generated images are stated to expire, because they genuinely do server-side, instead of implying permanence. This is harder than faking it and it is the reason the app can be wired to a real API without a rewrite: every simulated seam is marked and local.

**Where a placeholder would have been dishonest, the real thing was cheaper.** The editor's preview evaluates a small genuine interpreter for the platform's list DSL — output/item lists, references, inline alternates, weights — so typing into it actually produces output rather than replaying a recording. The landing page's five product demos are live components on a scripted timeline, not recorded video: they can't go stale, they need no asset pipeline, and each panel links into the tool it demonstrates. The "generating" animation is a WebGL2 dithering shader tuned so its internal resolution matches the static grid's dot size after the CSS stretch between them.

**Primitives were rebuilt rather than reskinned.** Select and Slider are Radix under custom track, range and thumb — not a native `<select>` popup or a repainted `input[type=range]` — and even the scrollbar is styled to match the system. Same principle throughout: our component, not device chrome. The placeholder art is a seeded generator that takes the loaded LoRA's hue, so a gallery of twenty thumbnails reads as distinct artwork and results visibly inherit the asset you chose. The pipeline is *visible in the output* rather than merely asserted.

**One bug worth the note it got.** The credits hook has to answer "did that spend go through?" synchronously, and reading the result out of a `setState` updater is unreliable — React 18 only sometimes runs the updater eagerly, so a spend that would succeed could report failure and pop the out-of-credits modal over a perfectly healthy balance. It's decremented against a ref up front and reconciled with committed state in an effect. The failure mode was inherited from an earlier version of the hook; it's the kind of bug that reproduces one time in five and gets dismissed as a fluke.

**Groundwork before pixels.** The reference platform was studied into a 38-concept knowledge graph, with a gap map cross-referencing every ticket requirement against what the reference actually does — including flagging requirements with no precedent to check against, and one tool whose reference implementation turned out to be an unimplemented stub, so its interaction had to be extrapolated rather than copied. The reference was treated as a *functional* source only: adapt the interaction, never the visual language.

## Stack

React, TypeScript, Vite, Tailwind, Bun. Radix primitives (Select, Slider, Tabs), WebGL2 shader, CSS custom-property design tokens, React Context for selection / credits / auth state, `localStorage` persistence. Structured shadcn-style with a `components/ui` primitive layer.

