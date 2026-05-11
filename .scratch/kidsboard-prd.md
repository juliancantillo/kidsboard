---
title: Kidsboard — RPG-style household activity tracker
status: needs-triage
type: PRD
created: 2026-05-10
---

# Kidsboard PRD

## Problem Statement

Parents of multiple children want a fun, visible, motivating way to track and reward everyday activities (chores, school, hygiene, reading, etc.) at home. Existing options are either too childish (sticker-chart apps aimed at toddlers), too abstract (generic to-do apps with no game-feel), or too punitive (behavior-tracking apps focused on negatives). What's missing is something that:

- Feels like an RPG character sheet rather than a sticker chart.
- Lets each kid see meaningful *progress* — levels, milestones, and a personal identity ("I'm a Level 12 Dishwasher") rather than just a star count.
- Treats achievements as a story: humorously-named milestones that capture moments worth celebrating.
- Gives kids a tangible spendable currency (points) they can redeem for real-world rewards the parent has set up.
- Runs on a household device (fridge tablet / family computer) without auth ceremony, account creation, or cloud dependencies.
- Scales to a small fixed roster of kids (4 today, N tomorrow) and stays useful as kids grow.

## Solution

Kidsboard is a single-binary Go web app, run on a trusted device in the home, that models each kid as an RPG character with per-category skill levels, XP, points, achievements, and a rewards shop.

The parent logs activities for kids (Dishwashing, Homework, etc.). Each activity instance grants XP (progress) and points (currency) per its activity type's reward shape. Cumulative activity drives per-category levels and unlocks achievements — composite, multi-rule milestones with humorously-flavored names and titles ("Antman of the Nanouniverse"). Kids browse a profile page showing their character sheet: avatar, per-category levels with progress bars, recent achievements, the next milestones they're closest to, their points balance, and a shop of available rewards. Kids can request a redemption from the shop; the parent approves or rejects it.

The data model and service layer arrive first, before any UI. The UI ships next as Tailwind-styled Go html/template pages with an RPG visual treatment (curated avatars, color accents per kid, character-sheet layouts).

## User Stories

### Parent — daily operation

1. As a parent, I want to log a single activity for a specific kid in one or two clicks, so that recording an event during a busy moment isn't a chore.
2. As a parent, I want to log multiple instances at once (e.g. "Mia did Dishwashing × 3"), so that I don't have to click three times when batch-logging.
3. As a parent, I want to backdate an activity I forgot to log, so that yesterday's chore still counts toward Mia's totals.
4. As a parent, I want to add a short free-text note to a logged activity, so that I can remember context like "did the whole sink" later.
5. As a parent, I want to void an activity I logged by mistake (with a reason), so that the XP/points stop counting without losing the audit trail.
6. As a parent, I want voids and corrections to NOT retroactively revoke achievements the kid has already earned, so that earned moments stay earned.
7. As a parent, I want to grant or deduct ad-hoc points outside the catalog with a reason, so that I can recognize a one-off behavior or correct a balance.

### Parent — kid management

8. As a parent, I want to create a new kid with a name, an RPG-style avatar (chosen from a curated set), and a color accent, so that each kid has a distinct visual identity.
9. As a parent, I want to reorder kids on the home screen, so that the kid-selector reflects the order I want.
10. As a parent, I want to edit a kid's name, avatar, or color, so that visual identity can evolve.
11. As a parent, I want to archive a kid (soft-delete) without losing their history, so that the kid disappears from the active home screen but their past records are preserved.
12. As a parent, I want to unarchive a kid, so that I can reverse an archive decision.

### Parent — configuration

