---
title: The origins the core accepts must be one setting that serving and issuing enrollment links both read.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-126-1
  - lane:strict
text_ref: design-decisions:D-126#1
origin:
  kind: import
  ref: design-decisions:D-126#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: The origins the core accepts must be one setting that serving and issuing enrollment links both read.
logic_claims:
  - CLAIM-47B604B8F47B0813
semantic_clauses:
  - The origins the core accepts must be one setting that serving and issuing enrollment links both read.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 711b2854972a8cb128c1a84d28d488ba64b07d044515c64a107d7736f9edc717
semantic_inventory:
  - claim_key: CLAIM-47B604B8F47B0813
    claim_text: The origins the core accepts must be one setting that serving and issuing enrollment links both read
    role: normative
    status: modeled
    span:
      start: 0
      end: 100
id: REQ-identity-origins-one-setting
type: req
---
The origins the core accepts must be one setting that serving and issuing enrollment links both read.

## Context

D-80 requires the origin given to the first-passkey command to be one the core allows (REQ-identity-enroll-origin), but `core enroll-link` runs as its own process in the core's pod, without `serve`'s flags. The implementation raised this as question #10 on docs/implementation-questions.md: today the CLI reads the `--auth-origin` flags from the running core pod's container spec and refuses any other origin itself. On 2026-10-31 the author chose the other option, one setting that serving and enrollment both read, as D-126, so the core itself refuses an origin it doesn't accept (REQ-identity-core-refuses-enroll-origin).

## Source

> The origins the core accepts are one setting that serving and issuing enrollment links both read, so the core itself refuses an enrollment link for an origin it doesn't accept.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-126#1
