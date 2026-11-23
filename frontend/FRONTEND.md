# Frontend guide

> Design and implementation guidance for the diploma project. This document defines the intended interface; it does not claim that the screens or features are implemented.

## 1. Read this first

Read [AGENTS.md](../AGENTS.md) before changing code. It defines the project scope, API contracts, evidence rules, security requirements, and milestone gates. This guide adds frontend design decisions. For field names and response types, use the [implemented OpenAPI contract](../contracts/openapi.yaml) and its generated TypeScript types.

Before working, inspect the frontend in the current checkout and read the [implementation status](../docs/status.md). Keep working components that meet the requirements. The [backend guide](../backend/BACKEND.md) covers the API and persistence boundaries.

**Design goal: a quiet, minimal workspace that makes the next useful action obvious.** Prioritize readable findings and a short import flow. Show a control only when it has a working purpose on that screen.

## 2. Visual direction

Use a light interface with white surfaces, a soft gray background, dark text, and thin borders. Use whitespace and typography to establish hierarchy.

| Element | Default |
| --- | --- |
| Page background | `#F7F7F5` |
| Surface | `#FFFFFF` |
| Primary text | `#202020` |
| Secondary text | `#595959` |
| Decorative border | `#DEDEDA` |
| Interactive control border | `#777777` |
| Primary button | `#252525` background, `#FFFFFF` text |
| Primary button hover | `#3A3A3A` |
| Error text | `#9B2C2C` on a light surface |
| Focus outline | `#202020`, 2 px with 3 px offset |

These are project design tokens. Check contrast in the actual rendered combinations, including focus, disabled, error, and hover states. Decorative borders must not be the only boundary identifying an input.

### Required styling rules

- No neon colors, glowing borders, gradients, glass effects, or animated backgrounds.
- No decorative charts, oversized statistics, marketing hero, slogan, security-themed wallpaper, or fake terminal.
- No large sidebar for three screens. Use a small header and ordinary back links.
- Use the system sans-serif font. Use a system monospace font for vectors, hashes, identifiers, and source pointers only.
- Default body text: 16 px, line height 1.5. Page titles: 24–28 px. Section titles: 18–20 px.
- Use a 4 px spacing scale: 4, 8, 12, 16, 24, 32. Main content maximum width: 1120 px. Side padding: 24 px on desktop and 16 px on small screens.
- Use 6 px corner radii for controls and panels. Avoid shadows by default.
- Aim for 40 px control height. Keep buttons large enough to operate comfortably; do not shrink them into tiny icons to reduce visual weight.
- At most one primary button in the current task section. Secondary actions use an outlined button or text link.
- Use restrained text labels for severity and verification. Red is for errors or specific warnings, not the entire interface. Color never carries meaning by itself.
- Start with one light theme. Do not add a theme switcher, animation library, icon library, or design system dependency solely for decoration.

## 3. Navigation and page hierarchy

The initial workflow is:

```text
Unlock → Select or create lab target → Import report → Read findings → Read evidence
```

Keep the shared header small: project name on the left; `Lock workspace` on the right when unlocked. Use a plain title such as `Vulnerability assessment` rather than a promotional headline.

Initial pages:

| Page | Purpose | Main action |
| --- | --- | --- |
| Unlock | Enter the local API token | Unlock |
| Targets | Select a registered lab target | Create target, only when needed |
| Target findings | Import a report and inspect its findings | Import report |
| Finding detail | Read observations and provenance | Open source report details when needed |

Keep later review, context, assessment, ranking, and scan controls absent until their milestones and APIs are implemented. Do not display inactive feature tiles or placeholder navigation.

## 4. Essential controls only

| Control | Show it when | Behavior |
| --- | --- | --- |
| Unlock | Workspace is locked | Submit the token form |
| Lock workspace | Workspace is unlocked | Clear the token, query cache, and private page state |
| Create target | A new registered target is needed | Reveal a short name form; server supplies the origin |
| Import report | A target is selected | Submit one selected report with optional provenance |
| Filters | Findings can be filtered | Reveal less frequently used filter fields inline |
| Clear filters | At least one filter is active | Restore the unfiltered list |
| Pagination | Another page or a return path exists | Navigate using the API cursor |
| Download original report | Operator opens private evidence details | Require explicit acknowledgement; use audited access |

Use links for navigation and buttons for actions. Use native `<details>` for simple disclosures. Avoid an overflow menu when one labeled link is sufficient.

Do not add refresh, delete, bulk selection, scan, export, compare, copy, settings, or notification buttons unless the feature is implemented and the current task needs it. A scan launcher is a later milestone with separate authorization requirements.

Do not ask for confirmation for ordinary navigation, filtering, or importing. Preserve the explicit acknowledgement required for original private evidence access. Minimalism must not remove that protection.

## 5. Screen specifications

### Unlock

Show a short title, one labeled password input, and `Unlock`. Explain in one sentence that the token stays in memory and clears on reload or lock. Do not include a registration flow or store the token in localStorage, sessionStorage, URLs, analytics, or the build.

If authentication fails, show `Token not accepted. Check it and try again.` Do not show the token in the error. Do not represent a token as accepted merely because it was entered; protected data must load successfully before claiming authentication succeeded.

### Targets

Show a short list of existing target names and registered origins. Each name links to its findings. If there are no targets, provide a small name input and `Create target`.

The initial backend permits one target per registered canonical origin. If that target already exists, offer its link instead of inviting duplicate creation. Do not ask users to enter arbitrary scan URLs. Do not automatically create a target during page load.

### Target findings

Use this hierarchy:

```text
← Targets                              Lock workspace

Juice Shop
Registered origin

Import report
[Choose file]                         [Import report]
▸ Optional report provenance

Findings                         [Severity ▾] [Filters]
Finding / location | Scanner severity | Verification | OWASP category | CVSS
…persisted findings…

Pagination, only when needed
▸ Source report details
```

