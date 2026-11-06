---
title: The overview page must hand the installed app, to keep offline, only the total's gross, net and liquidation values, each account's name, institution, sleeve and three values, the allocation's dimension, value and shares, the date they are valued on and when the session ends; no other page or fragment may hand over anything to keep, and no script, htmx included, may keep a page in localStorage or sessionStorage.
status: open
text_ref: docs/DESIGN.md#D-151
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-151
  - lane:strict
semantic_text: The overview page must hand the installed app, to keep offline, only the total's gross, net and liquidation values, each account's name, institution, sleeve and three values, the allocation's dimension, value and shares, the date they are valued on and when the session ends; no other page or fragment may hand over anything to keep, and no script, htmx included, may keep a page in localStorage or sessionStorage.
logic_claims:
  - CLAIM-B17A01D0B97B10A6
semantic_clauses:
  - The overview page must hand the installed app, to keep offline, only the total's gross, net and liquidation values, each account's name, institution, sleeve and three values, the allocation's dimension, value and shares, the date they are valued on and when the session ends; no other page or fragment may hand over anything to keep, and no script, htmx included, may keep a page in localStorage or sessionStorage.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c762733851af26b30b90f5e0560d627d752d4095d8ab339412acd8af9f88909f
semantic_inventory:
  - claim_key: CLAIM-B17A01D0B97B10A6
    claim_text: The overview page must hand the installed app, to keep offline, only the total's gross, net and liquidation values, each account's name, institution, sleeve and three values, the allocation's dimension, value and shares, the date they are valued on and when the session ends; no other page or fragment may hand over anything to keep, and no script, htmx included, may keep a page in localStorage or sessionStorage
    role: normative
    status: modeled
    span:
      start: 0
      end: 413
origin:
  kind: agent
  recorded_at: '2026-11-02T10:11:21.195Z'
id: REQ-ui-offline-overview-contents
type: req
---
The overview page must hand the installed app, to keep offline, only the total's gross, net and liquidation values, each account's name, institution, sleeve and three values, the allocation's dimension, value and shares, the date they are valued on and when the session ends; no other page or fragment may hand over anything to keep, and no script, htmx included, may keep a page in localStorage or sessionStorage.

## Context

Refines REQ-ui-offline-keeps-last-overview-only and REQ-ui-offline-never-keeps-records: which parts of the overview count is the UI's reading, the allocation's shares without their amounts and the accounts' names so the values can be told apart. htmx's history cache would otherwise keep snapshots of pages in sessionStorage when a switch pushes its URL. Recorded by the implementation agent of D-151 (impl/d151) with default firmness, refining the firm D-151 requirements the author recorded on 2026-11-01; the author has not reviewed this refinement yet.

## Source

> The installed app keeps only the last overview on the device for offline reading: totals, each account's values and the allocation's shares, with the time they're from. It is encrypted with a key the browser can't export and wiped on sign-out or when the session ends. Transactions, lots and account numbers are never kept.

docs/DESIGN.md#D-151
