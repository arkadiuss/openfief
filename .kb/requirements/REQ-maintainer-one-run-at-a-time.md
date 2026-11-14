---
title: At most one maintainer run may be queued or running at a time for the model portfolio.
status: open
firmness: default
origin:
  kind: agent
  ref: the M3 brief for the model-portfolio maintainer (D-204) on the model router, 2026-11-13
  recorded_at: '2026-11-13T21:07:45.569Z'
semantic_text: At most one maintainer run may be queued or running at a time for the model portfolio.
semantic_source_field: semantic_text
semantic_source_hash: 0c089074a9fd80e6176bd4a35c1bdfe25c3ffbea7450167a6e918c418f945a49
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-953D89E32E2ED3D6
    claim_text: At most one maintainer run may be queued or running at a time for the model portfolio
    role: descriptive
    status: missing
    span:
      start: 0
      end: 85
logic_claims:
  - CLAIM-953D89E32E2ED3D6
semantic_clauses:
  - At most one maintainer run may be queued or running at a time for the model portfolio
required_evidence: unverified
id: REQ-maintainer-one-run-at-a-time
type: req
---
At most one maintainer run may be queued or running at a time for the model portfolio.

## Context

Two runs at once would send the model the same figures twice and could propose against each other; the store enforces it so a second ask or the schedule finds the first.

## Source

> One run at a time per model portfolio.

Source: the M3 brief for the model-portfolio maintainer (D-204) on the model router, 2026-11-13