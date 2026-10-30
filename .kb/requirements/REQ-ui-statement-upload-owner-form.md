---
title: The built-in UI must offer the statement upload form to the owner only, preselect the core's choice of mode while letting the owner override it, stream the posted file to the core without storing it, and show the import's mode, counts, booked accounts, unmapped broker accounts and problems or the core's reason.
status: open
text_ref: docs/DESIGN.md#D-109
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-109
  - lane:strict
semantic_text: The built-in UI must offer the statement upload form to the owner only, preselect the core's choice of mode while letting the owner override it, stream the posted file to the core without storing it, and show the import's mode, counts, booked accounts, unmapped broker accounts and problems or the core's reason.
logic_claims:
  - CLAIM-CB984A6AEF018D5A
semantic_clauses:
  - The built-in UI must offer the statement upload form to the owner only, preselect the core's choice of mode while letting the owner override it, stream the posted file to the core without storing it, and show the import's mode, counts, booked accounts, unmapped broker accounts and problems or the core's reason.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 250463f2f80e56e40369b189fde2fb5f45983bfd64442d8f50111869422f758b
semantic_inventory:
  - claim_key: CLAIM-CB984A6AEF018D5A
    claim_text: The built-in UI must offer the statement upload form to the owner only, preselect the core's choice of mode while letting the owner override it, stream the posted file to the core without storing it, and show the import's mode, counts, booked accounts, unmapped broker accounts and problems or the core's reason
    role: normative
    status: modeled
    span:
      start: 0
      end: 311
origin:
  kind: agent
  recorded_at: '2026-10-30T05:18:06.659Z'
id: REQ-ui-statement-upload-owner-form
type: req
---
The built-in UI must offer the statement upload form to the owner only, preselect the core's choice of mode while letting the owner override it, stream the posted file to the core without storing it, and show the import's mode, counts, booked accounts, unmapped broker accounts and problems or the core's reason.

## Context

Restores, as a refinement, details of the superseded REQ-ui-statement-upload: owner only (D-79), the file streamed through, never stored, and the core's answer shown; adds D-109's mode choice and its report of booked and unmapped accounts.

## Source

> The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there.

docs/DESIGN.md#D-109
