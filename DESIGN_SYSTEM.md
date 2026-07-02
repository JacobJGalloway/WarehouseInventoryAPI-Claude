# Switchyard — Design System Handoff for Claude Code

> **What this is:** A complete visual and component specification for the Switchyard warehouse management platform. Use this document to implement or extend any surface of the app (Switchyard.UI React app, Switchyard-Go Dispatch Whiteboard, or future surfaces) with consistent style, tokens, and patterns.

---

## Architecture Overview

Switchyard is a white-label warehouse management platform delivered as:

- **Switchyard.UI** — React + TypeScript + Vite, port 5173. Screens: Login, Dashboard, Inventory, Bills of Lading, Users. CSS Modules + CSS custom properties.
- **Switchyard-Go** — Go server, port 8080. Dispatch Whiteboard (cross-warehouse kanban). Server-rendered templates + static assets.
- **Switchyard.InventoryAPI** — .NET 10, port 7000.
- **Switchyard.LogisticsAPI** — .NET 10, port 7001.

Auth is Auth0. Roles: Employee (read-only) / Manager (BOL create + modify) / Admin (+ user management). Roles are read from the `https://warehouselogistics/roles` claim.

The pilot client is **Digital Parts Logistics Company**. The platform is multi-tenant; tenants override four CSS variables (see Theming section) to re-skin the chrome.

---

## Tokens CSS File

Copy `colors_and_type.css` from the design system into the codebase root and import it in `index.css` or the app entry point. It defines all `--dp-*` and `--sy-*` custom properties. **Never hardcode hex values that have a token — always use the token.**

Self-hosted fonts required (SIL OFL 1.1):
- Inter — 400 / 500 / 600 / 700 (UI, display, body)
- JetBrains Mono — 400 / 500 / 600 / 700 (code, SKUs, counters)
- Dune Rise — logo use only (tenant override wordmarks)
- Omicron — master Switchyard wordmark face (logo use only)

---

## Color Palette

### Brand
| Token | Hex | Use |
|---|---|---|
| `--dp-blue` | `#3a73c1` | Primary — headings, active nav, CTAs, links |
| `--dp-blue-hover` | `#2f63aa` | Primary hover state |
| `--dp-blue-press` | `#2a579a` | Primary pressed state |
| `--dp-blue-tint` | `rgba(58,115,193,0.10)` | Tinted bg (info badges, hover fills) |
| `--dp-blue-ring` | `rgba(58,115,193,0.40)` | Focus ring |
| `--dp-graphite` | `#3b3a3c` | Dark-mode bg, nav bg, logo negative space |
| `--dp-offwhite` | `#eeeeee` | Light-mode page bg and card surface |
| `--dp-accent` | `#aa3bff` | Focus rings, hover tints — use very sparingly |
| `--dp-accent-bg` | `rgba(170,59,255,0.10)` | Accent hover tint |
| `--dp-accent-ring` | `rgba(170,59,255,0.50)` | Accent focus ring |

### Neutrals (light mode)
| Token | Hex | Use |
|---|---|---|
| `--dp-n-0` | `#ffffff` | Input / textarea fills only |
| `--dp-n-50` | `#f7f7f8` | — |
| `--dp-n-100` | `#eeeeee` | Page bg, cards |
| `--dp-n-200` | `#e5e4e7` | Default borders |
| `--dp-n-300` | `#d2d1d5` | — |
| `--dp-n-400` | `#9ca3af` | Muted fg |
| `--dp-n-500` | `#6b6375` | Body text |
| `--dp-n-600` | `#55525c` | — |
| `--dp-n-700` | `#3b3a3c` | Nav text, strong borders |
| `--dp-n-900` | `#08060d` | Headings, table text, the nav-rule black |

