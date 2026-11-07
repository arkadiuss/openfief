---
title: The balance report must list, for each connector instance whose latest completed import batch left rows out as unsupported, such as open XTB CFD positions, that batch and the number of rows, flagged as unsupported and not checked.
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: The balance report must list, for each connector instance whose latest completed import batch left rows out as unsupported, such as open XTB CFD positions, that batch and the number of rows, flagged as unsupported and not checked.
logic_claims:
  - CLAIM-A6F4BC94BBFA47A0
semantic_clauses:
  - The balance report must list, for each connector instance whose latest completed import batch left rows out as unsupported, such as open XTB CFD positions, that batch and the number of rows, flagged as unsupported and not checked.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b536450339876f8e418150881fd5bd6cf040f401c007310d0cffe0e1793ecae7
semantic_inventory:
  - claim_key: CLAIM-A6F4BC94BBFA47A0
    claim_text: The balance report must list, for each connector instance whose latest completed import batch left rows out as unsupported, such as open XTB CFD positions, that batch and the number of rows, flagged as unsupported and not checked
    role: normative
    status: modeled
    span:
      start: 0
      end: 229
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:25.261Z'
id: REQ-views-balance-check-unsupported
type: req
---
The balance report must list, for each connector instance whose latest completed import batch left rows out as unsupported, such as open XTB CFD positions, that batch and the number of rows, flagged as unsupported and not checked.

## Context

D-200 criterion 1 is the owner's own check on a private install with real accounts, never in CI (GR-12). The design session asked, through the orchestrator on 2026-11-06, for an owner-only command, `openfief check balances`, so the author can check it: every balance assertion with reported and derived values, the difference and pass or fail at 0.01 of the account currency, open XTB CFDs flagged as unsupported (D-24), values printed only to the owner's terminal, the core logging only that the check ran with counts, a non-zero exit status on a failure, a table and --json. Open CFD rows are ingest problems (D-24) whose count per batch the ingest store keeps; the rows themselves are not stored, so the report counts them per instance.

## Source

> Every broker-reported position and cash balance is stored as a balance assertion and matches the derived value to 0.01 of the account currency. Open XTB CFDs appear only as flagged, unsupported items.

docs/DESIGN.md#D-200
