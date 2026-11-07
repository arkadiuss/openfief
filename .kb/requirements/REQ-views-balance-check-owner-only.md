---
title: 'The balance report must be served to the owner alone: the views must refuse it to every other caller, such as a script token, a view plugin or a demo session, no read API method may serve it, and the CLI must reach it by running the core''s check command in the core''s pod through Kubernetes exec with the owner''s kubeconfig, over the accounts the authorizer lets the owner view.'
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: 'The balance report must be served to the owner alone: the views must refuse it to every other caller, such as a script token, a view plugin or a demo session, no read API method may serve it, and the CLI must reach it by running the core''s check command in the core''s pod through Kubernetes exec with the owner''s kubeconfig, over the accounts the authorizer lets the owner view.'
logic_claims:
  - CLAIM-91551CAD1022A290
semantic_clauses:
  - 'The balance report must be served to the owner alone: the views must refuse it to every other caller, such as a script token, a view plugin or a demo session, no read API method may serve it, and the CLI must reach it by running the core''s check command in the core''s pod through Kubernetes exec with the owner''s kubeconfig, over the accounts the authorizer lets the owner view.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 10c1f594c6bfd9f1c4a7e01ddb073f00c3f74959983f8c300bfd58090a0c94bb
semantic_inventory:
  - claim_key: CLAIM-91551CAD1022A290
    claim_text: 'The balance report must be served to the owner alone: the views must refuse it to every other caller, such as a script token, a view plugin or a demo session, no read API method may serve it, and the CLI must reach it by running the core''s check command in the core''s pod through Kubernetes exec with the owner''s kubeconfig, over the accounts the authorizer lets the owner view'
    role: normative
    status: modeled
    span:
      start: 0
      end: 377
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:22.726Z'
id: REQ-views-balance-check-owner-only
type: req
---
The balance report must be served to the owner alone: the views must refuse it to every other caller, such as a script token, a view plugin or a demo session, no read API method may serve it, and the CLI must reach it by running the core's check command in the core's pod through Kubernetes exec with the owner's kubeconfig, over the accounts the authorizer lets the owner view.

## Context

D-200 criterion 1 is the owner's own check on a private install with real accounts, never in CI (GR-12). The design session asked, through the orchestrator on 2026-11-06, for an owner-only command, `openfief check balances`, so the author can check it: every balance assertion with reported and derived values, the difference and pass or fail at 0.01 of the account currency, open XTB CFDs flagged as unsupported (D-24), values printed only to the owner's terminal, the core logging only that the check ran with counts, a non-zero exit status on a failure, a table and --json. The exec path follows D-80 (openfief enroll); the authorizer is D-163's.

## Source

> Every broker-reported position and cash balance is stored as a balance assertion and matches the derived value to 0.01 of the account currency. Open XTB CFDs appear only as flagged, unsupported items.

docs/DESIGN.md#D-200