### Semantic surfaces
| Token | Light value | Dark value | Use |
|---|---|---|---|
| `--dp-bg` | `#eeeeee` | `#363A3C` | Page background |
| `--dp-surface` | `#eeeeee` | `#363A3C` | Cards, modals |
| `--dp-surface-2` | `#eeeeee` | `#484A4C` | Elevated cards (Dispatch board) |
| `--dp-canvas-deep` | `#eeeeee` | `#363A3C` | Tool canvas (Dispatch board page) |
| `--dp-input-bg` | `#ffffff` | — | Input / select / textarea fill |
| `--dp-border` | `#e5e4e7` | `#505254` | Default hairline border |
| `--dp-border-strong` | `#3b3a3c` | `#eeeeee` | Strong borders |

### Semantic text
| Token | Value | Use |
|---|---|---|
| `--dp-fg` | `#6b6375` (light) | Body text |
| `--dp-fg-strong` | `#08060d` (light) | Headings, table text |
| `--dp-fg-muted` | `#9ca3af` (light) | Placeholder, secondary labels |
| `--dp-fg-link` | `#3a73c1` | Links |
| `--dp-fg-on-brand` | `#ffffff` | Text on brand-blue surfaces |

### Status (locked across themes)
| Token | Hex | Use |
|---|---|---|
| `--dp-success` | `#1f9d55` | Success state |
| `--dp-success-bg` | `rgba(31,157,85,0.10)` | — |
| `--dp-warning` | `#d97706` | Warning state |
| `--dp-warning-bg` | `rgba(217,119,6,0.12)` | — |
| `--dp-danger` | `#c23b3b` | Error / danger state |
| `--dp-danger-bg` | `rgba(194,59,59,0.10)` | — |
| `--dp-info` | `var(--dp-blue)` | Informational |
| `--dp-info-bg` | `var(--dp-blue-tint)` | — |

### Category tints (Dispatch Whiteboard)
Four tinted families for BOL / Driver / Tractor / Trailer cards. Each has `-bg`, `-border`, `-rule` (3px top accent), `-ink`.
- **Blue** → BOL: bg `#e7eef9`, border `#b9cde9`, rule `#3a73c1`
- **Sand** → Driver: bg `#f0e8d9`, border `#d8c9a4`, rule `#8a6a2c`
- **Graphite** → Tractor: bg `#e7e6e8`, border `#c4c2c8`, rule `#3b3a3c`
- **Sage** → Trailer: bg `#dde9df`, border `#b6cdba`, rule `#2f6b3d`

Dark-mode variants defined in `colors_and_type.css` under `[data-theme="dark"]`.

### Dispatch chip palette (dark mode only)
| Family | bg | fg |
|---|---|---|
| BOL | `#1b2f4e` | `#7eb8f5` |
| Driver | `#2e2315` | `#c9a870` |
| Equipment (tractor + trailer) | `#222718` | `#9aac60` |
| Trailer | `#162318` | `#5a9e6e` |

Card border = chip fg at 45% opacity. Status pills: dot solid, bg at 12%, border at 45%.

---

## Typography

### Font stacks
| Token | Value | Use |
|---|---|---|
| `--dp-font-sans` | `'Inter', system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif` | All UI text, body, display |
| `--dp-font-mono` | `'JetBrains Mono', ui-monospace, Consolas, monospace` | Code, SKUs, counters, IDs |
| `--dp-font-logo` | `'Dune Rise', 'Inter', system-ui, sans-serif` | Tenant wordmarks only |
| `--dp-font-wordmark` | `'Omicron', 'Dune Rise', 'Inter', system-ui, sans-serif` | Master Switchyard wordmark |

> **Dune Rise and Omicron are logo-only.** Never use them for UI text — they lack full punctuation coverage.

