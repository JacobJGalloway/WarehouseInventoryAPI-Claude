# Switchyard — Architecture v1.4

> **Intended audience:** Claude Code, sprint planning sessions, project handoff.
> **Status:** Pre-sprint architectural design. Approved scope, pending sprint start.
> **Parent system:** Switchyard (.NET Core 10 / React / PostgreSQL / Auth0)
> **Sprint model:** Hardening sprint. Scope is pilot-client ready — functional completeness, accessibility, and infrastructure restoration.
> **Sprint goal:** Empty Return flow, Delivered column redesign, Deadhead pairing enforcement, Auth0 hardening, WCAG AA + ARIA remediation, CQRS read replica restoration, and design system border documentation.
> **Prerequisite:** v1.3 merged to `main` and demo seed verified functional before this sprint begins.

---

## Table of Contents

1. [Sprint Goal & Definition of Done](#1-sprint-goal--definition-of-done)
2. [v1.4 Scope — In Sprint](#2-v14-scope--in-sprint)
3. [Feature: Empty Return Board State](#3-feature-empty-return-board-state)
4. [Feature: Delivered Column Redesign](#4-feature-delivered-column-redesign)
5. [Feature: Deadhead Pairing Enforcement](#5-feature-deadhead-pairing-enforcement)
6. [Feature: Rolling Refresh Tokens — Auth0](#6-feature-rolling-refresh-tokens--auth0)
7. [Feature: Color Contrast Audit (WCAG AA)](#7-feature-color-contrast-audit-wcag-aa)
8. [Feature: ARIA Compliance Audit](#8-feature-aria-compliance-audit)
9. [Feature: Card Border Language — Design System](#9-feature-card-border-language--design-system)
10. [Feature: CQRS Read Replica Restoration](#10-feature-cqrs-read-replica-restoration)
11. [Human-in-the-Loop Checkpoints](#11-human-in-the-loop-checkpoints)
12. [Out of Scope — v1.4](#12-out-of-scope--v14)

---

## 1. Sprint Goal & Definition of Done

v1.4 is a **hardening sprint** with two categories of work: dispatch board functional completeness (Empty Return, Delivered redesign, Deadhead pairing) and infrastructure/quality hardening (Auth0, WCAG AA, ARIA, CQRS, design system). Order within the sprint is flexible relative to v1.3 completion.

A v1.4 release to `main` is considered **complete** when:

- [ ] Empty Return sub-section is present on the board; drivers on empty return appear with ETA visible
- [ ] Delivered column operates as a dispatch review/close-out state; driver and equipment decouple from BOL at last stop confirmation
- [ ] Deadhead pairing enforces `DEADHEAD_CUTOFF_MINUTES` window; missed window routes driver to Empty Return
- [ ] Auth0 sessions use rolling refresh tokens; fixed-expiry client secrets removed
- [ ] WCAG AA color contrast passes in both light and dark themes
- [ ] ARIA audit complete — board columns, cards, icon-only buttons, skip-nav
- [ ] Card border semantic language documented in the Claude Design System
- [ ] CQRS read replica restored — separate read database in place, sync services running, dev/prod configs updated
- [ ] No regressions introduced to v1.3 functionality
- [ ] README startup sequence updated — Docker containers must start before the APIs; correct order documented in main README

---

## 2. v1.4 Scope — In Sprint

| # | Feature | Weight | Notes |
|---|---------|--------|-------|
| 1 | Empty Return board state | Medium | New board sub-section; ETA display; driver routing logic |
| 2 | Delivered column redesign | Medium | Driver/equipment decoupling at last stop; close-out card |
| 3 | Deadhead pairing enforcement | Medium | Window enforcement at board level; missed-window fallback to Empty Return |
| 4 | Rolling refresh tokens (Auth0) | Light | Config + token strategy only; no schema changes |
| 5 | Color contrast audit (WCAG AA) | Medium | Light and dark themes |
| 6 | ARIA compliance audit | Medium | Board columns, cards, icon-only buttons, skip-nav |
| 7 | Card border language — design system | Light | Documentation addition to existing Claude Design System |
| 8 | CQRS read replica restoration | Medium | Restore read/write split; wire sync services; update connection config |

> Items 1–3 are logically coupled (they describe the full post-BOL driver flow) and should be understood together before implementation begins, but can be built incrementally. Items 4–8 are independent and can be completed in any order.

---

## 3. Feature: Empty Return Board State

### Context
When a driver completes a BOL run and no deadhead pairing is secured within the `DEADHEAD_CUTOFF_MINUTES` window, the driver has no current assignment. They are physically en route back to the originating warehouse but are not yet available for a new BOL. The board currently has no representation for this state, which leaves dispatch with a visibility gap during pre-planning.

### Board design
Add an **Available** sub-section within the dispatch board for drivers on empty return:

- Driver card appears in the Available sub-section on last stop confirmation when no deadhead pairing is active
- Card displays driver name, originating warehouse destination, and **ETA** to arrival
- ETA is calculated from last stop confirmation timestamp + estimated return transit time (see Human checkpoint — ETA source)
- Card is visually distinct from assigned/active driver cards — no BOL reference, no load details
- On arrival confirmation (or estimated arrival), driver transitions to a fully available state for new BOL assignment

### Data considerations
No new domain entities required. This is a **board state and routing logic** change:
- Driver status requires a new state value (e.g., `EmptyReturn`) in the driver status enum
- ETA field can be a calculated value stored on the driver assignment record at last stop confirmation, not a persisted live-tracking value
- The Available sub-section reads from existing driver/assignment data filtered by the new status

### Definition of done for this feature
- [ ] `EmptyReturn` status (or equivalent) added to driver status enum
- [ ] Driver routes to Empty Return sub-section on last stop confirmation when no deadhead pairing is active
- [ ] ETA is visible on the Empty Return card
- [ ] Driver transitions out of Empty Return on arrival or new BOL assignment
- [ ] Demo seed includes at least one driver in Empty Return state for demo visibility

---

## 4. Feature: Delivered Column Redesign

### Current behavior
The Delivered column currently holds completed BOL cards with driver and equipment still associated. Driver and equipment remain coupled to the BOL after delivery, which blocks them from routing to their next state (Empty Return, Available, or Maintenance).

### Target behavior
Delivered represents the **dispatch review and close-out window** — the period between last stop confirmation and final archiving. During this window:

- The BOL card in Delivered becomes a **close-out card** — BOL details only; no driver or equipment reference
- Driver decouples from the BOL at last stop confirmation and routes independently:
  - Deadhead pairing secured → routes to deadhead run
  - Deadhead window missed → routes to Empty Return
  - Flagged for maintenance → routes to Maintenance queue
- Equipment decouples from the BOL at last stop confirmation and routes independently to its next state (available pool or maintenance)
- The close-out card remains in Delivered until dispatch completes review, client notification, and final paperwork, then archives

### Decoupling trigger
Last stop confirmation is the event that triggers decoupling. At that moment:
1. BOL status transitions to `Delivered` (pending close-out)
2. Driver assignment to BOL is closed; driver status transitions per deadhead pairing state
3. Equipment assignment to BOL is closed; equipment status transitions to available or maintenance

### Definition of done for this feature
- [ ] BOL close-out card in Delivered shows BOL details only — no driver or equipment
- [ ] Driver decouples from BOL at last stop confirmation
- [ ] Equipment decouples from BOL at last stop confirmation
- [ ] Driver routing post-decoupling follows deadhead pairing state (see Section 5)
- [ ] Close-out card archives on dispatch review completion
- [ ] Demo seed includes at least one BOL in Delivered close-out state

---

## 5. Feature: Deadhead Pairing Enforcement

### Context
A deadhead run pairs a driver completing a delivery with a return load, eliminating an empty return trip. The pairing must be secured before the driver reaches the last stop — if the window closes without a pairing, the contract is voided and the driver routes to Empty Return instead.

This feature was a known planned addition. The Empty Return board state (Section 3) is the fallback that makes this feature complete — without it, a missed-window driver has nowhere to land on the board.

### Enforcement logic
`DEADHEAD_CUTOFF_MINUTES` defines the window before last stop confirmation during which a deadhead pairing must be secured. Enforcement happens at the **board level**:

- When a driver enters the cutoff window (ETA to last stop ≤ `DEADHEAD_CUTOFF_MINUTES`), the board surfaces a pairing alert on the active BOL card
- If pairing is secured before last stop confirmation → driver routes to deadhead run on confirmation
- If last stop is confirmed without an active pairing → contract is voided; driver routes to Empty Return

### Configuration
`DEADHEAD_CUTOFF_MINUTES` should be an environment-configurable value, not a hardcoded constant. Confirm the current location of this value in the codebase before implementation — if it already exists as a constant, promote it to config; if it does not exist, introduce it as a configurable setting.

### Board state transitions summary

```
Last stop confirmed
├── Deadhead pairing active → Driver routes to deadhead run
│                             Equipment decouples → available/maintenance
│                             BOL → Delivered (close-out card)
└── No pairing (window missed) → Driver routes to Empty Return
                                  Equipment decouples → available/maintenance
                                  BOL → Delivered (close-out card)
```

### Human checkpoint
> Confirm with Jacob: should the board surface a visual warning when a driver enters the `DEADHEAD_CUTOFF_MINUTES` window without an active pairing? This is a dispatch UX decision — helpful for demo but needs product sign-off on placement and urgency level.

### Definition of done for this feature
- [ ] `DEADHEAD_CUTOFF_MINUTES` is environment-configurable
- [ ] Board surfaces pairing alert when driver enters cutoff window without active pairing
- [ ] Last stop confirmation with active pairing routes driver to deadhead run
- [ ] Last stop confirmation without pairing voids contract and routes driver to Empty Return
- [ ] BOL transitions to Delivered close-out card in both scenarios
- [ ] Demo seed includes at least one BOL at or near the cutoff window for demo visibility

---

## 6. Feature: Rolling Refresh Tokens — Auth0

### Problem
The current Auth0 session strategy uses fixed-expiry client secrets. When a secret expires mid-demo or mid-session, the user is logged out without warning. Rolling refresh tokens eliminate this by continuously extending the session as long as the user is active.

### Change scope
Configuration and token strategy change only. No schema changes. No new API endpoints. No frontend state changes beyond what Auth0's SDK handles automatically.

### Implementation

**Auth0 dashboard**
- Enable **Refresh Token Rotation** for the application
- Set an appropriate **Absolute Expiration** (e.g., 30 days) and **Inactivity Expiration** (e.g., 24 hours for demo sessions)
- Disable reuse detection if it causes friction during demo sessions on shared machines (optional — see Human checkpoint)

**Application configuration**
- Confirm `offline_access` scope is requested on login to receive a refresh token
- Configure Auth0 SDK (React side) with `useRefreshTokens: true`
- Remove any fixed-expiry client secret logic currently in place

### Human checkpoint
> Confirm with Jacob: prefer `cacheLocation: 'memory'` (safer — session lost on page refresh) or `cacheLocation: 'localstorage'` (session persists across refreshes — more demo-friendly but less secure)? Product call, not a code call.

### Definition of done for this feature
- [ ] Rolling refresh tokens enabled in Auth0 dashboard
- [ ] `offline_access` scope confirmed in login request
- [ ] SDK configured with `useRefreshTokens: true`
- [ ] Fixed-expiry client secret logic removed
- [ ] Active session renews without logout; expired inactive session redirects to login cleanly

---

## 7. Feature: Color Contrast Audit (WCAG AA)

### Scope
Verify and correct all text/background color combinations in both light and dark themes against WCAG AA minimums: **4.5:1** for normal text, **3:1** for large text and UI components.

### Areas of concern
- Dispatch card text against card background (including border-state variants — danger/warn/ok)
- Status badge text against badge backgrounds
- Column header text in both themes
- Navigation and sidebar elements
- Empty Return card text (new in v1.4 — audit at build time, not after)

### Approach
Fix at the CSS variable / design token level, not inline, so corrections propagate across both themes. After each variable change, cross-check the other theme — shared tokens affect both simultaneously.

### Definition of done for this feature
- [ ] All text/bg combinations in light theme meet WCAG AA
- [ ] All text/bg combinations in dark theme meet WCAG AA
- [ ] No new contrast failures introduced across themes
- [ ] Empty Return card variants verified at build time

---

## 8. Feature: ARIA Compliance Audit

### Scope
Audit and remediate accessibility markup across the dispatch board UI. No visual redesign — semantic and ARIA attribute corrections only.

### Target areas

**Board columns**
- Each column must have `role` and `aria-label` describing its dispatch state
- Column headers must be associated with their content regions

**Dispatch cards**
- Cards must be keyboard-focusable and announce load/driver/status context to screen readers
- Status indicators must not rely on color alone — text or icon label required alongside color coding
- Empty Return cards (new in v1.4) must be built to spec from the start

**Icon-only buttons**
- Every icon-only button must have `aria-label` with a meaningful action description
- `matTooltip` or equivalent is not a substitute

**Skip-nav**
- "Skip to main content" link must be present, first in DOM, and visible on keyboard focus

### Definition of done for this feature
- [ ] All board columns have `role` and `aria-label`
- [ ] All dispatch and close-out cards are keyboard-accessible and announce meaningful content
- [ ] All icon-only buttons have `aria-label`
- [ ] Skip-nav present, keyboard-visible, and functional
- [ ] Manual keyboard-navigation walkthrough of the board passes without dead ends

---

## 9. Feature: Card Border Language — Design System

### Context
Dispatch board card borders carry semantic meaning that is not currently documented. This creates a risk that future development introduces decorative borders that conflict with the established semantic system, or removes meaningful borders thinking they are decorative.

### Border semantic system
| Border state | Meaning |
|-------------|---------|
| Border present (danger variant) | Status alert — requires immediate attention |
| Border present (warn variant) | Status alert — attention needed, not critical |
| Border present (ok variant) | Status alert — acknowledged or in progress |
| No border | Clean status — no alerts, no action required |

Borders carry semantic weight. They must not be used decoratively. A card with no actionable status must have no border.

### Implementation
Add a **Card Border Language** section to the existing Claude Design System document. The section must cover:

- The semantic rule: border present = status alert; no border = clean status
- The three border variants (danger, warn, ok) with their CSS variable names and intended use cases
- An explicit prohibition on decorative border use
- A note that Empty Return cards (new in v1.4) follow the same system — an Empty Return driver with no issues carries no border

Do not create a new document. This addition belongs in the existing Claude Design System.

### Definition of done for this feature
- [ ] Card Border Language section added to Claude Design System
- [ ] All three border variants documented with CSS variable names and use cases
- [ ] Decorative border prohibition stated explicitly
- [ ] Empty Return card border behavior documented

---

## 10. Feature: CQRS Read Replica Restoration

### Context
Switchyard was originally built with read/write separation (CQRS pattern) using SQLite, where separate read and write databases were natural. When the system migrated to PostgreSQL, read and write connections were collapsed to point at the same database as a temporary stand-in. That stand-in has persisted. v1.4 restores the intended architecture: a separate PostgreSQL read database with sync services keeping it current.

This is a **restoration**, not a new design. The pattern, sync services, and connection configuration all existed before. The work is reconnecting what was disconnected during the PostgreSQL migration.

### Target architecture
- **Write database** (`switchyard_write` or equivalent): receives all CUD operations (Commands)
- **Read database** (`switchyard_read` or equivalent): serves all read operations (Queries); kept in sync via sync services
- **Sync services**: responsible for propagating write database changes to the read database; these services exist or existed — locate and restore before writing new code

### Implementation approach

1. **Locate sync services** in the codebase before writing anything. If they are present but disabled/misconfigured, restore configuration. If they have been removed, surface to Jacob before rebuilding.
2. **Provision read database**: create a separate PostgreSQL database for read operations. Connection strings for both databases go in environment config (not hardcoded).
3. **Update connection configuration**: read operations point to read database connection string; CUD operations point to write database connection string. Verify this separation holds across all repositories.
4. **Restore sync services**: ensure sync services are running and propagating changes correctly. A basic smoke test (write a record, confirm it appears on the read side within expected latency) is sufficient for v1.4.
5. **Update dev config**: dev environment currently points read connections at the write database — update to match the restored architecture.

### Human checkpoint
> Before provisioning the read database or touching sync services, surface the current state of the sync services to Jacob: found and intact / found but misconfigured / not found. Do not rebuild from scratch without confirmation.

### Definition of done for this feature
- [ ] Sync services located and confirmed running (or restored with Jacob's input)
- [ ] Separate read PostgreSQL database provisioned
- [ ] Read connection strings point to read database in all environments
- [ ] Write connection strings point to write database in all environments
- [ ] Sync services propagating changes from write to read correctly
- [ ] Smoke test passes: write operation appears on read side within expected latency
- [ ] Dev environment config updated to reflect restored architecture

---

## 11. Human-in-the-Loop Checkpoints

Claude Code must pause and surface a question to Jacob before proceeding at these points:

| # | Checkpoint | Question |
|---|-----------|----------|
| 1 | Empty Return ETA source | How should ETA be calculated — estimated transit time constant, distance-based estimate, or something else? |
| 2 | Deadhead cutoff window alert | Should the board surface a visual warning when a driver enters the cutoff window without an active pairing? Placement and urgency level need product sign-off. |
| 3 | Auth0 cache location | Prefer `memory` (safer, loses session on page refresh) or `localstorage` (persists session, more demo-friendly)? |
| 4 | Sync service status | Surface current state of sync services (intact / misconfigured / not found) before any rebuild work begins. |
| 5 | Contrast fix on shared tokens | If a CSS variable change affects both themes simultaneously, surface before committing. |

Do not guess on these. Surface the question, wait for a response, then proceed.

---

## 12. Out of Scope — v1.4

- Read replica health endpoint (Backlog)
- Extract User Management to dedicated identity service (Backlog)
- Scalar branding (Backlog — blocked by Scalar's limited logo support in the .NET package)
- Any features listed as "Backlog" in the README
- Yearly Yields or any other project
