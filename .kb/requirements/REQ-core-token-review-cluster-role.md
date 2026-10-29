---
title: The core's ClusterRole must allow only creating TokenReviews.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-82-2
  - lane:strict
text_ref: design-decisions:D-82#2
origin:
  kind: import
  ref: design-decisions:D-82#2
  recorded_at: '2026-10-29T08:16:42.119Z'
semantic_text: The core's ClusterRole must allow only creating TokenReviews.
logic_claims:
  - CLAIM-865543974E45E494
semantic_clauses:
  - The core's ClusterRole must allow only creating TokenReviews.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8b9f20ef028cf684d9a070c3d7771d70acfe9321b52756df5ee58f5626b42556
semantic_inventory:
  - claim_key: CLAIM-865543974E45E494
    claim_text: The core's ClusterRole must allow only creating TokenReviews
    role: normative
    status: modeled
    span:
      start: 0
      end: 60
id: REQ-core-token-review-cluster-role
type: req
---
The core's ClusterRole must allow only creating TokenReviews.

## Context

D-82 has the core check view plugins' tokens with Kubernetes' TokenReview API (REQ-identity-view-plugin-token-review), which needs a cluster-scoped permission, because TokenReview is not a namespaced resource. The author limits that permission to creating TokenReviews and nothing else. The operator stays namespace-only (D-15, REQ-AUTO-BBBDDCF207E884A1), and the egress proxy checks plugins' egress tokens with TokenReview too (D-63, REQ-plugins-egress-proxy-identifies-plugin); REQ-core-cluster-wide-permissions records that these two are OpenFief's only cluster-wide permissions. The author recorded D-81 and D-82 after the design session that answered the implementation's questions.

## Source

> The core checks view plugins' tokens with Kubernetes' TokenReview API, which also confirms that the plugin's pod still exists. Its ClusterRole allows only creating TokenReviews; with the egress proxy's (D-63), it is one of the two cluster-wide permissions OpenFief holds.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-82#2