### Type scale
| Token | px | Use |
|---|---|---|
| `--dp-text-xs` | 12 | Labels, captions |
| `--dp-text-sm` | 14 | Small UI, secondary text |
| `--dp-text-base` | 16 | Default |
| `--dp-text-md` | 18 | Body text (root font-size in codebase) |
| `--dp-text-lg` | 20 | — |
| `--dp-text-xl` | 24 | H2 |
| `--dp-text-2xl` | 32 | — |
| `--dp-text-3xl` | 40 | — |
| `--dp-text-4xl` | 56 | H1 |

### Leading & tracking
| Token | Value |
|---|---|
| `--dp-leading-tight` | 118% |
| `--dp-leading-normal` | 145% |
| `--dp-leading-loose` | 160% |
| `--dp-tracking-tight` | -0.03em (H1) |
| `--dp-tracking-normal` | 0.18px (body) |
| `--dp-tracking-wide` | 0.12em (wordmark) |
| `--dp-tracking-display` | 0.08em |

### Semantic type classes (defined in colors_and_type.css)
| Class | Description |
|---|---|
| `.dp-h1` | Inter 500, 56px, -1.68px tracking, brand blue |
| `.dp-h2` | Inter 500, 24px, -0.01em, brand blue |
| `.dp-h3` | Inter 600, 18px, fg-strong |
| `.dp-eyebrow` | Inter 600, 12px, 0.12em tracking, uppercase, graphite |
| `.dp-p` | Inter 400, 18px, 145% leading, fg |
| `.dp-small` | 14px, fg-muted |
| `.dp-code / .dp-counter` | JetBrains Mono, 15px, code-bg, radius-xs |
| `.dp-logo` | Dune Rise 500, 0.08em tracking, uppercase, brand blue |
| `.dp-display` | Inter 500, 0.06em tracking, uppercase, brand blue |
| `.dp-prose / .dp-body` | Full body reset: Inter, 18px, 145%, 0.18px tracking, antialiased |

---

## Spacing

4px base grid.

| Token | px |
|---|---|
| `--dp-space-0` | 0 |
| `--dp-space-1` | 4 |
| `--dp-space-2` | 8 |
| `--dp-space-3` | 12 |
| `--dp-space-4` | 16 |
| `--dp-space-5` | 20 |
| `--dp-space-6` | 24 |
| `--dp-space-8` | 32 |
| `--dp-space-10` | 40 |
| `--dp-space-12` | 48 |
| `--dp-space-16` | 64 |
| `--dp-space-20` | 80 |

---

## Radii

| Token | px | Use |
|---|---|---|
| `--dp-radius-xs` | 4 | Inputs, code, inline pills |
| `--dp-radius-sm` | 6 | Buttons, dropdowns |
| `--dp-radius-md` | 8 | Modals |
| `--dp-radius-lg` | 12 | Cards (optional) |
| `--dp-radius-pill` | 999 | Badge / pill |

---

## Shadows

| Token | Value |
|---|---|
| `--dp-shadow-sm` | `rgba(0,0,0,0.06) 0 1px 2px 0` |
| `--dp-shadow-md` | `rgba(0,0,0,0.10) 0 10px 15px -3px, rgba(0,0,0,0.05) 0 4px 6px -2px` |
| `--dp-shadow-lg` | `rgba(0,0,0,0.12) 0 20px 30px -10px, rgba(0,0,0,0.06) 0 8px 12px -6px` |
| `--dp-shadow-inset` | `inset 0 0 0 1px rgba(0,0,0,0.05)` |

Dark mode bumps opacity to 0.40 / 0.25 on `--dp-shadow-md`.

---

## Motion

| Token | Value |
|---|---|
| `--dp-ease-out` | `cubic-bezier(0.22, 1, 0.36, 1)` |
| `--dp-ease-in-out` | `cubic-bezier(0.65, 0, 0.35, 1)` |
| `--dp-dur-fast` | 120ms |
| `--dp-dur-base` | 180ms |
| `--dp-dur-slow` | 320ms |

**Only animate:** `box-shadow`, `border-color`, `color`, `background`, `opacity`. No transforms, no slides-from-off-screen, no bounces.

