---
title: The end-to-end test task must refuse to run when the current Kubernetes context is not a local cluster.
status: open
text_ref: docs/dev/testing.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-testing-md
  - lane:strict
semantic_text: The end-to-end test task must refuse to run when the current Kubernetes context is not a local cluster.
logic_claims:
  - CLAIM-5C5EEB3099815E2F
semantic_clauses:
  - The end-to-end test task must refuse to run when the current Kubernetes context is not a local cluster.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6503944fcfa10f136679c41e92166ff3547c1538e3d83e9e08984ce22895e79a
semantic_inventory:
  - claim_key: CLAIM-5C5EEB3099815E2F
    claim_text: The end-to-end test task must refuse to run when the current Kubernetes context is not a local cluster
    role: normative
    status: modeled
    span:
      start: 0
      end: 102
origin:
  kind: agent
  recorded_at: '2026-10-27T20:33:06.559Z'
id: REQ-tooling-e2e-local-only
type: req
---
The end-to-end test task must refuse to run when the current Kubernetes context is not a local cluster.

## Context

End-to-end tests run on a local k3s (the author's testing rule, docs/dev/testing.md); refusing other contexts keeps tests from touching a shared or production cluster. The refusal itself is a Claude Code proposal, marked P in docs/dev/testing.md.

## Source

> `task e2e` refuses any kubeconfig context that isn't local

docs/dev/testing.md#end-to-end-tests
