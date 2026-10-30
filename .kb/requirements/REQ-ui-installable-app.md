---
title: 'The built-in UI must be installable as an app: it serves a web app manifest with standalone display, start URL and scope at its root, the design tokens'' colors and app icons from the design system, and its CSP admits only its own manifest.'
status: open
text_ref: docs/DESIGN.md#D-69
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-69
  - lane:strict
semantic_text: 'The built-in UI must be installable as an app: it serves a web app manifest with standalone display, start URL and scope at its root, the design tokens'' colors and app icons from the design system, and its CSP admits only its own manifest.'
logic_claims:
  - CLAIM-FBD2DECD7D0CE78B
semantic_clauses:
  - 'The built-in UI must be installable as an app: it serves a web app manifest with standalone display, start URL and scope at its root, the design tokens'' colors and app icons from the design system, and its CSP admits only its own manifest.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 754e1bf022348c531703a334001afc98168f0fbd7bb4cd25c9ac87e6fece8bfa
semantic_inventory:
  - claim_key: CLAIM-FBD2DECD7D0CE78B
    claim_text: 'The built-in UI must be installable as an app: it serves a web app manifest with standalone display, start URL and scope at its root, the design tokens'' colors and app icons from the design system, and its CSP admits only its own manifest'
    role: normative
    status: modeled
    span:
      start: 0
      end: 238
origin:
  kind: agent
  recorded_at: '2026-10-30T08:08:15.489Z'
id: REQ-ui-installable-app
type: req
---
The built-in UI must be installable as an app: it serves a web app manifest with standalone display, start URL and scope at its root, the design tokens' colors and app icons from the design system, and its CSP admits only its own manifest.

## Context

The author asked for a mobile PWA. The UI is a backend-for-frontend that browsers reach only through the core (D-69, D-81), so the app is the same origin and the same pages, signed in with the same passkey; nothing about the session changes. The icons use the Seal, which the design system names as the app icon. Modeled by Claude Code for the author's request for a responsive UI and a mobile PWA; the author has not reviewed it yet.

## Source

> The built-in UI is a backend-for-frontend: the browser talks only to the UI view plugin.

docs/DESIGN.md D-69

> Seal ... Use for: App icon, docs hero, approval requests, stickers

docs/dev/design-system.md#the-marks