---

## Layout

| Token | Value |
|---|---|
| `--dp-container` | 1126px |
| `--dp-container-narrow` | 768px |

The content column is **1126px wide, centered, with 1px side borders** (`border-left: 1px solid #e5e4e7; border-right: 1px solid #e5e4e7`). Background behind it is bare body bg (no decoration).

Tables are always full-width within their container.

---

## Components

### Header (Shell)

```css
header {
  display: flex;
  align-items: center;
  padding: 3px 16px;           /* 3px vertical — verified from Header.module.css */
  background: var(--dp-bg);
  box-sizing: border-box;
}
```

- Left: brand logo (icon + wordmark, or full PNG lockup at 220px wide)
- Right: profile button — 28px Lucide `UserCircle` icon, 2px transparent border that turns `#aa3bff` on open / `rgba(170,59,255,0.5)` on hover, `border-radius: 50%`, `transition: border-color 120ms`
- Profile dropdown: `background: #fff` (light) / `#2a292b` (dark), `border: 1px solid #e5e4e7`, `border-radius: 6px`, `--dp-shadow-md`, `min-width: 180px`

### Nav strip

```css
nav {
  display: flex;
  align-items: stretch;
  padding: 0 16px;
  background: var(--dp-bg);
}
```

Nav items: `padding: 10px 14px`, `font-size: 15px`. Active item: `color: #3a73c1; font-weight: 500`. Rest: `color: #3b3a3c` (light) / `#eeeeee` (dark), `font-weight: 400`. No underline, no bg change on active.

**Signature rule:** immediately after the nav strip, a `<hr style="margin:0;border:none;border-top:2px solid #08060d">` — this 2px black rule is a visual signature of the system. Never remove it.

Nav order: `Dashboard · Inventory · Bills of Lading · Users · Whiteboard`. Whiteboard links to the Go service (external, opens new tab, shows ↗ glyph at 60% opacity).

### Button

Four variants, two sizes:

```
primary:   bg #3a73c1, color #fff, border 1px solid #3a73c1
           hover: bg #2f63aa, border #2f63aa

secondary: bg #fff, color #08060d, border 1px solid #e5e4e7
           hover: border-color #3a73c1, color #3a73c1

ghost:     bg transparent, color #3a73c1, border 1px solid transparent
           hover: border rgba(58,115,193,0.4), bg rgba(58,115,193,0.06)

danger:    bg #fff, color #c23b3b, border 1px solid #c23b3b
           hover: bg rgba(194,59,59,0.08)

disabled:  opacity 0.45, cursor not-allowed
```

Sizing:
- sm: `font-size: 12px; padding: 3px 10px`
- md: `font-size: 14px; padding: 7px 14px`

Common: `border-radius: 6px; font-weight: 500; transition: background 120ms, border-color 120ms, color 120ms`

### Field + Input + Select

```css
/* Field wrapper */
label { display: flex; flex-direction: column; gap: 4px; }
/* Field label */
span  { font-size: 13px; font-weight: 600; color: #08060d; }
/* Input / Select */
input, select {
  padding: 7px 10px;
  border: 1px solid #e5e4e7;
  border-radius: 4px;           /* --dp-radius-xs */
  background: #fff;             /* --dp-input-bg */
  color: #08060d;
  font-size: 14px;
  font-family: inherit;
}
```

Focus: `outline: 2px solid var(--dp-accent); outline-offset: 2px`

Dropdown placeholder text: `-- Select --`

### Badge / Pill

```css
span {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  font-weight: 500;
  padding: 3px 10px;
  border-radius: 999px;         /* --dp-radius-pill */
}
```