13. As a parent, I want to create activity categories (e.g. Chores, School, Hygiene) with names, descriptions, and visual flair (icon, color), so that the skill tree reflects what matters in our household.
14. As a parent, I want to create activity types within each category (e.g. "Wash Dishes" under Chores) with specified XP-per-unit and points-per-unit, so that I can balance rewards based on effort.
15. As a parent, I want to define some activity types as XP-only (no points), so that mandatory baseline behaviors progress the kid without inflating the currency.
16. As a parent, I want to archive categories and activity types when they become irrelevant, so that the picker stays clean while history remains intact.
17. As a parent, I want to create achievements with humorous names, descriptions, and an optional earned title, so that milestones feel like personal moments rather than generic badges.
18. As a parent, I want to define multi-rule achievements with AND or OR semantics, so that I can express both "do X *and* Y" and "do X *or* Y" milestones.
19. As a parent, I want each achievement rule to target a specific category or all categories globally, so that I can write both category-focused and cross-category milestones.
20. As a parent, I want each rule to use a metric of count, XP, points, or level, so that I can express thresholds in whatever dimension is most fitting.
21. As a parent, I want to set a points bonus on each achievement, so that unlocking feels rewarding beyond the title.
22. As a parent, I want to create rewards in a shop with a name, description, and points cost, so that kids have a tangible carrot to save for.
23. As a parent, I want to deactivate or archive rewards, so that out-of-season offers disappear from the kid's shop without breaking historical redemptions.

### Parent — redemption approval

24. As a parent, I want to see pending redemption requests in a clear queue, so that I can act on them quickly.
25. As a parent, I want to approve a redemption with one click, so that the kid's points are deducted and they know the reward is theirs.
26. As a parent, I want to reject a redemption with an optional reason, so that the kid understands the decision.
27. As a parent, I want to cancel a previously approved redemption, so that points are returned to the kid if the reward never happened.
28. As a parent, I want the system to recheck the kid's balance at approval time, so that a redemption can't be approved if points have been spent elsewhere in the meantime.

### Kid — profile / character sheet

29. As a kid, I want to select my avatar from the home screen and land on my profile, so that I don't have to log in.
30. As a kid, I want to see my per-category levels as visible progress bars, so that I know which skills I've leveled up.
31. As a kid, I want to see my current XP within each category and how much I need for the next level, so that I have a concrete near-term target.
32. As a kid, I want to see my points balance prominently, so that I always know my spending power.
33. As a kid, I want to see my available balance separately from total balance when I have pending requests, so that I understand what's "spoken for."
34. As a kid, I want to see my earned achievements with their humorous names, descriptions, and titles, so that I can revisit my proudest moments.
35. As a kid, I want a "NEW!" indicator on achievements earned since I last viewed my profile, so that I notice fresh unlocks.
36. As a kid, I want a celebration moment when an achievement is unlocked during an active session, so that the unlock feels like a real event and not just a silent stat change.
37. As a kid, I want to see the top three achievements I'm closest to earning, with per-rule progress meters, so that I know exactly what to do next.
38. As a kid, I want a per-category "skills" view showing each category's closest unearned achievement, so that I can pick a category to focus on.
39. As a kid, I want to see my recent activity history, so that I can confirm my work is being tracked.

### Kid — shop / redemption

40. As a kid, I want to browse the rewards shop showing each reward's name, description, and points cost, so that I know what I can save for.
41. As a kid, I want rewards I can't currently afford to be visibly distinguished from those I can, so that I know what's reachable.
42. As a kid, I want to request a redemption with one click, with the system rejecting the request if I can't afford it, so that I'm never given false hope.
43. As a kid, I want my pending requests to reserve their cost against my balance, so that I can't accidentally request two rewards that together exceed my balance.
44. As a kid, I want to see my pending requests on my profile with a clear "waiting on parent" status, so that I'm not confused about what's been decided.
45. As a kid, I want to be notified (in-app) when a parent approves or rejects my pending request, so that I learn the outcome.

### Parent — overview / queries

46. As a parent, I want a household overview that shows all kids' total XP across configurable time windows (this week, this month, all time), so that I can see effort patterns across the family.
47. As a parent, I want per-category breakdowns of XP totals across kids and windows, so that I can see who's leaning into which skills.
48. As a parent, I want to view a single kid's activity history with filters by category and date range, so that I can answer "what has Mia been doing this week?"

### System — invariants

