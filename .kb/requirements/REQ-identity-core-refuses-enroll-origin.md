---
title: The core itself must refuse an enrollment link for an origin it doesn't accept.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-126-2
  - lane:strict
text_ref: design-decisions:D-126#2
origin:
  kind: import
  ref: design-decisions:D-126#2
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: The core itself must refuse an enrollment link for an origin it doesn't accept.
logic_claims:
  - CLAIM-85FA584BEF9FE95D
semantic_clauses:
  - The core itself must refuse an enrollment link for an origin it doesn't accept.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 6d4798ec366a1fb7e95d6a30d136d648e3b880cf51b69cf76337f87250c5ed3c
semantic_inventory:
  - claim_key: CLAIM-85FA584BEF9FE95D
    claim_text: The core itself must refuse an enrollment link for an origin it doesn't accept
    role: normative
    status: modeled
    span:
      start: 0
      end: 78
id: REQ-identity-core-refuses-enroll-origin
type: req
---
The core itself must refuse an enrollment link for an origin it doesn't accept.

## Context

A passkey works only on the hostname it was created on (D-60), and the core accepts the owner's browser only on the origins it allows (REQ-identity-origins-per-host), so a link for any other origin would create a passkey that can never log in. The implementation raised this as question #10 on docs/implementation-questions.md: `core enroll-link` runs without `serve`'s flags, so today only the CLI checks the origin, against the `--auth-origin` flags it reads from the core pod's container spec. On 2026-10-31 the author decided, as D-126, that the origins become one setting both read (REQ-identity-origins-one-setting), so the core itself refuses.

## Source

> The origins the core accepts are one setting that serving and issuing enrollment links both read, so the core itself refuses an enrollment link for an origin it doesn't accept.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-126#2