Tones (dot · text · bg · border):
```
info:    #3a73c1  · rgba(58,115,193,0.40)  · rgba(58,115,193,0.10)
ok:      #1f9d55  · rgba(31,157,85,0.40)   · rgba(31,157,85,0.10)
warn:    #d97706  · rgba(217,119,6,0.40)   · rgba(217,119,6,0.12)
danger:  #c23b3b  · rgba(194,59,59,0.40)   · rgba(194,59,59,0.10)
neutral: #3b3a3c  · #e5e4e7               · #fff
```

The dot is a 6×6px circle with the tone color as `background`.

Status mapping helper:

```js
function statusToTone(status) {
  const v = (status || '').toLowerCase();
  if (v.includes('process'))                        return 'ok';
  if (v.includes('transit') || v.includes('ready')) return 'info';
  if (v.includes('project') || v.includes('partial'))return 'warn';
  if (v.includes('cancel') || v.includes('deactiv'))return 'danger';
  return 'neutral';
}
```

### Modal

```css
/* Overlay */
div {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.5);    /* no blur */
  display: flex; align-items: center; justify-content: center;
  z-index: 200;
}
/* Sheet */
div {
  background: #eeeeee; color: #08060d;
  border: 1px solid #e5e4e7;
  border-radius: 8px;              /* --dp-radius-md */
  padding: 20px 24px;
  min-width: 560px; max-width: 90vw; max-height: 80vh;
  overflow-y: auto;
  box-shadow: var(--dp-shadow-lg);
}
```

- Title: H2 style — `font-size: 22px; color: #3a73c1; font-weight: 500; margin: 0 0 16px`
- Close button: absolute top-right `(12px, 12px)`, bare ✕, no border, `color: #08060d`
- Footer: `display: flex; gap: 8px; justify-content: flex-end; margin-top: 18px`
- Modal surface uses the hardest-black text (`#08060d`) for maximum contrast — heavier than body text elsewhere.

### Table (DataTable)

```css
table {
  border-collapse: collapse;
  width: 100%;
  margin-top: 12px;
  font-size: 14px;
}
th {
  border: 1px solid #3b3a3c;   /* graphite grid — deliberate, not hairline */
  padding: 6px 12px;
  background: #eeeeee;
  color: #08060d;
  font-weight: 600;
  text-align: left;
}
td {
  border: 1px solid #3b3a3c;
  padding: 6px 12px;
  background: #eeeeee;
  color: #08060d;
}
```

Deactivated / muted rows: `opacity: 0.45` on the `<tr>`.

Empty cells: em-dash `—` (matches codebase `{value || '—'}`).

---

## Icons

Library: **Lucide** (`lucide-react` in the codebase). CDN: `https://unpkg.com/lucide@latest` or `lucide-react` npm package.

- Default size: **24px**, stroke width **2**
- Header profile button: 28px (kit prototype uses 28; codebase spec is 32)
- Color: `currentColor` — inherits from parent
- States: `--dp-fg` at rest → `--dp-fg-strong` on hover → `--dp-blue` when active
- Do not mix libraries. Do not use emoji as icons.

Brand icon (`assets/switchyard-icon-{light,dark}.svg`) is a logo asset only — header, favicon, splash. Never use it inline in buttons or body content.

---

## Logo & Brand Assets

Files in `assets/`:
- `switchyard-logo-full-{light,dark}.png` — Full lockup (icon + wordmark + tagline). **Use PNG** — the SVG source embeds Omicron as `<text>` which requires the font installed locally.
- `switchyard-wordmark-{light,dark}.svg` — "SWITCHYARD" wordmark only
- `switchyard-icon-{light,dark}.svg` — Brand mark only
- `switchyard-powered-by-{light,dark}.svg` — "Powered by Switchyard" badge (tenant surfaces)

Use light-mode assets on `#eeeeee` / light surfaces. Use dark-mode assets on `#3b3a3c` / dark surfaces.

In the React shell, mount the full lockup as `<img src={logo} alt="Switchyard" style={{width:220}}/>` (or 280px for larger contexts).

---

## Theming

