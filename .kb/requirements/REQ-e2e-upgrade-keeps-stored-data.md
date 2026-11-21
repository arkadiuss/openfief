---
title: An in-place upgrade of an install of the first published release, 2026.11.0-alpha.1, to a release candidate must keep every account, transaction, lot, valuation and attention item the install stored unchanged, run the household-grant cutover once, create the SpiceDB credentials, leave the agentic switch off and leave the core ready with the UI serving.
status: open
firmness: default
required_evidence: unverified
text_ref: docs/dev/sdlc.md#releases
tags:
  - e2e
  - release
origin:
  kind: agent
  ref: the design session's exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
  recorded_at: '2026-11-21T12:00:00.000Z'
semantic_text: An in-place upgrade of an install of the first published release, 2026.11.0-alpha.1, to a release candidate must keep every account, transaction, lot, valuation and attention item the install stored unchanged, run the household-grant cutover once, create the SpiceDB credentials, leave the agentic switch off and leave the core ready with the UI serving.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a5bb7de2783dbfdc2af1eed2baca163291123f0393e40bb112860109beccf275
logic_claims:
  - CLAIM-55DC726ED5C4A36D
semantic_inventory:
  - claim_key: CLAIM-55DC726ED5C4A36D
    claim_text: An in-place upgrade of an install of the first published release, 2026.11.0-alpha.1, to a release candidate must keep every account, transaction, lot, valuation and attention item the install stored unchanged, run the household-grant cutover once, create the SpiceDB credentials, leave the agentic switch off and leave the core ready with the UI serving
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 353
    subjects:
      - release.upgrade_in_place
id: REQ-e2e-upgrade-keeps-stored-data
type: req
---
An in-place upgrade of an install of the first published release, 2026.11.0-alpha.1, to a release candidate must keep every account, transaction, lot, valuation and attention item the install stored unchanged, run the household-grant cutover once, create the SpiceDB credentials, leave the agentic switch off and leave the core ready with the UI serving.

## Context

The design session set an upgrade test as exit criterion (a) of 2026.11.0-alpha.2: D-161 promises that stored data is migrated from the first release on, and the first published release is 2026.11.0-alpha.1. The same upgrade is where D-209's household-grant cutover first runs on an owner's data, where D-216's credentials Job first makes SpiceDB's Secrets, and where D-212's switch must stay off.

## Source

> From the first release on, every change to stored data ships with a migration of what is already stored.

Source: D-161 in docs/DESIGN.md; the design session's brief for exit criterion (a) of 2026.11.0-alpha.2, relayed by the coordinator, 2026-11-21
