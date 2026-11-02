---
title: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache and delete the kept overview on sign-out, drop its cache on a new version, and show the offline page when the network is gone, with the overview kept for offline reading on the overview's address and nothing of the record elsewhere.
status: open
text_ref: docs/DESIGN.md#D-151
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-151
  - lane:strict
semantic_text: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache and delete the kept overview on sign-out, drop its cache on a new version, and show the offline page when the network is gone, with the overview kept for offline reading on the overview's address and nothing of the record elsewhere.
logic_claims:
  - CLAIM-3C23A7DAE2E8C42A
semantic_clauses:
  - The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache and delete the kept overview on sign-out, drop its cache on a new version, and show the offline page when the network is gone, with the overview kept for offline reading on the overview's address and nothing of the record elsewhere.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4b4fc9cbf904897f326a590e76745d05d82a0a43529be23ca998fd22dec2f7a6
semantic_inventory:
  - claim_key: CLAIM-3C23A7DAE2E8C42A
    claim_text: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache and delete the kept overview on sign-out, drop its cache on a new version, and show the offline page when the network is gone, with the overview kept for offline reading on the overview's address and nothing of the record elsewhere
    role: normative
    status: modeled
    span:
      start: 0
      end: 425
origin:
  kind: agent
  recorded_at: '2026-11-02T10:11:12.608Z'
id: REQ-ui-service-worker-static-files
type: req
---
The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache and delete the kept overview on sign-out, drop its cache on a new version, and show the offline page when the network is gone, with the overview kept for offline reading on the overview's address and nothing of the record elsewhere.

## Context

Pages and fragments hold the owner's record, which the UI refuses to let browsers keep (Cache-Control no-store on pages); a service worker must not undo that. D-151 lets the installed app keep the last overview for offline reading, so the offline page is no longer always empty. Recorded by the implementation agent of D-151 (impl/d151) with default firmness, refining the firm D-151 requirements the author recorded on 2026-11-01; the author has not reviewed this refinement yet.

## Why this changed

The author decided D-151, which lets the installed app keep the last overview offline, so this replaces REQ-ui-service-worker-static-only, whose offline page showed no data; that requirement is closed.

## Source

> The installed app keeps only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from. It is encrypted with a key the browser can't export and wiped on sign-out or when the session ends. Transactions, lots and account numbers are never kept.

docs/DESIGN.md#D-151
