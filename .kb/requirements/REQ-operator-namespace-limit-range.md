---
title: The install's cluster-admin base must give the openfief namespace a LimitRange that sets a default CPU and memory request and a default memory limit for every container that sets none of its own.
status: open
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
semantic_text: The install's cluster-admin base must give the openfief namespace a LimitRange that sets a default CPU and memory request and a default memory limit for every container that sets none of its own.
logic_claims:
  - CLAIM-27558F20D5AC5F1A
semantic_clauses:
  - The install's cluster-admin base must give the openfief namespace a LimitRange that sets a default CPU and memory request and a default memory limit for every container that sets none of its own.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1c460c53c64650b1383e3cbeaf02664f6dccf93ca3790ac9a733f0521f74901e
semantic_inventory:
  - claim_key: CLAIM-27558F20D5AC5F1A
    claim_text: The install's cluster-admin base must give the openfief namespace a LimitRange that sets a default CPU and memory request and a default memory limit for every container that sets none of its own
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
origin:
  kind: agent
  recorded_at: '2026-11-02T20:55:51.262Z'
id: REQ-operator-namespace-limit-range
type: req
---
The install's cluster-admin base must give the openfief namespace a LimitRange that sets a default CPU and memory request and a default memory limit for every container that sets none of its own.

## Context

Explicit limits on plugin pods cover what the operator creates; the LimitRange is the backstop for anything else in the namespace, such as a pod applied with kubectl or CloudNativePG's database pods, whose Cluster sets no resources. Its memory limit, 1Gi, leaves the database room; it sets no default CPU limit because CPU is only throttled and the requests already share it. It lives in deploy/crds next to the namespace, so the release bundle's openfief-crds file carries it. The author decided the namespace gets a LimitRange (D-182). Modeled by Claude Code for the P0 security fixes before the alpha release, from the security review's finding that plugin pods had no resource limits; the author decided the fix as D-182.

## Source

> D-182: plugin pods get requests and limits, and the namespace gets a LimitRange.

docs/DESIGN.md D-182 (author decision on the security review's finding P0-4), GR-12
