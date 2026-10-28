---
title: A verified catalog entry must pin the publisher's signed image digest and the publisher's signer identity.
status: open
text_ref: design-decisions:D-44#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-44-1
  - lane:strict
semantic_text: A verified catalog entry must pin the publisher's signed image digest and the publisher's signer identity.
logic_claims:
  - CLAIM-3FA438C8B5D461B6
semantic_clauses:
  - A verified catalog entry must pin the publisher's signed image digest and the publisher's signer identity.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7471aa8c8e5f05eff1477ae17cc3ab53625810f642fe7e91fc96c88917aaabf8
semantic_inventory:
  - claim_key: CLAIM-3FA438C8B5D461B6
    claim_text: A verified catalog entry must pin the publisher's signed image digest and the publisher's signer identity
    role: normative
    status: modeled
    span:
      start: 0
      end: 105
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-44#1
  recorded_at: '2026-10-28T05:38:48.763Z'
required_evidence: asserted
id: REQ-plugins-verified-entry-pins
type: req
---
A verified catalog entry must pin the publisher's signed image digest and the publisher's signer identity.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> A verified plugin is one whose version the author reviewed: its catalog entry pins the publisher's signed image digest and the publisher's signer identity. Each new version needs a new catalog entry.

OpenFief author decisions (docs/DESIGN.md, section 3.5), design-decisions:D-44#1