Four `--sy-*` variables drive the white-label system:

```css
/* themes/digital-parts.css — pilot client */
.theme-digital-parts {
  --sy-primary:        #3a73c1;
  --sy-primary-hover:  #2f63aa;
  --sy-primary-press:  #2a579a;
  --sy-primary-tint:   rgba(58,115,193,0.10);
  --sy-primary-ring:   rgba(58,115,193,0.40);
  --sy-graphite:       #3b3a3c;
  --sy-offwhite:       #eeeeee;
  --sy-accent:         #aa3bff;
  --sy-accent-bg:      rgba(170,59,255,0.10);
  --sy-accent-ring:    rgba(170,59,255,0.50);
}
```

The theme file also bridges `--sy-primary → --dp-blue`, etc., so all `--dp-*` tokens automatically pick up the tenant's primary color. Adding a new tenant = copy the file, change four values.

Dark mode: set `data-theme="dark"` on `<html>` (or `.dp-dark` on any container). All `--dp-*` tokens resolve to dark values.

---

## Borders

- Default: `1px solid var(--dp-border)` (`#e5e4e7`) — hairline, for cards, inputs, modal sheets
- Table grid: `1px solid #3b3a3c` (graphite) — deliberately heavier than default; the app shows a visible grid
- Nav signature rule: `2px solid #08060d` — never use elsewhere
- Focus ring: `outline: 2px solid var(--dp-accent); outline-offset: 2px`

---

## Hover & Active Patterns

- **Buttons:** color deepens ~10%, no scale
- **Cards:** border appears or shadow (`--dp-shadow-md`) appears
- **Nav items:** color changes to `--dp-blue`, `font-weight: 500` — no underline, no bg change
- **Dropdown items:** `background: var(--dp-accent-bg)` (10% purple tint)
- **Links:** bottom border appears (`border-bottom: 1px solid currentColor`)
- Never scale elements. Never use `backdrop-filter`. No colored shadows.

---

## Voice & Copy

**Tone:** Direct, functional, professional. No jokes, no sparkle, no exclamation marks.

**Casing:**
- Title Case: page titles, primary actions ("Create User", "Bills of Lading", "Log In")
- Sentence case: secondary copy, table cells, descriptions
- Logo wordmark: always ALL-CAPS ("DIGITAL PARTS", "SWITCHYARD")

**Pronouns:** No "I". "You" only for direct instructions ("Select a location type to begin"). Prefer imperatives. "We" for company voice.

**Abbreviations:** BOL (in table headers and tight UI), "Bills of Lading" in page titles. SKU, PPE assumed known. Always spell out "Warehouse" and "Store".

**Standard copy:**
```
Empty states:  "No users found."  "No items found."  "No line entries found."
Loading:       "Loading..."
Errors:        "Failed to load inventory."  "Failed to create user. Check all fields and try again."
Buttons:       "Search"  "Create User"  "Deactivate"  "View"  "Close"  "Log In"  "Log Out"
Placeholders:  "-- Select --"
Empty cells:   — (em-dash)
```

**Emoji:** Not used. Do not introduce them.

---

## Role Permissions

From `usePermissions.ts` — Auth0 `https://warehouselogistics/roles` claim:

| Capability | Employee | Manager | Admin |
|---|---|---|---|
| `canReadInventory` | ✓ | ✓ | ✓ |
| `canReadBOL` | ✓ | ✓ | ✓ |
| `canCreateBOL` | — | ✓ | ✓ |
| `canModifyBOL` | — | ✓ | ✓ |
| `canManageUsers` | — | — | ✓ |

Gate UI elements (nav items, action buttons) behind these booleans — don't rely solely on API 403s for UX.

---

## Dispatch Whiteboard — Supplemental Spec

### Column model

