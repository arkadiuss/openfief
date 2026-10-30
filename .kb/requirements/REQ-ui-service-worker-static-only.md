---
title: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache on sign-out and drop it on a new version, and show the offline page, with no data, when the network is gone.
status: open
text_ref: docs/DESIGN.md#D-69
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-69
  - lane:strict
semantic_text: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache on sign-out and drop it on a new version, and show the offline page, with no data, when the network is gone.
logic_claims:
  - CLAIM-8CA1CB1405056DD7
semantic_clauses:
  - The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache on sign-out and drop it on a new version, and show the offline page, with no data, when the network is gone.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bb1efc4f15ae7b6f990ae5f6db004d9154e9a4dcc7dcfb924a6cdfadcf7a589a
semantic_inventory:
  - claim_key: CLAIM-8CA1CB1405056DD7
    claim_text: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache on sign-out and drop it on a new version, and show the offline page, with no data, when the network is gone
    role: normative
    status: modeled
    span:
      start: 0
      end: 301
origin:
  kind: agent
  recorded_at: '2026-10-30T08:08:23.164Z'
id: REQ-ui-service-worker-static-only
type: req
---
The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache on sign-out and drop it on a new version, and show the offline page, with no data, when the network is gone.

## Context

Pages and fragments hold the owner's record, which the UI already refuses to let browsers keep (Cache-Control no-store on pages); a service worker must not undo that. Offline support beyond a static page, or caching data, is out of scope until the author decides it. Modeled by Claude Code for the author's request for a responsive UI and a mobile PWA; the author has not reviewed it yet.

## Source

> The built-in UI is a backend-for-frontend: the browser talks only to the UI view plugin, and the plugin calls the core's authentication service.

docs/DESIGN.md D-69

> Empty states. Kasztan asleep, one line saying what's missing, and the action that fills it.

docs/dev/design-system.md#components
