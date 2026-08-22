
## Problem

A car breaks down at the side of a road. The driver needs someone competent, nearby, and priced before they agree — and they need to know the van is actually coming. The mechanic needs jobs routed to them without a dispatcher on the phone. Neither side wants to type; one is stressed on a hard shoulder, the other is driving.

That shapes the whole product. Onboarding has to be near-typeless: photograph a licence, photograph a plate, confirm what was read. A request has to be a few taps — what's wrong, which car, where. And once a mechanic accepts, the driver's anxiety is answered by exactly one thing on screen: a van moving toward them with an ETA.

The build had a second constraint that mattered as much. The backend belonged to another team, it was live but incomplete, and the app was being written against it in parallel. Building carefully against a moving API is a different discipline from building a feature.

## What I built

A Flutter app that is three products by role:

- **Car owner** — sign up by photographing a driver's licence, add vehicles by photographing the plate, request help by issue type and location, browse nearby mechanics by distance, rating, ETA and price, hire, track the mechanic live on a map, pay, review, and keep a job history.
- **Mechanic** — profile with certifications and service list, dispatch inbox for nearby jobs, accept and advance a job through its status machine, post live location, and track earnings.
- **Admin** — account management and oversight.

Around fifteen feature areas including auth, OCR scanning, vehicles, requests, live tracking, payments, memberships, reviews, SOS, chat and notifications. 474 Dart files.

## How it was done

**The rulebook came before the code, and it is checked in.** The repo carries a `brain/` directory that is a mandatory context layer: nine rule files covering the API contract, data layer, state, layout, error handling, navigation, performance, UI standards and testing. Every file in the codebase falls into exactly one category, and every category names exactly one rule file that governs it — *there is no second rulebook*. Adding a feature starts by copying a registry template and filling it out completely before any Dart is written. This is a large app built with AI assistance, and that is precisely why the constraints are written down and enforced rather than held in someone's head: an agent that reads the rules produces the same code as the person who wrote them.

**Features are addressable.** The client's specification came across as a spreadsheet and was converted into a master feature list where every item has an ID — `AUTH-02`, `OCR-05`, `REQ-08` — with a priority and notes. Endpoints are registered against the feature IDs they satisfy. It means "is this built?" and "why does this endpoint exist?" both have lookup answers rather than opinions.

**The API contract has laws, and they are the interesting part.** No endpoint may be implemented before it is documented — the local API doc is a contract, not notes. Authorization headers are the interceptor's job; a repository that writes its own `Bearer` header is wrong by definition. Token refresh is automatic: the interceptor catches a 401, refreshes silently, retries the original request, and on failure clears tokens and redirects to login — no notifier or widget ever calls refresh. Role-based routing lives in the router guard reading session state, not in the login notifier, so no screen decides where a user "should" go. Every Dio call is wrapped so backend error shapes map to named `Failure` types and a raw `DioException` never reaches a widget.

**Responsive layout is exactly two modes and one widget.** Mobile under 600px with a bottom bar, tablet at or above with a navigation rail and master-detail. No desktop mode, no landscape-phone special case. Reading `MediaQuery.size` to make a layout decision is explicitly forbidden, as is `LayoutBuilder` used as a top-level layout switch — every decision flows through one `AppResponsiveLayout`. The reason is stated plainly in the rule: breakpoint logic scattered across widgets is how a codebase becomes unmaintainable, and the prohibition is easier to enforce than the cleanup.

**Every model and repository method ships with a test — for a specific reason.** The rule exists because of a real bug class caught in live testing: an endpoint returning a bare array where the model expected a wrapped object. That parses as an empty result rather than an error, so it fails silently and looks like "no data". The test requirement is documented with that finding attached, which is the difference between a coverage target and a rule people follow.

**I tested the live API myself and reported back.** Rather than assume the backend matched its schema, I ran real accounts against it and captured every payload. That produced a findings document — what works as documented, what is broken — and then a formal gaps report to the backend team, split into *missing*, *wrong* and *inconsistent* with priority order inside each. Its contents changed what the app could ship: the chat endpoints existed in the schema but returned nulls for every field; the route endpoint 404'd, blocking navigation waypoints; and there was **no file-upload endpoint anywhere in the API** — a single omission that blocked licence OCR at signup, mechanic certification documents, and profile photos, because the profile photo field was a plain URL string that assumed the image already lived somewhere. Password reset and OTP verification had no endpoints at all, behind screens that were already built.

Finding that early is the difference between a shipped feature and a screen that lies. It also meant writing down, in the app's own docs, which of its surfaces are currently non-functional and why — including the sign-up OTP step and the forgot-password screen.

**The rulebook records its own debt.** One of the five API laws is annotated with a note that the existing auth provider violates it and is scheduled for migration. A standards document that only describes the parts already clean is a marketing document; this one is honest about where the code hasn't caught up.

## Stack

Flutter, Riverpod with codegen, Freezed and `json_serializable`, Dio with auth/refresh interceptors, `go_router` with role guards, `flutter_hooks`, `easy_localization`, Mapbox, Firebase Cloud Messaging. Generated files are committed deliberately because CI runs no codegen step. Planned OCR services: AWS Textract for licences, a plate-recognition API, and a public vehicle-lookup API for make/model/year from a plate.

