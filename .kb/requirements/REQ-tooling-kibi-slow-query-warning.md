---
title: The docs build must ask kibi for the requirement coverage report at most 300 requirements at a time, after a query that attaches the store, print how long each page took, and warn when a page took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.
status: open
text_ref: docs/dev/kibi.md#setup-notes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-kibi-md-setup-notes
  - lane:strict
semantic_text: The docs build must ask kibi for the requirement coverage report at most 300 requirements at a time, after a query that attaches the store, print how long each page took, and warn when a page took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.
logic_claims:
  - CLAIM-453DE1B77AABDEE8
semantic_clauses:
  - The docs build must ask kibi for the requirement coverage report at most 300 requirements at a time, after a query that attaches the store, print how long each page took, and warn when a page took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d409f62bc571ea71f6b5ba467a9787c4cbaeecb38da55c90cf55827b8b09514f
semantic_inventory:
  - claim_key: CLAIM-453DE1B77AABDEE8
    claim_text: The docs build must ask kibi for the requirement coverage report at most 300 requirements at a time, after a query that attaches the store, print how long each page took, and warn when a page took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI
    role: normative
    status: modeled
    span:
      start: 0
      end: 293
origin:
  kind: agent
  recorded_at: '2026-11-07T19:28:01.855Z'
id: REQ-tooling-kibi-slow-query-warning
type: req
---
The docs build must ask kibi for the requirement coverage report at most 300 requirements at a time, after a query that attaches the store, print how long each page took, and warn when a page took 90 seconds or more, three quarters of kibi's 120-second query limit, as a workflow warning in CI.

## Context

A coverage query costs a fixed ~28 seconds per call (kibi's quadratic contradiction-witness step) plus time for every row. The whole report in one query (1,231 rows) took 123 to 131 seconds as a command, about 95 of them in the query, close to kibi's 120-second limit, so the earlier reasoning that paging only multiplied the fixed cost was wrong. The design session approved paging again with pages of about 300 rows, each page timed. The command's own work adds about 30 seconds to the query, so the warning stays at 90 seconds until kibi fixes its quadratic step upstream. The property key keeps its earlier name so the existing fact is reused. Requested by the author through the design session; modeled by Claude Code, not yet reviewed.

## Source

> Restore paging with ~300-row pages: docs:gen asks kibi for the coverage report a page at a time (--limit 300 --offset N) and joins the pages' rows ... run each page through scripts/kibi-timed.sh so the duration guard (prints the duration, warns from 90 s as a ::warning:: under Actions) applies PER PAGE. Keep the warm-up query.

docs/dev/kibi.md#setup-notes