| # | Column | Primary card |
|---|---|---|
| 1 | Draft | BOL |
| 2 | Pending | BOL |
| 3 | Loading | BOL |
| 4 | Ready | BOL |
| 5 | In Delivery | Driver (BOLs tucked under) |
| 6 | Delivered | BOL |
| 7 | Available | Driver (ready or on rest) |
| 8 | Maintenance | Equipment |

### Rendering principle
The board is a pure projection of state: `render = f(active_state)`. Cards appear/move/disappear because their state maps to columns — never via imperative add/remove. Active set = BOLs with status ∈ { Draft, Pending, Loading, Ready, In Delivery, Delivered-not-filed }. Terminal state = **Filed** (POD confirmed + paperwork in) → card leaves on next reconcile.

### Card pattern (unified)
- Card bg: `--dp-surface-2` (elevated above canvas-deep)
- Border: **not yet implemented as family-colored.** Original intent was a permanent family-colored, load-bearing border (BOL/Driver/Tractor/Trailer) as the card-type signal when cards are fanned/stacked. Dropped in v1.3 pending this document; see [Card Border Language](#card-border-language) below for what's live today.
- Status: pip + status text always paired — pip alone insufficient
- No tinted card body. Shadow on hover/expanded only.
- Column containers: no bg, no border — label + count + hairline rule + cards

### Connection states
`auto` (30s poll, v1 default) → `live` (SSE, v1.x) → `reconnecting` → `manual`. Indicator: colored pip + text string. "Updates available" banner appears full-width below nav in manual state only.

### Affordances rule
> If the action changes which **column** a card sits in → Whiteboard action.
> If it changes which **card** it is (assignment, field edit, reroute) → upstream page action.

### Card Border Language

Card borders on the dispatch board carry semantic weight — they are not decorative, and must never be added or removed for visual variety.

| Border | CSS | Meaning |
|---|---|---|
| Neutral hairline | `border: 1px solid var(--border)` — the default `.card` rule | Ordinary chrome, present on every card. Not itself a status signal. |
| Danger | `.card.urgent` → `border-color: var(--danger)` (`#C23B3B`) | Status alert — requires immediate attention (e.g. roadside breakdown with load attached). |
| Warn | `.card.long-wait` → `border-color: var(--warn)` (`#D97706`) | Status alert — attention needed, not critical (e.g. long-wait threshold exceeded). |
| Ok | `.card.is-ready` → `border-left: 3px solid var(--ok)` (`#1F9D55`) | Status alert — acknowledged / ready to advance (e.g. trailer loaded and ready to depart). |

**The rule:** a colored border variant (danger/warn/ok) means the card has an active, specific status condition. A card with no active condition renders only the neutral hairline. Never apply `.urgent`, `.long-wait`, `.is-ready`, or an equivalent colored-border class to a card as decoration or for visual emphasis alone — if you're reaching for a border to make a card "stand out," that's a sign the card needs a real status condition, a pill, or better copy instead.

**Empty Return (v1.4):** Empty Return driver cards follow the same system. A driver on-schedule for empty return carries only the neutral hairline — no colored border. A colored border only appears if a real condition applies (e.g. missed return ETA).

> **Not yet built:** an earlier design also called for a permanent family-colored border per card type (BOL/Driver/Tractor/Trailer) as a card-type identifier, independent of status — see the note under [Card pattern (unified)](#card-pattern-unified) above. That work was deferred in v1.3 until this section existed. If it ships, it needs a visually distinct treatment (e.g. different edge or weight) from the danger/warn/ok status border so the two systems don't collide on the same card.

---

## What Is Not Built Yet

- Customer-facing Sales UI (checkout flow)
- Inventory withdrawal UI
- Analytics endpoints and Dashboard reports (placeholder: `<p>Company reports coming soon.</p>`)
- Scalar API doc branding
- Light mode for Switchyard-Go (scheduled for v1.2)
- System-wide theme preference synced via Auth0 (v1.x)
- Full-name SVG assets with text outlined to paths (v1.2 asset task)