49. As a parent, I want activity reward values snapshotted at log time, so that adjusting an activity type's XP later doesn't rewrite past totals.
50. As a parent, I want redemption costs snapshotted at request time, so that adjusting a reward's price later doesn't rewrite past redemptions.
51. As a parent, I want history rows to never be hard-deleted, so that the household record is forensic.
52. As a parent, I want soft-deleted config entities (archived categories, rewards, etc.) to still resolve correctly in historical views, so that past data renders without "[deleted]" placeholders.
53. As a parent, I want the system to handle cascade unlocks (achievement A grants points that satisfy achievement B's rule) in the same evaluation pass, so that all earned milestones surface immediately rather than dripping in across future events.

## Implementation Decisions

### Domain entities

The persistence model comprises 10 tables across two groups.

**Config / CRUD entities (soft-deletable via `archived_at`):**

- **kids** — `id`, `name`, `avatar_slug`, `color`, `display_order`, `archived_at`, `created_at`.
- **categories** — `id`, `slug` (unique among non-archived), `name`, `description`, `icon`, `color`, `archived_at`, `created_at`.
- **activity_types** — `id`, `category_id`, `slug`, `name`, `description`, `xp_per_unit`, `points_per_unit`, `archived_at`, `created_at`.
- **achievements** — `id`, `slug`, `name`, `description`, `title` (nullable, the earned-flair string), `combinator` (`ALL` / `ANY`), `bonus_points`, `archived_at`, `created_at`.
- **achievement_rules** — `id`, `achievement_id`, `category_id` (nullable; NULL = global), `metric` (`count` / `xp` / `points` / `level`), `threshold`.
- **rewards** — `id`, `slug`, `name`, `description`, `cost_points`, `active`, `archived_at`, `created_at`.

**Event entities (append-only; admin flags allowed):**

- **activities** — `id`, `kid_id`, `activity_type_id`, `quantity`, `xp_awarded` (snapshotted total), `points_awarded` (snapshotted total), `note`, `occurred_at` (parent-editable), `created_at` (immutable), `voided_at`, `void_reason`.
- **kid_achievements** — `kid_id`, `achievement_id`, `earned_at`, `unseen` (cleared when surfaced). PK `(kid_id, achievement_id)`.
- **redemptions** — `id`, `kid_id`, `reward_id`, `points_spent` (snapshot), `status` (`pending` / `approved` / `rejected` / `cancelled`), `requested_at`, `decided_at`.
- **point_adjustments** — `id`, `kid_id`, `points_delta` (signed, non-zero), `reason`, `created_at`, `voided_at`, `void_reason`.

DB-level CHECK constraints encode invariants (`quantity > 0`, `xp_awarded >= 0`, `cost_points > 0`, `combinator IN ('ALL','ANY')`, `metric IN ('count','xp','points','level')`, etc.). Partial unique indexes scope `slug` uniqueness to non-archived rows.

### Modules

**Deep modules (each tested in isolation):**

- **AchievementEngine** — single entry point `Reevaluate(ctx, db, kidID) ([]Achievement, error)` performing fixed-point iteration: each pass evaluates all unearned achievements for a kid; loop terminates when a pass yields zero new earns. Per-rule metric computation dispatches on `metric` (count / xp / points / level), with `level` thresholds converted to XP thresholds at rule load. Category-scoped `points` rules consider only `activities.points_awarded`; global `points` rules include achievement bonuses and adjustments. Combinator-aware boolean fold (`ALL`=AND, `ANY`=OR).
- **BalanceCalculator** — `Earned`, `Spent`, `Reserved`, `Balance`, `AvailableBalance`, `PointsEarnedInCategory` queries. Pure aggregation across activities, kid_achievements (joined to achievements.bonus_points), point_adjustments, redemptions. Filters voided rows and approved redemptions correctly. Available balance subtracts pending redemption costs from balance.
- **LevelingCurve** — pure-function module: `LevelForXP(xp) int`, `XPForLevel(level) int64`, `XPForNextLevel(currentXP) int64`. Single source of leveling truth.
- **NextAchievementProgress** — `Top3UnearnedForKid`, `ClosestInCategory`. Combinator-aware progress: min(rule_ratio) for ALL achievements, max(rule_ratio) for ANY achievements. Returns rules-with-progress for UI.
- **ProfileBuilder** — `BuildProfile(ctx, kidID) ProfileView` composing kid + per-category levels + balance details + earned achievements + top-3 next + recent activity into a single read view-model.
- **AvatarWhitelist** — `Has(slug) bool`, `List() []Avatar` sourced from the `static/avatars/` directory at boot.
- **ValidationAdapter** — converts `validator.ValidationErrors` from go-playground/validator into the project's `ValidationError{Fields: map[string]string}` shape.

**Shallow modules (CRUD glue; integration-tested via service tests):**

- 10 services (one per aggregate root + the two exceptions): `KidService`, `CategoryService`, `ActivityTypeService`, `ActivityService`, `AchievementService` (wraps AchievementEngine + CRUD + Mark Seen), `RewardService`, `RedemptionService`, `PointAdjustmentService`, `BalanceService` (wraps BalanceCalculator), `ProfileService` (wraps ProfileBuilder).
- 8 repositories — one per persistence-bearing entity (kids, categories, activity_types, activities, achievements [with rules], rewards, redemptions, point_adjustments), each defined as an interface with a SQLite implementation that thin-wraps sqlc-generated `Queries` and maps rows to `domain.*` types.

### Architecture

- **Layered package structure** under `internal/`: `domain/` (pure structs), `storage/` (DBTX, tx helpers, migrations, sqlc-generated `sqldb/` private package), `repository/` (interfaces + sqlite impls), `service/` (interfaces + impls + DTOs + validation + errors), `http/` (controllers, router, middleware — future), `view/` (templates, renderer — future), `seed/` (Go-defined config upserted by `kidsboard seed`).
- **Dependency injection** via interfaces. Services depend on repository interfaces and other service interfaces, never concrete types. Wiring assembled at startup in `cmd/serve.go`.
- **Transaction threading** is explicit. A `DBTX` interface (aliased from sqlc's generated `sqldb.DBTX`) is satisfied by both `*sql.DB` and `*sql.Tx`. Repositories take `DBTX` as a parameter on every method; they do not hold a `db` field. Services that need atomicity wrap with a generic `WithTx[T]` helper that opens `BEGIN IMMEDIATE` and threads the tx through dependent service calls. The `AchievementEngine.Reevaluate` runs inside the same transaction as the mutation that triggered it.
- **Sentinel errors** for domain conditions (`ErrNotFound`, `ErrInsufficientBalance`, `ErrAchievementAlreadyEarned`, `ErrRedemptionAlreadyDecided`, `ErrRewardInactive`, `ErrInvalidInput`). Repositories translate `sql.ErrNoRows` to `ErrNotFound` at the boundary. A `ValidationError` custom type carries per-field error messages from form submissions.
- **Two-phase validation** in services: input DTOs (one per mutating service method) are tagged for `go-playground/validator` and validated first; cross-table invariants (archived FKs, balance checks, avatar whitelist membership) run as hand-rolled checks after struct validation passes. Domain types stay free of validation tags.

### Achievement engine semantics

- Re-eval is triggered synchronously, inside the same transaction, after every mutating event that can affect achievement state: activity log, activity void, point adjustment, point adjustment void, redemption cancel (which releases points).
- Earned achievements are returned from the triggering service call AND marked `unseen=true`. The HTTP layer (when built) can surface a celebration moment immediately, while the profile shows a NEW! indicator for unlocks earned outside the active session. `MarkSeen(kidID, achievementID)` clears the flag.
- Once earned, achievements stay earned. Voiding an underlying activity reduces future totals but does not cascade-revoke historical earned rows.
- Cross-achievement cascades (A grants points → B's rule passes) are handled by the fixed-point loop: B unlocks in the same Reevaluate call, in a subsequent pass.
- `metric=level` is sugar for `metric=xp` with the threshold transformed via `LevelingCurve.XPForLevel(threshold)` at rule load.

### Redemption semantics

- Redemption requests are subject to strict balance checks at both request time and approval time. A request fails if `AvailableBalance < cost`; an approval fails if `Balance < cost`.
- Pending requests *reserve* points: `AvailableBalance = Balance − sum(pending.cost)`. The kid never sees a balance they couldn't actually spend.
- Status transitions: `pending → approved`, `pending → rejected`, `approved → cancelled`. Cancellation of an approved redemption returns the points (the balance derivation filters `status='approved'`).

### Storage & tooling stack

- **modernc.org/sqlite** as the SQLite driver — pure Go, no CGO, trivial cross-compilation for ARM / Raspberry Pi deploys.
- **pressly/goose** for migrations, embedded via `//go:embed`, run on app start.
- **sqlc** for query generation. Schema is sourced from the migration files (single source of truth). Generated package is private to the storage layer; repositories map sqlc row types into `domain.*` types at the boundary.
- **go-playground/validator** for DTO validation, with custom validators registered for `avatar_slug` (whitelist check) at startup.
- **mockery** for generating service mocks consumed by service-layer unit tests.
- **cobra** for the CLI (already in place). Subcommands: `serve` (start HTTP server), `seed` (upsert Go-defined config rows by slug — though all entities are also CRUD-able from the UI per design).
- **IDs**: `INTEGER PRIMARY KEY` everywhere. Seeded/CRUD-able entities additionally carry a `slug TEXT` with a partial unique index scoped to non-archived rows.

### SQLite configuration

Boot-time pragmas: `journal_mode=WAL`, `synchronous=NORMAL`, `foreign_keys=ON`, `busy_timeout=5000`, `temp_store=MEMORY`. Write transactions use `BEGIN IMMEDIATE` (via `sql.LevelSerializable`) to acquire the write lock up front and avoid mid-transaction BUSY. Connection pool size is left at the Go default.

## Testing Decisions

### What makes a good test in this codebase

- Tests assert observable *behavior* — given inputs and a starting database state, what does the public method return and what does the database look like afterward. They do not assert that a specific repository method was invoked with specific args.
- Tests prefer real `:memory:` SQLite to mocked repositories when the test surface is a service that orchestrates database work. Mocks are reserved for cross-service dependencies (e.g. injecting a fake `AchievementEngine` into `ActivityService` to test that Log invokes Reevaluate exactly once per call, without exercising the engine itself).
- Test data is constructed via builders / table-driven cases rather than long imperative setup. Each test owns its own `:memory:` DB; nothing is shared across tests.

### Modules to test (locked)

- **AchievementEngine** — dedicated test suite covering:
  - Single-rule pass (count, xp, points-global, points-category-scoped, level).
  - ALL combinator partial (one rule passes, achievement not earned).
  - ALL combinator full pass.
  - ANY combinator (one rule of several passes).
  - Cross-category global rules (`category_id IS NULL`).
  - Cascade: achievement A grants points; achievement B's points rule passes only because A earned; both surface from a single Reevaluate call.
  - Fixed-point termination (no further unlocks after a pass; exits cleanly).
  - Idempotency: running Reevaluate twice in a row yields zero new earns on the second call.
  - Already-earned achievements not re-evaluated.
  - Voided activities excluded from metric computation.
  - Achievement stays earned after the activity that triggered it is voided.

- **BalanceCalculator** — dedicated test suite covering:
  - Earned = activities + earned-achievement bonuses + positive adjustments.
  - Spent = approved redemptions + negative adjustments (as positive).
  - Reserved = sum of pending redemption costs.
  - AvailableBalance = Balance − Reserved.
  - Cancelled redemptions return points (do not count toward Spent).
  - Voided adjustments excluded.
  - Voided activities excluded.
  - Per-category points (`PointsEarnedInCategory`) excludes achievements and adjustments.

- **LevelingCurve** — pure-function tests:
  - Boundary values (XP=0 → Level 1; arbitrary XP → expected level via formula).
  - Round-trip: `LevelForXP(XPForLevel(N)) == N` for a range of N.
  - Monotonic behavior.

- **NextAchievementProgress** — combinator-aware progress tests:
  - ALL achievement with mixed rule progress uses min.
  - ANY achievement with mixed rule progress uses max.
  - Top-3 ordering correctly ranks unearned by overall ratio.
  - Achievements already earned are excluded.
  - Category-filter variant returns the right achievement.

- **ProfileBuilder** — end-to-end composition tests:
  - Empty kid: zero XP per category, zero balance, no earned, top-3 = first 3 unearned by ratio.
  - Kid with mixed activity / earned / pending: profile fields populate correctly.
  - Profile reflects soft-archived categories correctly (still resolves names for historical earned rows).

- **Repository aggregate queries** — focused query tests against `:memory:` SQLite:
  - Activity aggregates (`CountForKid`, `SumXPForKid`, with and without category filter, with and without voided filter).
  - Achievement engine support queries (`ListUnearnedForKid`, `MarkEarned` idempotency under `UNIQUE` constraint).
  - Balance support queries (sums across each event table with the correct WHERE filters).
  - Redemption pending sum, approved sum.

### Prior art / patterns

- No prior tests exist in the repo (greenfield). The test patterns will set the precedent for the codebase.
- Test helpers (`newTestDB(t)`, builders for kids / categories / activities) live in `internal/service/testutil/` and are reused across service and engine tests.
- Cross-service mock generation via `mockery` is driven by a `Makefile` target invoking `//go:generate` directives on each service interface file. Mocks are committed under `internal/service/mocks/` for IDE navigation.

## Out of Scope

The following are explicitly NOT part of this PRD; they may follow in later iterations:

- **HTTP / templates layer.** This PRD covers data model + services (including the deep engine and balance modules) only. The HTTP controllers, html/template rendering, Tailwind styling, kid-selector UI, and parent admin pages are a follow-up PRD.
- **Authentication.** This is a household app on a trusted device; the kid-selector is the only identity affordance.
- **Multi-tenancy.** Single-household, single-instance.
- **Cloud sync / multi-device consistency.** Local SQLite, single-instance.
- **Push notifications / email.** Out of scope; in-app surfacing only.
- **Photo uploads for avatars.** Curated avatar set only; no file upload.
- **Procedural / generated avatars.** Curated set only.
- **Tiered achievements as a first-class entity.** Tiers are expressed by creating separate achievement rows (Bronze / Silver / Gold). No `achievement_tiers` table.
- **Periodic / repeatable achievements** ("every 100 dishes"). One-shot per kid only.
- **Nested boolean rule groups** (AND-of-ORs). Flat rule lists combined by a single ALL/ANY combinator per achievement.
- **Per-category leveling curves.** Single formula in code for all categories.
- **Leaderboard as a primary UI surface.** Profile is the primary surface; the leaderboard *queries* (XP totals by window, by category) are supported in the data model and exposed via service methods for a future overview UI.
- **Background workers / schedulers.** Re-eval is synchronous in-tx; no goroutine lifecycle.
- **Achievement revocation cascade.** Voiding an activity does not retroactively revoke any kid_achievement.
- **Approval flow for activity logging.** Parent-only, no kid-side approval queue for activities.
- **Audit log of who-did-what.** Logger identity is not tracked (single trusted device).
- **Cross-kid achievements** ("siblings combined wash 1000 dishes"). All metrics are scoped to a single kid.

## Further Notes

- The user is building this for their own four kids; the design biases toward "single-binary, easy deploy to a small home device" over "enterprise-scale ergonomics." This influenced choices like no CGO (modernc), synchronous re-eval (no worker), no auth, and embedded migrations.
- The `RPG character sheet` aesthetic is the differentiating design constraint. Curated avatars, per-category levels (skill tree), achievement *titles* shown on profile, named milestones with humorous flavor text — all support this. Resist sticker-chart UI patterns when the UI PRD lands.
- The fixed-point re-evaluation algorithm is the load-bearing piece of business logic. It deserves the most testing investment.
- Configuration entities (categories, activity_types, achievements, rewards) are CRUD-able from the UI per the agreed design. Seed code exists for initial population but is not the only path to creation. This means the `kidsboard seed` command is for bootstrap and re-bootstrap only — parents can also create entities via the admin UI once it lands.
- A future PRD should consider whether to add a dedicated "household activity feed" view aggregating recent activities, achievement unlocks, and redemptions across all kids — useful for parents reviewing the day. Not required for the data layer.
