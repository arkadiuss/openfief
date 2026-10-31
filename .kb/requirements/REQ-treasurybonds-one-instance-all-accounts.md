---
title: A single treasury-bonds connector instance must be able to serve every bond account.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-118-2
  - lane:strict
text_ref: design-decisions:D-118#2
origin:
  kind: import
  ref: design-decisions:D-118#2
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: A single treasury-bonds connector instance must be able to serve every bond account.
logic_claims:
  - CLAIM-E89976CA2870FDC1
semantic_clauses:
  - A single treasury-bonds connector instance must be able to serve every bond account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 63b086b40c221420cbdb4857508339f7b229c166916668d1a3145fa018d9b94e
semantic_inventory:
  - claim_key: CLAIM-E89976CA2870FDC1
    claim_text: A single treasury-bonds connector instance must be able to serve every bond account
    role: normative
    status: modeled
    span:
      start: 0
      end: 83
id: REQ-treasurybonds-one-instance-all-accounts
type: req
---
A single treasury-bonds connector instance must be able to serve every bond account.

## Context

Once the core hands each bond lot to the connector with the account it was entered on (REQ-ingest-bond-lot-handed-with-account), the connector no longer needs one instance per account: the owner installs it once and binds it to each bond account, as with any connector instance (D-48). The author decided this on 2026-10-31, answering the implementation's question about the connector naming its account in its config (question #4 on docs/implementation-questions.md).

## Source

> The core hands each bond lot to the connector together with the account the owner entered it on, so one treasury-bonds instance serves every bond account and its config names none.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-118#2