This is a layout sketch, not a screenshot or implementation claim. At narrow widths, stack the import controls. Keep the findings table readable with contained horizontal scrolling and a clear accessible label; avoid horizontal scrolling for the entire page.

The file input accepts JSON. Briefly state the supported ZAP report format and upload limit. File type and size checks in the browser are convenience checks; the server remains authoritative.

Provenance is optional at import time because unknown metadata is allowed, but incomplete provenance must remain visible afterward. Label it `Provenance incomplete` when appropriate. Collapsing the form must not silently discard entered values. Keep report-declared metadata visibly distinct from operator-supplied provenance.

Place common severity filtering beside the list. Put report ID, verification, and OWASP category filters in a compact inline disclosure. All filters required by AGENTS.md must remain usable. Explain that findings are unmapped until reviewed rules exist; never fabricate categories to populate a filter.

Show a short inline import result with counts and a link/disclosure to provenance. Keep full hashes, parser versions, detailed warnings, and raw JSON out of the main list while preserving access to them. Warn clearly when an import has unlocated alerts or incomplete provenance.

The findings table contains separate columns for scanner severity, verification, category, and CVSS. Show the mapped category identifier when available, otherwise the classification state. Do not collapse them into a single risk badge. Use the API's observation count and severity source; do not calculate a new aggregate in the browser.

### Finding detail

Show a back link, finding name, method/path, and a compact definition list of its current states. Put observations in source order as supplied by the API. Each observation exposes its report reference, source pointer, raw confidence code, CWE assertions, description, remediation, and redaction notice.

When the user arrives from a filtered list, preserve a safe internal return URL so `Back to findings` restores that list. Do not accept external redirect URLs.

Keep long hashes, fingerprint details, provenance, and historical records in labeled disclosures. Render scanner HTML as text. Do not turn report URLs into previews or load external images, links, frames, or scripts from report content.

Private attack/evidence payloads stay withheld or redacted according to the API. An empty excerpt must never be described as evidence that no attack occurred. Original-report download remains explicit and audited.

## 6. States, feedback, and wording

| Situation | Required display |
| --- | --- |
| No reports | `No reports imported.` with the import action |
| Report contains no located findings | State that result separately from no reports |
| Active filters match nothing | `No findings match these filters.` plus Clear filters |
| Loading | Short text such as `Loading findings…`; no fake metrics |
| Upload pending | `Importing…`; prevent duplicate submission |
| Exact duplicate | `This report was already imported.` and the existing record |
| Conflicting provenance | Explain that stored provenance is immutable; do not imply new metadata was saved |
| Validation error | Safe explanation next to the field and source pointer where relevant |
| Unexpected failure | Safe error and request ID; preserve useful form input |
| Network outcome uncertain | Say the outcome is unknown; offer deliberate recovery, never blind mutation retry |
| Missing assessment | `Not assessed` |
| Incomplete assessment | `Incomplete` |
| Complete assessment with score 0 | Display `0`, with its vector available |
| Stale assessment | Label it historical/stale; do not present it as current |

Read all review, classification, and assessment states from API data. Avoid hard-coded future status displays. Do not invent scores, counts, dates, successful operations, or scan coverage.

Use inline notices for persistent results, not disappearing toasts. Preserve filters in the URL where practical. Clear stale cursor state when filters or the selected finding change. On authentication failure, provide a clear way to lock and enter a new token.

## 7. Frontend implementation boundaries

- Use the stack and locked dependencies defined in AGENTS.md and the actual repository.
- Keep the typed API client in `src/api`, routing/providers in `src/app`, and feature screens in `src/features`.
- Extract cohesive components as screens grow: `finding-table`, `import-form`, `report-details`, `observation-list`, `error-state`, and `pagination` are reasonable candidates. Do not create every proposed file in advance.
- Use TanStack Query for remote data and mutation status. Keep temporary form state local. Clear private caches on lock.
- Use generated API types. Do not reproduce URL scope checks, fingerprints, CVSS calculations, classification rules, or eligibility logic in TypeScript.
- Never automatically retry import or other mutations after an uncertain response. Duplicate handling is defined by the backend contract.
- Avoid unnecessary abstractions and package additions. Prefer semantic HTML and a small shared CSS token file.

## 8. Accessibility and verification

Every form control has a visible label. Every action is keyboard accessible. Focus remains visible, and submission errors are announced without moving focus unpredictably. Use table headers, a caption, and accessible status/error regions. Disclosure titles describe their content. Critical information must not depend on hover, color, or animation.

For a redesign, preserve existing working behavior and run relevant checks. Verify:

- Token entry, failed authentication, lock, and reload clearing.
- Import, duplicate result, conflicting provenance, validation errors, and uncertain network outcomes.
- Loading, empty, filtered-empty, and failure states.
- Filtering, pagination, browser back navigation, and persisted findings after restart.
- Safe rendering of scanner HTML and withheld payload notices.
- Distinct unknown, unreviewed, unmapped, incomplete, stale, and zero-score states.
- Acknowledgement-gated private download and no unsolicited report URL requests.
- Keyboard flow, visible focus, mobile layout, and desktop table readability.

Check the actual rendered interface, not only compilation. Report which browser checks ran and which remain unverified. Do not launch scans as part of frontend testing.

## 9. Definition of done

The interface is ready for its current milestone when the core task can be completed with clear labels and few controls, every visible action works, factual states come from persisted API data, private data stays protected, and the applicable AGENTS.md acceptance gates have executable evidence.

Simplicity means fewer distractions and clearer choices. It must not remove evidence traceability, error information, accessibility, or necessary review controls.
