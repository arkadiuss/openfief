---
title: The docs build must ask kibi for the requirement coverage report in one query, print how long it took, and warn when it took 60 seconds or more, half of kibi's 120-second query limit, as a workflow warning in CI.
status: open
text_ref: docs/dev/kibi.md#setup-notes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-setup-notes
  - lane:strict
semantic_text: The docs build must ask kibi for the requirement coverage report in one query, print how long it took, and warn when it took 60 seconds or more, half of kibi's 120-second query limit, as a workflow warning in CI.
logic_claims:
  - CLAIM-5546B95D1BAD5EC6
semantic_clauses:
  - The docs build must ask kibi for the requirement coverage report in one query, print how long it took, and warn when it took 60 seconds or more, half of kibi's 120-second query limit, as a workflow warning in CI.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a78383018ef93b84ab639b20097c490e3cb8e6677b2a8d14e27ddec3f0112ea4
semantic_inventory:
  - claim_key: CLAIM-5546B95D1BAD5EC6
    claim_text: The docs build must ask kibi for the requirement coverage report in one query, print how long it took, and warn when it took 60 seconds or more, half of kibi's 120-second query limit, as a workflow warning in CI
    role: normative
    status: modeled
    span:
      start: 0
      end: 211
origin:
  kind: agent
  recorded_at: '2026-11-07T19:28:01.855Z'
id: REQ-tooling-kibi-slow-query-warning
type: req
---
The docs build must ask kibi for the requirement coverage report in one query, print how long it took, and warn when it took 60 seconds or more, half of kibi's 120-second query limit, as a workflow warning in CI.

## Context

Paging the coverage report to stay under kibi's 120-second query limit paid a fixed cost of about 35 seconds a page and hid the real cause, a store with records stored twice. The author asked to go back to one query, keep the 120-second limit, and see a slowing query long before it fails the build. Requested by the author through the design session; modeled by Claude Code, not yet reviewed.

## Source

> Duration guard: docs:gen prints how long the coverage query took; CI emits a ::warning:: once it passes half the 120 s timeout (60 s).

docs/dev/kibi.md#setup-notes
