---
title: At most one sync of a connector instance must run at a time, within one core process and across core replicas through a stored claim that expires, and each attempt must be bounded by a timeout whose expiry counts as a failed attempt.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: At most one sync of a connector instance must run at a time, within one core process and across core replicas through a stored claim that expires, and each attempt must be bounded by a timeout whose expiry counts as a failed attempt.
logic_claims:
  - CLAIM-950D1300614267D2
semantic_clauses:
  - At most one sync of a connector instance must run at a time, within one core process and across core replicas through a stored claim that expires, and each attempt must be bounded by a timeout whose expiry counts as a failed attempt.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: eb0508c6f1ba76e0ffdad34e85c46fdd80ddfef39f48eb17ff396429901bc716
semantic_inventory:
  - claim_key: CLAIM-950D1300614267D2
    claim_text: At most one sync of a connector instance must run at a time, within one core process and across core replicas through a stored claim that expires, and each attempt must be bounded by a timeout whose expiry counts as a failed attempt
    role: normative
    status: modeled
    span:
      start: 0
      end: 232
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:41.661Z'
id: REQ-ingest-scheduler-single-flight
type: req
---
At most one sync of a connector instance must run at a time, within one core process and across core replicas through a stored claim that expires, and each attempt must be bounded by a timeout whose expiry counts as a failed attempt.

## Context

Two concurrent syncs of one instance would race for its watermark and open batch. The core's Deployment may run several replicas (River supports it), so an in-process lock alone is not enough; a lease that expires lets a crashed attempt's claim lapse. Modeled by Claude Code for the core-connectors task (orchestrator brief: the core scheduler, watermarks, connector bindings); the author has not reviewed this requirement yet.

## Source

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49
