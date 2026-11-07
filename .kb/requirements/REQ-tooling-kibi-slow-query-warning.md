---
title: The docs build must ask kibi for the requirement coverage report in one query after a query that attaches the store, print how long it took, and warn when it took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.
status: open
text_ref: docs/dev/kibi.md#setup-notes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-setup-notes
  - lane:strict
semantic_text: The docs build must ask kibi for the requirement coverage report in one query after a query that attaches the store, print how long it took, and warn when it took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.
logic_claims:
  - CLAIM-31F84E41A5116998
semantic_clauses:
  - The docs build must ask kibi for the requirement coverage report in one query after a query that attaches the store, print how long it took, and warn when it took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 80cbf84f742edfdbed1592cd276a8bab159997a705ed0bce9acd86ddf5d70348
semantic_inventory:
  - claim_key: CLAIM-31F84E41A5116998
    claim_text: The docs build must ask kibi for the requirement coverage report in one query after a query that attaches the store, print how long it took, and warn when it took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI
    role: normative
    status: modeled
    span:
      start: 0
      end: 259
origin:
  kind: agent
  recorded_at: '2026-11-07T19:28:01.855Z'
id: REQ-tooling-kibi-slow-query-warning
type: req
---
The docs build must ask kibi for the requirement coverage report in one query after a query that attaches the store, print how long it took, and warn when it took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.

## Context

Paging the coverage report to stay under kibi's 120-second query limit paid a fixed cost of about 35 seconds a page and hid the real cause, a store with records stored twice. The author asked to go back to one query, keep the 120-second limit, and see a slowing query long before it fails the build. The command's own work adds about 30 seconds to the query, so a warning at half the limit fired on every run; the design session set it at 90 seconds until kibi fixes its quadratic contradiction-witness step upstream, when the guard moves to timing the query alone at half the limit. Attaching and deduplicating a just-published store took 16 seconds of the first query, so a cheap query attaches it first. Requested by the author through the design session; modeled by Claude Code, not yet reviewed.

## Source

> Duration guard: docs:gen prints how long the coverage query took; CI emits a ::warning:: once it passes half the 120 s timeout (60 s).
> (d) 90 s for now. A warning that fires on every run teaches everyone to ignore it. Note in kibi.md that it moves to (b), timing the query alone, once upstream fixes the quadratic witness step.

docs/dev/kibi.md#setup-notes
