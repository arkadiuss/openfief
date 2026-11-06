---
title: The core must let the owner start one sync of a connector instance at once through an owner-only service, after the authorizer allows a change of every account bound to it, and must refuse it with that reason while a sync of the instance runs.
status: open
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
semantic_text: The core must let the owner start one sync of a connector instance at once through an owner-only service, after the authorizer allows a change of every account bound to it, and must refuse it with that reason while a sync of the instance runs.
logic_claims:
  - CLAIM-F697BAF8E0B32439
semantic_clauses:
  - The core must let the owner start one sync of a connector instance at once through an owner-only service, after the authorizer allows a change of every account bound to it, and must refuse it with that reason while a sync of the instance runs.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 89cc95691004f611106cf980d23732777860017af66ff7c83732f7408bc9ad5c
semantic_inventory:
  - claim_key: CLAIM-F697BAF8E0B32439
    claim_text: The core must let the owner start one sync of a connector instance at once through an owner-only service, after the authorizer allows a change of every account bound to it, and must refuse it with that reason while a sync of the instance runs
    role: normative
    status: modeled
    span:
      start: 0
      end: 242
origin:
  kind: agent
  recorded_at: '2026-11-06T04:23:02.580Z'
id: REQ-ingest-owner-retries-sync
type: req
---
The core must let the owner start one sync of a connector instance at once through an owner-only service, after the authorizer allows a change of every account bound to it, and must refuse it with that reason while a sync of the instance runs.

## Context

D-196 lists retrying a sync among the items settled in place in the Inbox; the design session approved an owner-only RPC that goes through the authorizer (D-163) and respects the scheduler's single-flight claim (D-49). Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-196]** ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place ...

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

> **[D-49]** The core calls connectors: it owns schedules, retries, import batches and watermarks.

docs/DESIGN.md#D-49

> **[D-163]** Every read path and owner service of the core asks one authorization interface whether the caller may act on each account.

docs/DESIGN.md#D-163
