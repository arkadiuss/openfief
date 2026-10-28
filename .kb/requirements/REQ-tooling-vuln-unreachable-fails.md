---
title: When a vulnerability database is unreachable, the vulnerability check must fail.
status: open
text_ref: docs/dev/sdlc.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md
  - lane:strict
semantic_text: When a vulnerability database is unreachable, the vulnerability check must fail.
logic_claims:
  - CLAIM-91C501A30E7295C7
semantic_clauses:
  - When a vulnerability database is unreachable, the vulnerability check must fail.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e381926e2518b74ee937775cbc25ad1635bd4c3fae3594cfc8bb33ce466c616b
semantic_inventory:
  - claim_key: CLAIM-91C501A30E7295C7
    claim_text: When a vulnerability database is unreachable, the vulnerability check must fail
    role: condition
    status: modeled
    span:
      start: 0
      end: 79
firmness: default
rationale: The author opened the session's network and asked to drop the workarounds made for the restricted sandbox; a check that passes when it could not run hides risk.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:00.970Z'
id: REQ-tooling-vuln-unreachable-fails
type: req
---
When a vulnerability database is unreachable, the vulnerability check must fail.

## Context

The earlier rule let the check warn and pass locally when a database was unreachable, a workaround for a sandbox whose network blocked the npm audit endpoint and vuln.go.dev. The author opened the network and asked to drop the workarounds (design session, 2026-10-28), so an unreachable database now fails the check everywhere, as a finding does.

## Why this changed

REQ-tooling-vuln-offline let an unreachable database pass locally only because the development sandbox could not reach the npm audit endpoint or vuln.go.dev. The author opened the network and asked to drop such workarounds, and a check that passes without running hides risk.

## Source

> `task vuln` fails on any finding, and also when a vulnerability database can't be reached.

docs/dev/sdlc.md#tooling
