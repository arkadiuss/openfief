---
title: A transactions upload of a statement covering several accounts must book every mapped account bound to that connector instance.
status: open
text_ref: design-decisions:D-140#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-140-3
  - lane:strict
semantic_text: A transactions upload of a statement covering several accounts must book every mapped account bound to that connector instance.
logic_claims:
  - CLAIM-E393C41DF54FCB0E
semantic_clauses:
  - A transactions upload of a statement covering several accounts must book every mapped account bound to that connector instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 782297a5a92a4d64ab7c548338fa17e4dce462e9d5af785363278654126f733c
semantic_inventory:
  - claim_key: CLAIM-E393C41DF54FCB0E
    claim_text: A transactions upload of a statement covering several accounts must book every mapped account bound to that connector instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 126
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-140#3
  recorded_at: '2026-11-01T12:24:07.514Z'
id: REQ-ingest-multi-account-transactions-books-every-mapped
type: req
---
A transactions upload of a statement covering several accounts must book every mapped account bound to that connector instance.

## Why this changed

On 2026-11-01 the author recorded D-140, answering the implementation's question #24: an upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, so an opening upload seeds only the accounts not seeded yet and a transactions upload books every mapped account. Reason not stated. REQ-ingest-statement-multi-account had such a statement book every mapped account whatever the mode, so this requirement replaces it together with REQ-ingest-multi-account-opening-seeds-unseeded-only and REQ-ingest-statement-multi-account-reports-unmapped.

## Context

An IBKR Flex statement of linked accounts is such a statement (D-109). A broker account is mapped to an OpenFief account by the owner (D-72), and nothing of an unmapped one is booked (REQ-ingest-unmapped-broker-account). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, which the owner can override: an opening upload seeds only the accounts not seeded yet, and a transactions upload books every mapped account.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-140#3
