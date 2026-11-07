---
title: While a running plugin instance expects an extra CA bundle, the operator must stamp a checksum of its ca-bundle Secret data on the pod template and read the Secret again at least every minute, so creating, changing or deleting the Secret rolls the pod, reading only that Secret by name through a Role the owner install creates.
status: open
text_ref: design-session:connector-endpoint-ca-refinements
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-refinements
  - lane:strict
semantic_text: While a running plugin instance expects an extra CA bundle, the operator must stamp a checksum of its ca-bundle Secret data on the pod template and read the Secret again at least every minute, so creating, changing or deleting the Secret rolls the pod, reading only that Secret by name through a Role the owner install creates.
logic_claims:
  - CLAIM-83640C80FC38E6B5
semantic_clauses:
  - While a running plugin instance expects an extra CA bundle, the operator must stamp a checksum of its ca-bundle Secret data on the pod template and read the Secret again at least every minute, so creating, changing or deleting the Secret rolls the pod, reading only that Secret by name through a Role the owner install creates.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6a389823ecd69a7b7bd54e5a9b252bee77c23280c4c2ce997369324bf95ff767
semantic_inventory:
  - claim_key: CLAIM-83640C80FC38E6B5
    claim_text: While a running plugin instance expects an extra CA bundle, the operator must stamp a checksum of its ca-bundle Secret data on the pod template and read the Secret again at least every minute, so creating, changing or deleting the Secret rolls the pod, reading only that Secret by name through a Role the owner install creates
    role: normative
    status: modeled
    span:
      start: 0
      end: 326
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:14.655Z'
id: REQ-operator-ca-bundle-rolls-pod
type: req
---
While a running plugin instance expects an extra CA bundle, the operator must stamp a checksum of its ca-bundle Secret data on the pod template and read the Secret again at least every minute, so creating, changing or deleting the Secret rolls the pod, reading only that Secret by name through a Role the owner install creates.

## Context

Design session refinements of the connector endpoint decision (an https base URL plus an optional extra CA bundle Secret <instance>--ca-bundle added to the system roots), given to the implementation agent by the orchestrator on 2026-11-07, with an update that the expectation of a bundle must be explicit, a boolean extraCA instance setting, never inferred from a base URL that differs from its default, since mirrors and other providers may have public certificates. The operator own Role reads no Secret (D-15, GR-12); following the D-122 precedent for the in-cluster service, the owner install with --extra-ca creates a Role and RoleBinding in openfief letting the operator get that one Secret, so it polls instead of watching. The checksum is a hash only, never the data.

## Source

> Refresh rolls the pod: the operator stamps a checksum of the ca-bundle Secret data on the pod template (annotation), so creating/changing/deleting the Secret rolls the pod by itself.

design-session:connector-endpoint-ca-refinements
