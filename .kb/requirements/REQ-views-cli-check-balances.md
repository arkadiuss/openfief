---
title: openfief check balances must print the balance report only to the owner's terminal, as a table or as JSON with --json, and exit with a non-zero status when any balance assertion fails; the core must log only that the check ran, with its counts, never an account, an instrument or a value.
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: openfief check balances must print the balance report only to the owner's terminal, as a table or as JSON with --json, and exit with a non-zero status when any balance assertion fails; the core must log only that the check ran, with its counts, never an account, an instrument or a value.
logic_claims:
  - CLAIM-5EF6D06383264EBE
semantic_clauses:
  - openfief check balances must print the balance report only to the owner's terminal, as a table or as JSON with --json, and exit with a non-zero status when any balance assertion fails; the core must log only that the check ran, with its counts, never an account, an instrument or a value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 53ec2b8012a925d04ba1ab4806245bb80675dffc5214c67e115cb4b428a9923e
semantic_inventory:
  - claim_key: CLAIM-5EF6D06383264EBE
    claim_text: openfief check balances must print the balance report only to the owner's terminal, as a table or as JSON with --json, and exit with a non-zero status when any balance assertion fails; the core must log only that the check ran, with its counts, never an account, an instrument or a value
    role: normative
    status: modeled
    span:
      start: 0
      end: 287
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:27.839Z'
id: REQ-views-cli-check-balances
type: req
---
openfief check balances must print the balance report only to the owner's terminal, as a table or as JSON with --json, and exit with a non-zero status when any balance assertion fails; the core must log only that the check ran, with its counts, never an account, an instrument or a value.

## Context

D-200 criterion 1 is the owner's own check on a private install with real accounts, never in CI (GR-12). The design session asked, through the orchestrator on 2026-11-06, for an owner-only command, `openfief check balances`, so the author can check it: every balance assertion with reported and derived values, the difference and pass or fail at 0.01 of the account currency, open XTB CFDs flagged as unsupported (D-24), values printed only to the owner's terminal, the core logging only that the check ran with counts, a non-zero exit status on a failure, a table and --json. GR-12 keeps real values out of logs.

## Source

> Every broker-reported position and cash balance is stored as a balance assertion and matches the derived value to 0.01 of the account currency. Open XTB CFDs appear only as flagged, unsupported items.

docs/DESIGN.md#D-200
